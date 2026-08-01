#!/usr/bin/env python3
"""
No-browser Crew client for local ClaraVerse development.

Crew is a projects -> members -> cards pipeline with mandatory human
review (backend/internal/services/crew_service.go, crew_worker.go).
Unlike chat/workflow there is no WebSocket here: you create a card,
promote it to "queued", and a server-side background worker (2
goroutines, polling) picks it up, runs it, and parks it at "review"
for a human (or this script) to approve. This client polls
GetProject until the card leaves "working".

Card lifecycle: draft -> queued -> working -> review -> done
(models.Card* constants in backend/internal/models/crew.go)

Shares auth/state plumbing with agent_chat.py (same cached token,
same ~/.cache/claraverse-agent-env/state.json, adding a "project_id").

Usage:
  ./scripts/agent_crew.py --roles
  ./scripts/agent_crew.py --new-project "Test Project" --brief "A test project"
  ./scripts/agent_crew.py --hire researcher --name "Rae" [--project <id>]
  ./scripts/agent_crew.py --card "Summarize the input" --assignee <member_id> [--project <id>]
  ./scripts/agent_crew.py --run <card_id> [--approve | --reject "feedback"]
  ./scripts/agent_crew.py --board [--project <id>]     # dump project+members+cards
  ./scripts/agent_crew.py --e2e                         # project->hire->card->run->approve, one shot

Env: same as agent_chat.py (BASE_URL, EMAIL, PASSWORD, TIMEOUT).
"""

from __future__ import annotations

import argparse
import json
import sys
import time

import agent_chat as ac


def roles(token: str) -> list[dict]:
    return ac.http_json("GET", "/api/crew/roles", token=token).get("roles", [])


def create_project(token: str, name: str, brief: str) -> dict:
    return ac.http_json("POST", "/api/crew/projects", {"Name": name, "Brief": brief}, token=token)


def hire_member(token: str, project_id: str, role_key: str, name: str, tools: list[str], model: str) -> dict:
    return ac.http_json(
        "POST",
        f"/api/crew/projects/{project_id}/members",
        {"role_key": role_key, "name": name, "tools": tools, "model": model},
        token=token,
    )


def create_card(token: str, project_id: str, title: str, detail: str, assignee_ids: list[str]) -> dict:
    return ac.http_json(
        "POST",
        f"/api/crew/projects/{project_id}/cards",
        {"title": title, "detail": detail, "assignee_ids": assignee_ids},
        token=token,
    )


def promote_card(token: str, card_id: str) -> dict:
    return ac.http_json("POST", f"/api/crew/cards/{card_id}/queue", {}, token=token)


def review_card(token: str, card_id: str, approve: bool, feedback: str) -> dict:
    return ac.http_json(
        "POST", f"/api/crew/cards/{card_id}/review", {"approve": approve, "feedback": feedback}, token=token
    )


def get_project(token: str, project_id: str) -> dict:
    return ac.http_json("GET", f"/api/crew/projects/{project_id}", token=token)


def find_card(board: dict, card_id: str) -> dict | None:
    for c in board.get("cards") or []:
        if c.get("id") == card_id:
            return c
    return None


