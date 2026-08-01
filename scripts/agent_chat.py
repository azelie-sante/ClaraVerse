#!/usr/bin/env python3
"""
No-browser chat client for local ClaraVerse development.

Drives ClaraVerse's real chat pipeline — local-JWT auth, the /ws/chat
WebSocket, streaming, tool calls, model selection — entirely from the
terminal, so an agent (or a human) can send messages, inspect tool
calls, and read token/latency stats without ever opening a browser.

Wire protocol (as of the Crew-era backend, see
backend/internal/models/websocket.go for the source of truth):
  client -> server: {"type": "chat_message", "conversation_id", "content",
                      "model_id"?, "system_instructions"?, "disable_tools"?}
  server -> client: "stream_chunk"    (Content)         — assistant text, streamed
                     "reasoning_chunk"(Content)          — thinking tokens, if any
                     "tool_call"      (ToolName, Arguments, Status)
                     "tool_result"    (ToolName, Result)
                     "conversation_title" (Title)        — auto title after turn 1
                     "stream_end"     (Tokens: {input,output,cached,
                                                 duration_ms,cost_usd,model})
                     "error"          (ErrorCode/code, ErrorMessage/message)

The server keys conversation history off conversation_id server-side
(chat_service looks it up by ID), so the client does NOT need to
resend history — just keep reusing the same conversation_id for a
multi-turn conversation.

Usage:
  ./scripts/agent_chat.py "hello, what model are you?"
  ./scripts/agent_chat.py --new "start a fresh conversation"
  ./scripts/agent_chat.py --model '<model-id>' "message"
  ./scripts/agent_chat.py --list-models
  ./scripts/agent_chat.py --repl                    # multi-turn, one process
  ./scripts/agent_chat.py --json "message"           # raw JSONL event dump (for benchmarks)
  echo "message" | ./scripts/agent_chat.py -          # read prompt from stdin

Env:
  BASE_URL   default http://localhost:3000
  WS_URL     derived from BASE_URL if unset
  EMAIL      default agent-test@claraverse.local  (auto-registered on first use)
  PASSWORD   default AgentTest1!
  MODEL_ID   default model_id for chat_message (else server default)
  TIMEOUT    seconds to wait for a full reply (default 180)

State (conversation id, cached token) lives in
~/.cache/claraverse-agent-env/state.json — pass --new to start a
fresh conversation, or delete the file to fully reset identity.
"""

from __future__ import annotations

import argparse
import asyncio
import json
import os
import sys
import time
import urllib.error
import urllib.request
from pathlib import Path

try:
    import websockets
except ImportError:
    print("websockets package required: pip install websockets", file=sys.stderr)
    sys.exit(3)


BASE_URL = os.environ.get("BASE_URL", "http://localhost:3000").rstrip("/")
WS_URL = os.environ.get(
    "WS_URL", BASE_URL.replace("http://", "ws://").replace("https://", "wss://")
)
EMAIL = os.environ.get("EMAIL", "agent-test@claraverse.local")
PASSWORD = os.environ.get("PASSWORD", "AgentTest1!")
MODEL_ID = os.environ.get("MODEL_ID", "")
TIMEOUT = int(os.environ.get("TIMEOUT", "180"))

STATE_DIR = Path(os.path.expanduser("~/.cache/claraverse-agent-env"))
STATE_FILE = STATE_DIR / "state.json"


def load_state() -> dict:
    if STATE_FILE.exists():
        try:
            return json.loads(STATE_FILE.read_text())
        except (json.JSONDecodeError, OSError):
            return {}
    return {}


def save_state(state: dict) -> None:
    STATE_DIR.mkdir(parents=True, exist_ok=True)
    STATE_FILE.write_text(json.dumps(state))


def http_json(method: str, path: str, body: dict | None = None, token: str | None = None) -> dict:
    data = json.dumps(body).encode() if body is not None else None
    req = urllib.request.Request(
        f"{BASE_URL}{path}",
        method=method,
        data=data,
        headers={"Content-Type": "application/json"} if data else {},
    )
    if token:
        req.add_header("Authorization", f"Bearer {token}")
    with urllib.request.urlopen(req, timeout=30) as r:
        return json.loads(r.read())


def login_or_register() -> str:
    """Returns an access token for EMAIL, registering it if it doesn't exist yet."""
    try:
        resp = http_json("POST", "/api/auth/login", {"email": EMAIL, "password": PASSWORD})
        return resp["access_token"]
    except urllib.error.HTTPError as e:
        if e.code not in (401, 404):
            raise
    resp = http_json(
        "POST",
        "/api/auth/register",
        {"name": "Agent Test", "email": EMAIL, "password": PASSWORD},
    )
    return resp["access_token"]


def get_token() -> str:
    state = load_state()
    token = state.get("token")
    if token:
        # Cheap validity check before trusting a cached token. Deliberately
        # NOT /api/models — it degrades to an "anonymous" tier instead of
        # rejecting a bad token, so it never actually catches expiry.
        # /api/agents enforces auth at the handler level (401 on failure),
        # so it's a reliable probe. Access tokens are short-lived (~15min),
        # so this cache mostly avoids re-registering, not re-logging-in.
        try:
            http_json("GET", "/api/agents", token=token)
            return token
        except urllib.error.HTTPError:
            pass
    token = login_or_register()
    state["token"] = token
    save_state(state)
    return token


def list_models(token: str) -> list[dict]:
    return http_json("GET", "/api/models", token=token).get("models", [])


