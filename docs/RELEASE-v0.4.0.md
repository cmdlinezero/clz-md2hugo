# v0.4.0 — Deterministic Hugo content generation

## What changed

- Added `clz-md2hugo generate` for rendering site-owned structured JSON through site-owned Go templates.
- Added deterministic output paths and rendering.
- Added generated-file ownership markers and safe stale-file cleanup.
- Added atomic file replacement.
- Added `--check` for CI validation without writes.
- Added tests covering generation, idempotent checks, stale cleanup, and preservation of human-owned files.

## Why

Multiple Hugo sites need the same build-time capability: keep structured content as the single source of truth and derive Hugo content deterministically. Moving that lifecycle into `clz-md2hugo` avoids site-specific generator scripts while keeping each site's schema and presentation outside the tool.

## Compatibility

The release is additive. Existing Markdown compiler commands and the Certin content contract are unchanged. Generator configuration version 1 uses JSON input and JSON configuration.

## Dev-team verification

```bash
go test ./...
clz-md2hugo generate --config path/to/md2hugo.generate.json
clz-md2hugo generate --config path/to/md2hugo.generate.json --check
```

## Release tag

```bash
git tag -a v0.4.0 -m "Add deterministic Hugo content generation"
git push origin v0.4.0
```