def wait_for_card(token: str, project_id: str, card_id: str, timeout: float) -> dict:
    """Polls GetProject until the card leaves 'queued'/'working'. Returns the card."""
    deadline = time.time() + timeout
    last_status = None
    while time.time() < deadline:
        board = get_project(token, project_id)
        card = find_card(board, card_id)
        if card is None:
            raise RuntimeError(f"card {card_id} disappeared from project {project_id}")
        status = card.get("status")
        if status != last_status:
            print(f"[card {card_id}] {status}", file=sys.stderr)
            last_status = status
        if status in ("review", "done"):
            return card
        if card.get("error"):
            print(f"[card {card_id}] error: {card['error']}", file=sys.stderr)
        time.sleep(2)
    raise TimeoutError(f"card {card_id} still '{last_status}' after {timeout}s")


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    parser.add_argument("--roles", action="store_true")
    parser.add_argument("--new-project", metavar="NAME")
    parser.add_argument("--brief", default="")
    parser.add_argument("--project", default="", help="project_id (default: last created)")
    parser.add_argument("--hire", metavar="ROLE_KEY")
    parser.add_argument("--name", default="")
    parser.add_argument("--tools", default="", help="comma-separated tool names for --hire")
    parser.add_argument("--model", default=ac.MODEL_ID)
    parser.add_argument("--card", metavar="TITLE")
    parser.add_argument("--detail", default="")
    parser.add_argument("--assignee", default="", help="member_id for --card (default: last hired)")
    parser.add_argument("--run", metavar="CARD_ID", help="promote a draft card to queued and wait for review")
    parser.add_argument("--approve", action="store_true")
    parser.add_argument("--reject", metavar="FEEDBACK")
    parser.add_argument("--board", action="store_true")
    parser.add_argument("--e2e", action="store_true", help="project -> hire -> card -> run -> approve, one shot")
    parser.add_argument("--timeout", type=float, default=float(ac.TIMEOUT))
    args = parser.parse_args()

    token = ac.get_token()
    state = ac.load_state()

    if args.roles:
        for r in roles(token):
            print(f"{r['key']}\t{r.get('label', '')}\t{r.get('blurb', '')}")
        return 0

    if args.e2e:
        print("── create project ──", file=sys.stderr)
        p = create_project(token, "agent-env e2e", "Automated end-to-end smoke test for the Crew pipeline.")
        project_id = p["id"]
        print(f"   project_id={project_id}", file=sys.stderr)

        print("── hire member (no tools, for speed) ──", file=sys.stderr)
        m = hire_member(token, project_id, "researcher", "Agent-Env Tester", [], args.model)
        member_id = m["id"]
        print(f"   member_id={member_id}", file=sys.stderr)

        print("── create + queue card ──", file=sys.stderr)
        c = create_card(token, project_id, "Say hello", "Reply with exactly: CREW E2E OK", [member_id])
        card_id = c["id"]
        promote_card(token, card_id)
        print(f"   card_id={card_id}", file=sys.stderr)

        print("── wait for worker to run it ──", file=sys.stderr)
        card = wait_for_card(token, project_id, card_id, args.timeout)
        if card["status"] != "review":
            print(f"FAILED: card ended in status={card['status']!r}, expected 'review'", file=sys.stderr)
            return 1
        output = (card.get("latest_output") or "")[:300]
        print(f"   output: {output!r}", file=sys.stderr)

        print("── approve ──", file=sys.stderr)
        review_card(token, card_id, True, "")
        board = get_project(token, project_id)
        final = find_card(board, card_id)
        ok = final["status"] == "done"
        print("done" if ok else f"unexpected final status: {final['status']}")
        return 0 if ok else 1

    if args.new_project:
        p = create_project(token, args.new_project, args.brief)
        state["project_id"] = p["id"]
        ac.save_state(state)
        print(p["id"])
        return 0

    project_id = args.project or state.get("project_id")

    if args.hire:
        if not project_id:
            parser.error("no project_id — pass --project or run --new-project first")
        tools = [t for t in args.tools.split(",") if t]
        m = hire_member(token, project_id, args.hire, args.name or args.hire, tools, args.model)
        state["member_id"] = m["id"]
        ac.save_state(state)
        print(m["id"])
        return 0

    if args.card:
        if not project_id:
            parser.error("no project_id — pass --project or run --new-project first")
        assignee = args.assignee or state.get("member_id")
        if not assignee:
            parser.error("no assignee — pass --assignee or run --hire first")
        c = create_card(token, project_id, args.card, args.detail, [assignee])
        state["card_id"] = c["id"]
        ac.save_state(state)
        print(c["id"])
        return 0

    if args.run:
        if not project_id:
            parser.error("no project_id — pass --project")
        promote_card(token, args.run)
        card = wait_for_card(token, project_id, args.run, args.timeout)
        print(json.dumps(card, indent=2))
        if args.approve:
            review_card(token, args.run, True, "")
        elif args.reject:
            review_card(token, args.run, False, args.reject)
        return 0 if card["status"] == "review" else 1

    if args.board:
        if not project_id:
            parser.error("no project_id — pass --project or run --new-project first")
        print(json.dumps(get_project(token, project_id), indent=2))
        return 0

    parser.error("pass one of --roles / --new-project / --hire / --card / --run / --board / --e2e")
    return 2


if __name__ == "__main__":
    sys.exit(main())
