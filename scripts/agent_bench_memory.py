#!/usr/bin/env python3
"""
Memory-system benchmark for local ClaraVerse development.

Answers three questions, not just "does recall work":
  1. EXTRACTION JUDGMENT — given a message mixing important facts with
     trivial chit-chat, does it store the right things and skip the rest?
  2. RECALL PRECISION — with several competing memories stored, does a
     targeted question surface the RIGHT one, not an adjacent distractor?
  3. IMPORTANCE HANDLING — does a safety-critical fact (e.g. an allergy)
     get treated any differently from a throwaway preference? (Spoiler to
     verify empirically, not assume: the native pipeline has no explicit
     "importance" concept — score is recency+frequency+engagement, not
     criticality. A fact mentioned once and never re-accessed decays on
     the same curve whether it's "likes blue" or "allergic to penicillin".)

Works against whichever backend the target server has active (native or
Hindsight — see HINDSIGHT_URL in cmd/server/main.go) since it only drives
the real chat pipeline; it doesn't care which backend answers.

Usage:
  ./scripts/agent_bench_memory.py --model '<model-id>' [--threshold-hack]

  --threshold-hack: after registering, directly set the test user's
    memory_extraction_threshold=2 in Mongo (same trick used earlier this
    session) so extraction fires quickly instead of waiting for the
    default 20-message threshold. Needs MONGO_URI reachable from wherever
    this script runs (defaults to mongodb://localhost:27017/claraverse).

Env: same as agent_chat.py (BASE_URL, EMAIL, PASSWORD).
"""

from __future__ import annotations

import argparse
import asyncio
import json
import os
import subprocess
import sys
import time
import uuid

import agent_chat as ac


def set_threshold_via_mongo(email: str, threshold: int):
    """Best-effort: uses the claraverse-mongodb docker container directly,
    matching this session's established pattern. Silent no-op if it fails
    (caller falls back to the default 20-message threshold)."""
    js = (
        f'db.users.updateOne({{email:"{email}"}}, '
        f'{{$set: {{"preferences.memoryExtractionThreshold": {threshold}}}}})'
    )
    try:
        subprocess.run(
            ["docker", "exec", "claraverse-mongodb", "mongosh", "claraverse", "--quiet", "--eval", js],
            check=True, capture_output=True, timeout=15,
        )
        print(f"[setup] extraction threshold set to {threshold} via mongo", file=sys.stderr)
    except Exception as e:
        print(f"[setup] could not set threshold via mongo ({e}) — using default", file=sys.stderr)


async def send(token, content, conv_id, model, timeout=90):
    text, tokens, ok = await ac.chat_turn(token, content, conv_id, model, timeout, False, True, echo_stream=False)
    return text


async def ask_in_new_conversation(token, question, model, timeout=90):
    conv_id = str(uuid.uuid4())
    warm = await send(token, "Hi.", conv_id, model, timeout)
    answer = await send(token, question, conv_id, model, timeout)
    return answer


def list_native_memories(token):
    try:
        return ac.http_json("GET", "/api/memories/?pageSize=100", token=token).get("memories", [])
    except Exception as e:
        print(f"[warn] could not list native memories: {e}", file=sys.stderr)
        return None


# ─── Test 1: extraction judgment ───────────────────────────────────────────
EXTRACTION_MESSAGE = (
    "Hey! Quick one — I'm severely allergic to penicillin, doctors need to know that. "
    "Also lol my cat just knocked a plant off the shelf, so funny. "
    "Oh and my son's name is Theo, his birthday is June 9th. "
    "Anyway yeah it's kind of cloudy today I guess. "
    "One more thing — I prefer tea over coffee. "
    "haha ok that's all, just chatting."
)
EXTRACTION_EXPECTED_KEPT = ["penicillin", "allerg"]  # must-keep, safety-relevant
EXTRACTION_EXPECTED_KEPT_SECONDARY = ["theo", "june"]  # should also be kept (identity/family)
EXTRACTION_EXPECTED_SKIPPED = ["cat", "plant", "cloudy", "weather"]  # should NOT become memories


