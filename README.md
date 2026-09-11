# MD2HUGO 

A small Go compiler for public Hugo Markdown content repository. 
It uses Cobra for the CLI and Go standard-library packages for parsing, 
validation, graph construction, and JSON generation.

## Architecture

The md2hugo compiler accepts an arbitrary directory tree of Markdown content 
and recursively compiles every supported document into one predictable 
artifact: `content.json`.

```text
Markdown files
     |
     v
 recursive scan
     |
     v
 parse -> validate -> resolve relationships
     |
     v
 Certin Content Graph
     |
     v
 content.json
     |
     v
 Hugo Content Adapter
```

One JSON file does not mean one content type. 
The artifact contains separate collections for collections, learning paths, 
certifications, tutorials, quizzes, questions, flashcards and topics.

## Commands

```bash
go run ./md2hugo validate --input examples
go run ./md2hugo build --input examples --output dist/content.json
# Preview builds can include unpublished content:
go run ./md2hugo build --input examples --output dist/content.json --include-unpublished
```

The compiler recursively scans the input directory, so multiple files and nested folders are supported. Build output contains published content by default; `--include-unpublished` is intended for preview builds.

## Content types

Supported frontmatter `type` values:

- `collection`
- `learning-path`
- `certification`
- `tutorial`
- `quiz`
- `question`
- `flashcard`
- `challenge`
- `topic`

If `type` is omitted, the current MVP defaults to `tutorial` for backward compatibility.

## Publication lifecycle

Content supports:

```yaml
status: draft | review | published | archived
```

For backward compatibility, `draft: true/false` is still accepted and maps to `draft` / `published`.

Production CI can choose to emit only `published` content later; preview CI can include `draft` and `review` content.

## Relationships

Content can reference other content by stable ID:

```yaml
collections: [kubernetes-tutorials]
certifications: [cka]
prerequisites: [kubernetes-pods]
topics: [deployments, kubectl]
```

The compiler validates that referenced collection, certification, prerequisite, child, and activity IDs exist in the same content graph.

## Ratings and product artwork

Content may define separate IDs for ratings and shared product artwork:

```yaml
rating_id: docker-import-export
product_id: docker
```

`rating_id` identifies the individual content item used by the ratings API. `product_id` identifies reusable product artwork, so several Docker tutorials can share `product_id: docker` while keeping distinct rating IDs. Both fields are emitted unchanged as `rating_id` and `product_id` in `content.json`.

## Current block support

- YAML-like frontmatter for the current content conventions
- `draft: true/false` and explicit `status`
- Tutorial `{{< step ... >}}` extraction
- Markdown blocks
- `info`, `warn`, and `hint` fenced blocks
- Mermaid diagrams
- Terminal code blocks and attributes (`terminal`, `file-id`, `cmd-id`, `copy`)
- Duplicate content ID detection
- Recursive deterministic file traversal/order
- Cross-document relationship validation
- Single compiled content graph JSON artifact

## Example output

```json
{
  "version": 2,
  "generated": "...",
  "flashcardDecks": [],
  "challengeLabs": []
  "collections": [],
  "learningPaths": [],
  "certifications": [],
  "tutorials": [],
  "quizzes": [],
  "questions": [],
  "topics": [],
  "stats": {}
}
```

The generated file is intended to be consumed by the Certin Hugo Content Adapter, not rendered directly by Hugo as a page.

## Example templates

Starter templates are provided under `examples/templates/` for:

- collection
- challenge 
- flashcards 
- learning-path
- certification
- quiz
- question
- topic

A complete multi-type fixture is provided under `examples/test-content/`. It combines those records with the Docker tutorial example and is intended to exercise recursive discovery, cross-document relationships, and question data in one `content.json` artifact.

```bash
go run ./md2hugo validate --input examples/test-content
go run ./md2hugo build --input examples/test-content --output examples/test-content.json
```

Questions support `options`, a zero-based `answer` index, and an `explanation` in frontmatter. Quizzes reference question IDs using `activities`.

## Learning modalities

The compiler supports first-class `flashcard-deck` and `challenge-lab` content. A flashcard deck contains `flashcard` fenced blocks; a challenge lab contains `challenge` fenced blocks. Both are emitted into the single content graph as `flashcardDecks` and `challengeLabs`.

Example:

```text
```flashcard {id="image" front="What is a Docker image?"}
A read-only template used to create containers.
```
```

Challenge blocks may include `id`, `title`, `objective`, `command`, `hint`, and `success` attributes.

### Tutorial check blocks

Tutorials can include explicit learner progress checks using fenced code blocks:

```text
```check {id="verify-docker" message="Mark complete"}
```
```

The `id` is optional but recommended for stable progress tracking. The `message` defaults to `Mark complete`.
