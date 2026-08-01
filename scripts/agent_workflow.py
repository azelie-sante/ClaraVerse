#!/usr/bin/env python3
"""
No-browser workflow client for local ClaraVerse development.

Drives the Workflow surface (agent-builder pipelines: input -> blocks ->
output) via REST + the /ws/workflow WebSocket, entirely from the terminal.
Shares auth/state plumbing with agent_chat.py (same cached token, same
~/.cache/claraverse-agent-env/state.json).

The fastest path to a *runnable* workflow is cloning one of the seeded
built-in templates (see backend/internal/services/workflow_template_store.go)
rather than hand-authoring a block graph — the template gallery guarantees
a valid, executable block/connection schema.

Wire protocol (backend/internal/handlers/workflow_websocket.go):
  client -> server: {"type": "execute_workflow", "agent_id", "input": {...}}
  server -> client: "connected"
                     "execution_started"  (execution_id)
                     "execution_update"   (block_id, status, inputs, output, error) x N
                     "execution_complete" (status, final_output, duration_ms,
                                           api_response: {result, data, artifacts, files})
                     "error"

Usage:
  ./scripts/agent_workflow.py --list-templates
  ./scripts/agent_workflow.py --clone <template_id> [--name "My copy"]
  ./scripts/agent_workflow.py --list-agents
  ./scripts/agent_workflow.py --run --agent <agent_id> [--input '{"key":"val"}']
  ./scripts/agent_workflow.py --run                      # uses last-cloned/last-run agent
  ./scripts/agent_workflow.py --run --json                # raw JSONL event dump

Env: same as agent_chat.py (BASE_URL, WS_URL, EMAIL, PASSWORD, TIMEOUT).
State: reuses agent_chat's state file, adding an "agent_id" key.
"""

from __future__ import annotations

import argparse
import asyncio
import json
import sys
import time

import agent_chat as ac

try:
    import websockets
except ImportError:
    print("websockets package required: pip install websockets", file=sys.stderr)
    sys.exit(3)


def list_templates(token: str) -> list[dict]:
    return ac.http_json("GET", "/api/workflow-templates", token=token).get("templates", [])


def clone_template(token: str, template_id: str, name: str = "") -> tuple[str, dict]:
    body = {"name": name} if name else {}
    resp = ac.http_json("POST", f"/api/workflow-templates/{template_id}/clone", body, token=token)
    agent = resp["agent"]
    return agent["id"], resp


def list_agents(token: str) -> list[dict]:
    return ac.http_json("GET", "/api/agents", token=token).get("agents", [])


async def run_workflow(
    token: str, agent_id: str, input_data: dict, timeout: float, raw_json: bool
):
    """Executes a workflow over /ws/workflow. Returns (api_response, ok)."""
    ws_url = f"{ac.WS_URL}/ws/workflow?token={token}"
    payload = {"type": "execute_workflow", "agent_id": agent_id, "input": input_data}

    api_response = None
    ok = False
    deadline = time.time() + timeout

    async with websockets.connect(ws_url, max_size=20 * 1024 * 1024) as ws:
        await ws.send(json.dumps(payload))

        while time.time() < deadline:
            try:
                raw = await asyncio.wait_for(ws.recv(), timeout=deadline - time.time())
            except asyncio.TimeoutError:
                if not raw_json:
                    print(f"[timeout after {timeout}s waiting for execution_complete]", file=sys.stderr)
                break

            if raw_json:
                print(raw, flush=True)

            try:
                msg = json.loads(raw)
            except json.JSONDecodeError:
                continue

            mtype = msg.get("type", "")
            if mtype == "execution_started":
                if not raw_json:
                    print(f"[execution {msg.get('execution_id', '')} started]", file=sys.stderr)
            elif mtype == "execution_update":
                if not raw_json:
                    print(
                        f"[block {msg.get('block_id', '?')}] {msg.get('status', '')}"
                        + (f" error={msg['error']}" if msg.get("error") else ""),
                        file=sys.stderr,
                    )
            elif mtype == "execution_complete":
                api_response = msg.get("api_response")
                ok = msg.get("status") == "completed"
                break
            elif mtype == "error":
                ok = False
                if not raw_json:
                    print(f"[error]: {msg.get('error', msg)}", file=sys.stderr)
                break

    return api_response, ok


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    parser.add_argument("--list-templates", action="store_true")
    parser.add_argument("--clone", metavar="TEMPLATE_ID", help="clone a template into a new agent+workflow")
    parser.add_argument("--name", default="", help="name for --clone")
    parser.add_argument("--list-agents", action="store_true")
    parser.add_argument("--run", action="store_true", help="execute a workflow")
    parser.add_argument("--agent", default="", help="agent_id to run (default: last cloned/run agent)")
    parser.add_argument("--input", default="{}", help="JSON object passed as workflow input")
    parser.add_argument("--json", action="store_true", help="dump raw server events as JSONL")
    parser.add_argument("--timeout", type=float, default=float(ac.TIMEOUT))
    args = parser.parse_args()

    token = ac.get_token()

    if args.list_templates:
        for t in list_templates(token):
            print(f"{t['id']}\t{t.get('name', '')}\t({t.get('category', '')})")
        return 0

    if args.list_agents:
        for a in list_agents(token):
            has_wf = "workflow" if a.get("has_workflow") else "no-workflow"
            print(f"{a['id']}\t{a.get('name', '')}\t[{has_wf}]")
        return 0

    state = ac.load_state()

    if args.clone:
        agent_id, resp = clone_template(token, args.clone, args.name)
        state["agent_id"] = agent_id
        ac.save_state(state)
        print(f"cloned -> agent_id={agent_id}", file=sys.stderr)
        print(agent_id)
        return 0

    if args.run:
        agent_id = args.agent or state.get("agent_id")
        if not agent_id:
            parser.error("no agent_id given and none cached — run --clone <template_id> first, or pass --agent")
        try:
            input_data = json.loads(args.input)
        except json.JSONDecodeError as e:
            parser.error(f"--input must be valid JSON: {e}")

        state["agent_id"] = agent_id
        ac.save_state(state)

        api_response, ok = asyncio.run(run_workflow(token, agent_id, input_data, args.timeout, args.json))
        if not args.json:
            if api_response:
                print(api_response.get("result", ""))
                meta = api_response.get("metadata", {})
                if meta:
                    print(f"[metadata] {meta}", file=sys.stderr)
            else:
                print("[no api_response received]", file=sys.stderr)
        return 0 if ok else 1

    parser.error("pass one of --list-templates / --clone / --list-agents / --run")
    return 2


if __name__ == "__main__":
    sys.exit(main())