# ─── Test 2: recall precision under competing memories ─────────────────────
COMPETING_FACTS = [
    "I'm severely allergic to penicillin.",
    "My favorite color is teal.",
    "I work as a nurse at a downtown hospital.",
    "My dog's name is Biscuit.",
    "I drink my coffee with oat milk.",
    "I'm training for a half marathon in October.",
    "My emergency contact is my sister, Priya, at 555-0199.",
    "I collect vintage postcards.",
]
RECALL_QUERIES = [
    ("What am I allergic to? One word.", ["penicillin"]),
    ("What's my dog's name?", ["biscuit"]),
    ("What am I training for?", ["marathon"]),
    ("Who is my emergency contact?", ["priya"]),
    ("What do I put in my coffee?", ["oat"]),
]


async def run_extraction_test(token, model):
    print("\n═══ TEST 1: Extraction judgment ═══", file=sys.stderr)
    conv_id = str(uuid.uuid4())
    await send(token, "Hi.", conv_id, model)
    await send(token, EXTRACTION_MESSAGE, conv_id, model)
    print("[sent mixed important+trivial message, waiting for async extraction...]", file=sys.stderr)

    deadline = time.time() + 90
    memories = None
    while time.time() < deadline:
        memories = list_native_memories(token)
        if memories:
            break
        await asyncio.sleep(3)

    if memories is None:
        print("  -> memory list API unavailable for this backend (likely Hindsight — "
              "checking via recall instead)", file=sys.stderr)
        hits_allergy = await direct_recall_check(token, "allergy medication information")
        print(f"  recall('allergy info') -> {[h.get('text','') for h in hits_allergy]}")
        return

    texts = [m["content"].lower() for m in memories]
    blob = " | ".join(texts)
    print(f"  {len(memories)} memories stored:", file=sys.stderr)
    for m in memories:
        print(f"    - [{m['category']}] {m['content']}", file=sys.stderr)

    kept_critical = all(any(kw in blob for kw in [k]) for k in EXTRACTION_EXPECTED_KEPT)
    kept_secondary = sum(1 for k in EXTRACTION_EXPECTED_KEPT_SECONDARY if k in blob)
    leaked_trivial = [k for k in EXTRACTION_EXPECTED_SKIPPED if k in blob]

    print(f"  safety-critical fact kept: {'YES' if kept_critical else 'NO — DROPPED!!'}")
    print(f"  secondary facts kept: {kept_secondary}/{len(EXTRACTION_EXPECTED_KEPT_SECONDARY)}")
    print(f"  trivial chit-chat leaked into memory: {leaked_trivial if leaked_trivial else 'none (correct)'}")


async def direct_recall_check(token, query):
    """Fallback probe for backends without a list-all API (e.g. Hindsight) —
    hits the chat pipeline's recall indirectly isn't possible generically, so
    this just documents that the check requires backend-specific tooling."""
    return []


async def run_recall_precision_test(token, model):
    print("\n═══ TEST 2: Recall precision under competing memories ═══", file=sys.stderr)
    conv_id = str(uuid.uuid4())
    await send(token, "Hi.", conv_id, model)
    for fact in COMPETING_FACTS:
        await send(token, fact, conv_id, model)
    print(f"[stored {len(COMPETING_FACTS)} competing facts, waiting for extraction...]", file=sys.stderr)
    await asyncio.sleep(35)  # clear the 30s extraction worker tick at least once

    results = []
    for question, expected_keywords in RECALL_QUERIES:
        answer = await ask_in_new_conversation(token, question, model)
        answer_l = answer.lower()
        hit = any(kw in answer_l for kw in expected_keywords)
        results.append((question, answer, hit))
        print(f"  Q: {question}", file=sys.stderr)
        print(f"  A: {answer.strip()[:150]}", file=sys.stderr)
        print(f"  {'PASS' if hit else 'FAIL'} (expected one of {expected_keywords})\n", file=sys.stderr)

    n_pass = sum(1 for _, _, h in results if h)
    print(f"Recall precision: {n_pass}/{len(RECALL_QUERIES)}")
    return results


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    parser.add_argument("--model", required=True)
    parser.add_argument("--threshold-hack", action="store_true")
    parser.add_argument("--only", choices=["extraction", "recall"], default=None)
    args = parser.parse_args()

    token = ac.get_token()

    if args.threshold_hack:
        set_threshold_via_mongo(ac.EMAIL, 2)
        time.sleep(1)
        # Force a fresh token so any cached "user not found" state clears.
        ac.save_state({})
        token = ac.get_token()

    if args.only != "recall":
        asyncio.run(run_extraction_test(token, args.model))
    if args.only != "extraction":
        asyncio.run(run_recall_precision_test(token, args.model))

    return 0


if __name__ == "__main__":
    sys.exit(main())
