# clz-md2hugo

[![Go Version](https://img.shields.io/github/go-mod/go-version/cmdlinezero/clz-md2hugo)](https://golang.org)
[![License](https://img.shields.io/github/license/cmdlinezero/clz-md2hugo)](LICENSE)
[![GitHub release](https://img.shields.io/github/v/release/cmdlinezero/clz-md2hugo?include_prereleases)](https://github.com/cmdlinezero/clz-md2hugo/releases)

`clz-md2hugo` is a Go content compiler for Certin-authored Markdown. It parses source metadata and tutorial structure into a portable JSON content contract that can be consumed by Hugo/Certin Portal and other renderers. The compiler owns content structure and portable semantics; downstream renderers own presentation.

---

## 🚀 Features

- **Portable Content Compilation:** Converts authored Markdown and front matter into a renderer-independent JSON content graph.
- **Lossless Markdown Blocks:** Preserves ordinary Markdown syntax inside `markdown` blocks; presentation is deferred to the downstream renderer.
- **Structured Semantic Blocks:** Extracts semantic fences such as checks, code, hints, callouts, and diagrams into typed blocks.
- **Explicit Lab Capabilities:** Normalizes declared lab modalities into a portable `lab` contract without inferring modality from content.
- **Catalogue & Payload Projection:** Produces discovery metadata and tiered payload artifacts while preserving access boundaries.
- **Fast & Lightweight:** Built in Go with minimal external dependencies.

---

## 📦 Installation

### Option 1: Using `go install` (Recommended)

Make sure you have Go installed (version 1.18+ recommended):

```bash
go install https://github.com/cmdlinezero/clz-md2hugo@latest

```

### Option 2: Build from Source

```bash
# Clone the repository
git clone https://github.com/cmdlinezero/clz-md2hugo.git
cd clz-md2hugo

# Build the binary
go build -o clz-md2hugo .

# Optional: Move to system PATH
mv clz-md2hugo /usr/local/bin/

```

---

## 🛠️ Usage

### Quick Start

Convert a single file:

```bash
clz-md2hugo path/to/article.md -o content/posts/

```

Process an entire directory of Markdown files:

```bash
clz-md2hugo -i path/to/notes/ -o content/posts/

```

### Options & Flags

| Flag | Short | Description | Default |
| --- | --- | --- | --- |
| `--input` | `-i` | Input Markdown file or directory path | `./` |
| `--output` | `-o` | Output directory path for Hugo posts | `./content` |
| `--format` | `-f` | Front matter format (`yaml` or `toml`) | `yaml` |
| `--draft` | `-d` | Set output posts as draft (`true`/`false`) | `false` |
| `--help` | `-h` | Display help information | — |

---

## 📁 Example Output

Given a input file `my-note.md`:

```markdown
# My First Note
This is the body content of the note.

```

`clz-md2hugo` generates `content/posts/my-note.md`:

```yaml
---
title: "My First Note"
date: 2026-04-21T10:00:00Z
draft: false
---

This is the body content of the note.

```

---

## 🛠️ Development & Building

Run tests locally:

```bash
go test ./...

```

Run linter:

```bash
golangci-lint run

```

## Build a Catalogue

Create a catalogue based on the folder containing the content to be generated.

```bash
 ./clz-md2hugo build \
   -i examples/certin-site \
   -o examples/generated/certin-site.json
```
## Access tiers (content graph v3)

Content can declare an access tier in frontmatter. If omitted, the tier defaults to `public`.

```yaml
access:
  tier: pro
```

Supported tiers are `public`, `member`, and `pro`. Access is resolved by `md2hugo`, not by Hugo. An item without an explicit tier inherits the most restrictive tier from its containment parents (`public < member < pro`); an explicit child tier overrides inherited access. Collection membership, `children`, and quiz `activities` are containment relationships and participate in inheritance. `certifications` is an alignment/association relationship and does not propagate access. Certifications may themselves be contained by collections through `collections`, and may contain other nodes (including collections) through `children`.

The generated canonical graph remains complete and now reports `"version": 3`. Public/protected projections are intentionally outside this milestone.

## Public catalogue and protected payload projections

The canonical `content.json` graph is an internal build artifact. Do not publish it directly: it contains full tutorial, assessment, flashcard, and lab payloads regardless of access tier.

Use `project` to build a public-safe catalogue separately from full payloads:

```bash
certin-content project \
  --input examples/content \
  --catalogue dist/catalogue.json \
  --payload-dir private/payloads
```

This produces:

- `dist/catalogue.json` — safe discovery metadata for public, member, and pro content. It includes effective `access.tier` and relationship IDs, but excludes tutorial steps/blocks, question options/answers/explanations, flashcard card payloads, and challenge payloads.
- `private/payloads/public.json` — full payload for content whose effective tier is `public`.
- `private/payloads/member.json` — full payload for content whose effective tier is `member`.
- `private/payloads/pro.json` — full payload for content whose effective tier is `pro`.

`member.json` and `pro.json` are backend/private artifacts. Their directory must not be copied into Hugo's `static`, `public`, or other web-served output. Authentication/entitlement enforcement remains the responsibility of the serving backend; md2hugo only compiles the effective access classification and artifact boundary.



## Portable lab contract

Tutorials that explicitly declare `learning_mode: lab` compile their modality metadata into a normalized, renderer-independent `lab` object. For example, a Codelab-only tutorial emits:

```json
{
  "lab": {
    "enabled": true,
    "modes": ["codelab"],
    "default_mode": "codelab"
  }
}
```

The source mapping is:

- `learning_mode: lab` → `lab.enabled`
- `lab_modes` → `lab.modes`
- `default_lab_mode` → `lab.default_mode`

`default_mode` must be one of the declared modes, and a lab must declare at least one supported mode. Current supported modes are `codelab` and `terminal`. Non-lab tutorials omit the `lab` object.

Lab modality is **explicit metadata**. The compiler never infers Codelab support from `check` blocks and never infers Terminal support from code or terminal blocks. Content features and runtime modality are separate concerns. Catalogue projection carries the same normalized `lab` contract so discovery metadata and full content payloads agree on runtime capabilities.

## Portable Markdown blocks

Tutorial Markdown remains Markdown in the compiled content contract. `clz-md2hugo` determines structure but does not determine presentation. Ordinary authored Markdown is preserved inside blocks such as:

```json
{
  "type": "markdown",
  "content": "## The nine step method\n\n1. **Classify the workload:** determine the activity."
}
```

Semantic fenced blocks are extracted separately into typed structures such as `check`, `code`, `hint`, `callout`, and `diagram`. This extraction does not authorize the compiler to strip presentation syntax from surrounding ordinary Markdown.

Downstream consumers such as Certin Portal/Hugo own Markdown-to-HTML presentation. In other words: **md2hugo determines structure; Hugo determines presentation.** See `MARKDOWN-CONTRACT.md` for the detailed boundary and regression expectations.

---

## Go coding tutorial smoke test

Coding tutorials may declare `runtime: go` with an entrypoint such as `main.go`.
The canonical Go/WASM smoke-test source is
`examples/test-content/coding-tutorials/go-first-program.md`.

Coding-step check fences use `output="..."`; the compiler projects this to the
portable `expected_output` field consumed by the Portal runtime. The smoke test
checks both basic `fmt.Println` execution and standard-library
`strings.ToUpper` execution.

## 🤝 Contributing

Contributions are welcome! If you'd like to help improve `clz-md2hugo`:

1. Fork the Repository
2. Create your Feature Branch (`git checkout -b feature/AmazingFeature`)
3. Commit your Changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the Branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

---

## 📄 License

Distributed under the GNU License. See `LICENSE` for more information.
