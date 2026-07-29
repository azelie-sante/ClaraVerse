// Renames the collections that outlived the Nexus feature removal.
//
// The Nexus surface was deleted, but four of its collections back features
// that are still live (assistant persona, long-term memory, and the RAG
// knowledge base). Their Go constants dropped the "nexus_" prefix, so the
// physical collections have to be renamed to match or the app will come up
// pointing at empty collections while the old data sits there orphaned.
//
// Safe to re-run: each rename is skipped when the source is missing or the
// target already exists.
//
// Usage:
//   mongosh "$MONGODB_URI" backend/scripts/migrate_legacy_collections.js
//
// The collections belonging to the deleted feature (nexus_tasks,
// nexus_daemons, nexus_sessions, nexus_daemon_templates, nexus_projects,
// nexus_saves, nexus_orchestration_state, nexus_artifacts) are NOT touched.
// Nothing reads them anymore; drop them by hand once you are satisfied the
// removal went cleanly.

const renames = [
  ['nexus_persona', 'persona'],
  ['nexus_engrams', 'engrams'],
  ['nexus_knowledge_files', 'knowledge_files'],
  ['nexus_knowledge_collections', 'knowledge_collections'],
];

const existing = new Set(db.getCollectionNames());

for (const [from, to] of renames) {
  if (!existing.has(from)) {
    print(`skip   ${from} → ${to} (source missing)`);
    continue;
  }
  if (existing.has(to)) {
    print(`SKIP   ${from} → ${to} (target already exists — merge by hand)`);
    continue;
  }
  db.getCollection(from).renameCollection(to);
  print(`renamed ${from} → ${to}`);
}

print('done.');
