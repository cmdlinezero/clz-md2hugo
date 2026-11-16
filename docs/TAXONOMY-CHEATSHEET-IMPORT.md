# Google Cloud cheat-sheet reconciliation (v1)

This command consumes a **locally downloaded** `README.md` from
https://github.com/priyankavergadia/google-cloud-4-words.

```sh
clz-md2hugo taxonomy import --source /path/to/google-cloud-4-words/README.md \
  --taxonomy taxonomies/google-cloud/taxonomy.json \
  --source-commit UPSTREAM_COMMIT_SHA \
  --output dist/google-cloud-reconciliation.json
```

Review the JSON report before editing the curated taxonomy. `matched` includes
existing names and aliases; `review` identifies probable punctuation variants;
`proposed` lists new candidates. The importer **does not** change taxonomy.json,
quiz data, domain mappings, or the Practice screen. Source categories are evidence
for curation, **not** Certin exam domains or official exam weights. It preserves
multiple source categories for the same product. It does not establish whether a
product is currently offered or deprecated: check official product documentation.

Source: Google Cloud in 4 Words, Priyanka Vergadia, CC BY 4.0
https://creativecommons.org/licenses/by/4.0/ . Record upstream commit SHA and
changes when incorporating source descriptions or categories. No third-party
logos or images are imported.
