// One-time collection rename. Run once against your database, then delete
// this file.
//
//   mongosh "$MONGODB_URI" backend/scripts/migrate_legacy_collections.js
//
// Safe to re-run: a rename is skipped when the source is missing or the
// target already exists.

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
