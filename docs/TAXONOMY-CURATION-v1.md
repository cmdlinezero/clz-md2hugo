# Google Cloud taxonomy curation queue (v1)

Run after `taxonomy import`:

```sh
clz-md2hugo taxonomy curate \
  --input dist/google-cloud-reconciliation.json \
  --taxonomy taxonomies/google-cloud/taxonomy.json \
  --output dist/google-cloud-curation.json
```

This produces **review suggestions only**. It never changes the authoritative taxonomy, exam domains, question banks, Practice UI, or portal assets. Decisions: `exclude` for reference-only categories, `review_scope` for broader ecosystems, `review_parent` for known feature candidates, `review_alias` for candidate aliases, and `review_add` for otherwise unmatched entries. `lifecycle_review` flags historically renamed/retired candidates; it does **not** assert a current Google Cloud product status. Verify against official documentation.

The source categories are not Certin quiz domains. Stable IDs and existing taxonomy entries remain untouched. No automatic approval or import is supported in v1. The report carries the upstream source URL, commit SHA, and CC BY 4.0 license for provenance; keep attribution and record any modifications when publishing derived material.
