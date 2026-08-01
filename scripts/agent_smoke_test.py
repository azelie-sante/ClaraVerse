#!/usr/bin/env python3
"""
Fast pass/fail check that the chat pipeline actually works, with no
browser involved. Meant for an agent to run after making backend/
frontend changes locally, as a "did I break chat" gate.

Checks, in order:
  1. Backend is reachable (GET /health)
  2. Auth works (login/register via agent_chat.get_token)
  3. At least one model is configured (GET /api/models)
  4. A real chat turn round-trips: send a message, get a non-empty
     stream_end with token usage, on a fresh conversation

Exit 0 = all checks passed. Exit 1 = something's broken; the failing
check is printed to stderr.

Usage:
  ./scripts/agent_smoke_test.py
  BASE_URL=http://localhost:3002 ./scripts/agent_smoke_test.py   # native dev backend
"""

from __future__ import annotations

import asyncio
import sys
import time
import urllib.error
import urllib.request
import uuid

import agent_chat as ac


def check(label: str, fn):
    print(f"── {label} ──", file=sys.stderr)
    t0 = time.time()
    try:
        result = fn()
        print(f"   ok ({time.time() - t0:.1f}s)", file=sys.stderr)
        return result
    except Exception as e:
        print(f"   FAILED: {e}", file=sys.stderr)
        raise


def main() -> int:
    try:
        check("backend reachable", lambda: urllib.request.urlopen(f"{ac.BASE_URL}/health", timeout=10).read())

        token = check("auth (login/register)", ac.get_token)

        models = check("models configured", lambda: ac.list_models(token))
        if not models:
            print("   FAILED: no models configured — add a provider first", file=sys.stderr)
            return 1
        print(f"   {len(models)} model(s), using {models[0]['id']}", file=sys.stderr)

        def do_chat():
            # disable_tools=True: this gate checks "does chat work", not
            # "how fast is tool-calling" — tool-schema injection adds
            # unpredictable, sometimes multi-minute latency on local models
            # (see agent_chat.py's --no-tools flag), which would make this
            # gate flaky for an unrelated reason.
            return asyncio.run(
                ac.chat_turn(
                    token,
                    "Reply with exactly: SMOKE TEST OK",
                    str(uuid.uuid4()),
                    models[0]["id"],
                    timeout=90,
                    raw_json=False,
                    quiet=True,
                    disable_tools=True,
                    echo_stream=False,
                )
            )

        text, tokens, ok = check("chat round-trip (tools disabled)", do_chat)
        if not ok or not text.strip():
            print(f"   FAILED: empty or errored reply (ok={ok}, text={text!r})", file=sys.stderr)
            return 1
        print(f"   reply: {text.strip()[:200]!r}", file=sys.stderr)
        ac.print_token_summary(tokens)

    except (urllib.error.URLError, urllib.error.HTTPError) as e:
        print(f"\n❌ Smoke test FAILED: {e}", file=sys.stderr)
        return 1
    except Exception as e:
        print(f"\n❌ Smoke test FAILED: {e}", file=sys.stderr)
        return 1

    print("\n✅ Smoke test PASSED — chat pipeline is healthy", file=sys.stderr)
    return 0


if __name__ == "__main__":
    sys.exit(main())
