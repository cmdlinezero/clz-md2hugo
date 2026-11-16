# Google Cloud taxonomy promotion dry run

`taxonomy promote` is **read-only** with respect to the authoritative taxonomy. It writes a review plan, not a replacement taxonomy. It requires an explicit JSON approvals file. By default, all arrays are empty and the plan has no changes.

```sh
./clz-md2hugo taxonomy curate --input dist/google-cloud-reconciliation.json --output dist/google-cloud-curation.json
./clz-md2hugo taxonomy promote --input dist/google-cloud-curation.json --taxonomy taxonomies/google-cloud/taxonomy.json --approvals taxonomies/google-cloud/promotion-approvals.json --output dist/google-cloud-promotion-plan.json
```

Approvals use `suggested_id` from the curation report:

```json
{"schema_version":1,"add":["cloud-run"],"aliases":["cloud-iam","cloud-memorystore"],"parents":["bigquery-ml","bigquery-bi-engine","bigquery-gis","bigquery-data-transfer-service","cloud-sql-insights"]}
```

These are examples only, not pre-approved selections. `add` requires `review_add`, `aliases` requires `review_alias` with an existing target, and `parents` requires `review_parent` with an existing parent. Invalid selections and collisions fail closed. Parent edges are suggestions (`feature_of`), not content hierarchy or prerequisites. Source categories are not exam domains. Lifecycle warnings remain unverified. No quiz assets or protected question content are emitted.

A mixed category entry like Cloud Firestore (`Database`, `Mobile (Firebase)`) is a core Google Cloud candidate; Firebase-only entries remain scope reviews. The report preserves CC BY 4.0 source provenance and upstream commit; any later publishing step must retain attribution and modification notices.

## Controlled apply (explicit opt-in)

The default `taxonomy promote` command remains read-only. To write the approved
changes, supply **both** `--apply` and a unique `--backup` path:

```sh
./clz-md2hugo taxonomy promote \
  --input dist/google-cloud-curation.json \
  --taxonomy taxonomies/google-cloud/taxonomy.json \
  --approvals google-cloud-promotion-approvals-smoke-test.json \
  --output dist/google-cloud-promotion-applied.json \
  --apply --backup taxonomies/google-cloud/taxonomy.before-smoke-test.json
./clz-md2hugo taxonomy validate --file taxonomies/google-cloud/taxonomy.json
```

The apply path rejects empty approvals, unresolved lifecycle warnings, invalid
IDs, and duplicate backups. It uses an exclusive lock, writes a backup before
an atomic rename, and validates the proposed JSON before replacement. Existing
custom taxonomy metadata is preserved. Parent-approved features become
technology nodes with `technology_relationships` entries of type `feature_of`;
these edges are **not** prerequisites or content hierarchy.

The taxonomy version is intentionally **not** bumped automatically; versioning
and final review are separate curator decisions. Do not rerun the same approvals
against the already-updated taxonomy. To roll back, restore the backup manually
and rerun `taxonomy validate`. Commit the approved changes only after reviewing
the diff and source attribution. Keep backup files out of public Hugo assets.
