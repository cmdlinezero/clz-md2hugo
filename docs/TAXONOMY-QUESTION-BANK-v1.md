# Taxonomy and Question Bank v1 (increment A)

The curated seed at `taxonomies/google-cloud/taxonomy.json` is **draft**, derived from the supplied PDE quiz, and **not** a complete import of Google Cloud in 4 Words. Review the taxonomy and reconcile it with the upstream source before publication. Its attribution includes the upstream CC BY 4.0 source and adaptation notice. Do not assume a source commit was verified.

Commands:

```sh
certin-content taxonomy validate --file taxonomies/google-cloud/taxonomy.json
certin-content quiz validate --input /path/to/pde-quizzes --taxonomy taxonomies/google-cloud/taxonomy.json --bank professional-data-engineer
certin-content quiz ingest --input /path/to/pde-quizzes --taxonomy taxonomies/google-cloud/taxonomy.json --bank professional-data-engineer --output dist/question-bank.json
```

Input: legacy quiz JSON (`id`, `provider`, `subject`, `tasks`), with task `tag`, `title`, `domain`, `products`, `options`, `answer`, `explanation`. Question bodies remain intact in the output. A bank must contain one provider/subject pair. Stable question IDs use `<quiz-id>:<JSON-tag>`. Files are processed in lexical path order, questions sorted by canonical ID, and output contains no volatile timestamps.

Validation rejects unknown domains/products, duplicate question IDs, missing question fields, or answers that do not exactly match an option. Taxonomy aliases are case-insensitive and trim surrounding whitespace. The `General Google Cloud / Architecture` classification is preserved as a topic exception, not a technology.

This increment **does not** modify the Hugo Practice screen, create sessions, project a knowledge graph, publish artifacts to GCS, or import the complete upstream cheat sheet. The bank output is not access-tier filtered: do not ship protected question payloads to public Hugo assets. Integrate tier-safe delivery before production deployment.

Testing: `GO111MODULE=off GOTOOLCHAIN=local go test ./internal/questionbank` runs the isolated ingestion tests with Go 1.23. The complete compiler requires the Go toolchain declared by `go.mod` (1.26.3) and its dependencies.
