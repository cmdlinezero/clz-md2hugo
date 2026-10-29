# clz-md2hugo v0.4.1

## Deterministic source composition

This patch release extends the v0.4 generator with explicit joins between a primary collection and secondary JSON data sources.

### What changed

- Generator jobs can declare one or more `merge` sources.
- Joins use explicit `localKey` and `foreignKey` fields.
- Joined records are exposed beneath an `as` alias rather than overwriting primary fields.
- Array, selected-array, and object-backed secondary sources are supported.
- Missing matches, duplicate foreign keys, invalid join keys, and alias collisions fail generation.
- Merge traversal and output remain deterministic and work with `--check`.

### Why

Sites often keep editorial content separate from factual or externally refreshed metadata. Deterministic composition allows those sources to remain independently owned while producing one Hugo page model at build time.

The feature is generic: `clz-md2hugo` does not know about repositories, portfolio projects, products, courses, or any other site-specific concept.

### Compatibility

Existing v0.4.0 generator configurations remain valid. `merge` is optional and does not change jobs that omit it.

### Release

After the full repository test suite passes:

```bash
git tag -a v0.4.1 -m "Add deterministic generator source composition"
git push origin v0.4.1
```