async def chat_turn(
    token: str,
    content: str,
    conversation_id: str,
    model_id: str,
    timeout: float,
    raw_json: bool,
    quiet: bool,
    disable_tools: bool = False,
    echo_stream: bool = True,
):
    """Sends one chat_message, streams the reply. Returns (text, tokens_dict, ok)."""
    ws_url = f"{WS_URL}/ws/chat?token={token}"
    payload = {
        "type": "chat_message",
        "conversation_id": conversation_id,
        "content": content,
    }
    if model_id:
        payload["model_id"] = model_id
    if disable_tools:
        payload["disable_tools"] = True

    text_parts: list[str] = []
    tokens = None
    ok = False
    deadline = time.time() + timeout

    async with websockets.connect(ws_url, max_size=20 * 1024 * 1024) as ws:
        await ws.send(json.dumps(payload))

        while time.time() < deadline:
            try:
                raw = await asyncio.wait_for(ws.recv(), timeout=deadline - time.time())
            except asyncio.TimeoutError:
                if not quiet:
                    print(f"[timeout after {timeout}s waiting for stream_end]", file=sys.stderr)
                break

            if raw_json:
                print(raw, flush=True)

            try:
                msg = json.loads(raw)
            except json.JSONDecodeError:
                continue

            mtype = msg.get("type", "")
            if mtype in ("stream_chunk", "stream_resume"):
                chunk = msg.get("content", "")
                text_parts.append(chunk)
                if not raw_json and echo_stream:
                    print(chunk, end="", flush=True)
            elif mtype == "reasoning_chunk":
                if not quiet and not raw_json:
                    print(msg.get("content", ""), end="", flush=True, file=sys.stderr)
            elif mtype == "tool_call":
                if not quiet and not raw_json:
                    name = msg.get("tool_name", "?")
                    status = msg.get("status", "")
                    print(f"\n[tool_call {name} {status}]", file=sys.stderr)
            elif mtype == "tool_result":
                if not quiet and not raw_json:
                    name = msg.get("tool_name", "?")
                    result = (msg.get("result") or "")[:200]
                    print(f"[tool_result {name}]: {result}", file=sys.stderr)
            elif mtype == "conversation_title":
                if not quiet and not raw_json:
                    print(f"\n[title: {msg.get('title', '')}]", file=sys.stderr)
            elif mtype == "stream_end":
                tokens = msg.get("tokens")
                ok = True
                break
            elif mtype == "error":
                ok = False
                if not raw_json:
                    print(
                        f"\n[error {msg.get('code', '')}]: {msg.get('message', msg)}",
                        file=sys.stderr,
                    )
                break

    if not raw_json and echo_stream:
        print()  # close the streamed line
    return "".join(text_parts), tokens, ok


def print_token_summary(tokens: dict | None):
    if not tokens:
        return
    dur = tokens.get("duration_ms", 0)
    out = tokens.get("output", 0)
    tps = (out / (dur / 1000)) if dur else 0
    print(
        f"[tokens] in={tokens.get('input', 0)} out={out} cached={tokens.get('cached', 0)} "
        f"duration={dur}ms ({tps:.1f} tok/s) model={tokens.get('model', '')} "
        f"cost=${tokens.get('cost_usd', 0):.6f}",
        file=sys.stderr,
    )


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    parser.add_argument("prompt", nargs="?", help="message to send ('-' reads stdin)")
    parser.add_argument("--new", action="store_true", help="start a fresh conversation")
    parser.add_argument("--model", default=MODEL_ID, help="model_id to use for this turn")
    parser.add_argument("--repl", action="store_true", help="interactive multi-turn loop")
    parser.add_argument("--json", action="store_true", help="dump raw server events as JSONL")
    parser.add_argument("--quiet", action="store_true", help="suppress tool call/result chatter")
    parser.add_argument(
        "--no-tools", action="store_true",
        help="disable tools for this turn — much faster on slow local models, "
             "since the backend skips tool-schema injection and predictor calls",
    )
    parser.add_argument("--list-models", action="store_true", help="list available model_ids and exit")
    parser.add_argument("--timeout", type=float, default=TIMEOUT)
    args = parser.parse_args()

    token = get_token()

    if args.list_models:
        for m in list_models(token):
            print(f"{m['id']}\t{m.get('display_name', m.get('name', ''))}\t(provider: {m.get('provider_name', '')})")
        return 0

    state = load_state()
    if args.new or "conversation_id" not in state:
        import uuid

        state["conversation_id"] = str(uuid.uuid4())
        save_state(state)
    conversation_id = state["conversation_id"]

    if args.repl:
        print(f"[conversation {conversation_id}] Ctrl-D to exit", file=sys.stderr)
        while True:
            try:
                line = input("> ")
            except EOFError:
                print()
                return 0
            if not line.strip():
                continue
            text, tokens, ok = asyncio.run(
                chat_turn(token, line, conversation_id, args.model, args.timeout, args.json, args.quiet, args.no_tools)
            )
            print_token_summary(tokens)
            if not ok:
                print("[turn failed]", file=sys.stderr)

    prompt = args.prompt
    if prompt == "-" or prompt is None:
        prompt = sys.stdin.read().strip()
    if not prompt:
        parser.error("a prompt is required (positional arg, '-' for stdin, or --list-models/--repl)")

    text, tokens, ok = asyncio.run(
        chat_turn(token, prompt, conversation_id, args.model, args.timeout, args.json, args.quiet, args.no_tools)
    )
    print_token_summary(tokens)
    return 0 if ok else 1


if __name__ == "__main__":
    sys.exit(main())
