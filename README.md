# clz-md2hugo

[![Go Version](https://img.shields.io/github/go-mod/go-version/cmdlinezero/clz-md2hugo)](https://golang.org)
[![License](https://img.shields.io/github/license/cmdlinezero/clz-md2hugo)](LICENSE)
[![GitHub release](https://img.shields.io/github/v/release/cmdlinezero/clz-md2hugo?include_prereleases)](https://github.com/cmdlinezero/clz-md2hugo/releases)

`clz-md2hugo` is a lightweight CLI tool written in Go that converts standard Markdown files into Hugo-compatible content files with front matter generation.

---

## 🚀 Features

- **Automated Front Matter Injection:** Automatically formats titles, dates, tags, and categories for Hugo content folders.
- **Fast & Lightweight:** Built in Go with minimal external dependencies.
- **Batch Processing:** Convert single Markdown files or process entire directories recursively.
- **Customizable Front Matter:** Supports YAML/TOML metadata formatting.

---

## 📦 Installation

### Option 1: Using `go install` (Recommended)

Make sure you have Go installed (version 1.18+ recommended):

```bash
go install [github.com/cmdlinezero/clz-md2hugo@latest](https://github.com/cmdlinezero/clz-md2hugo@latest)

```

### Option 2: Build from Source

```bash
# Clone the repository
git clone [https://github.com/cmdlinezero/clz-md2hugo.git](https://github.com/cmdlinezero/clz-md2hugo.git)
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

---

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
