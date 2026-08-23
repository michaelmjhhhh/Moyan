# Moyan

Offline desktop dictionary reader. Reads user-supplied MDX/MDD v1/v2 dictionary packages on disk. Does not ship dictionary content.

Target platforms: Windows and macOS. License: GPLv3.

Lookup, indexing, rendering, and resource access stay on-device. The process reaches the network only when the user starts an update check.

## Scope

Supported:

- Import a dictionary package: an `.mdx` file plus optional sibling `.mdd`, CSS, images, and audio. Files are referenced in place.
- Library: ordered list of imported package paths. Import appends. A failed import does not change existing records. Paths are written to `<user-config-dir>/Moyan/library.json` and restored on start.
- Lookup: exact, prefix, and lightweight tolerant match. One query may produce one entry per open package. Sidebar selection chooses which entry the reading pane shows.
- Entry HTML/CSS and package resources render as untrusted, scriptless content.

Unsupported:

- Protected (encrypted) packages and MDX/MDD v3
- Dictionary-authored scripts
- Full-text search, global lookup, bookmarks, export, sync
- Bundled or downloaded dictionary content

`internal/mdictcore` is a temporary source-donor parser. Parser hardening, bounded allocation, checksum validation, and fixture coverage are required before arbitrary user-supplied packages are treated as safe to open.

## Tree

| Path | Role |
| --- | --- |
| `pkg/dictionary` | Public reader: `Open`, `Lookup`, `Candidates` |
| `internal/mdictcore` | MDX/MDD parser (Medict-derived) |
| `app.go` | Wails bindings and library persistence |
| `frontend/` | React + TypeScript + Vite application shell |

## Build

Requires Go 1.25, Node 22, and Wails v2.10.2.

```bash
npm --prefix frontend ci
go test ./...
cd frontend && npm test && cd ..
wails build
```

macOS CI builds a universal Darwin binary on tag `v*` (`.github/workflows/release.yml`). Artifacts are unsigned; install steps are in `docs/macos-install.md`.

## Third-party sources

### MDX/MDD format

Moyan consumes the MDict dictionary-package layout: `.mdx` (headwords and records) and optional `.mdd` (binary resources). That format originates with the MDict software published by 上海卓越电子科技有限公司. This repository does not include MDict itself.

### Parser (Medict)

`internal/mdictcore` is derived from Medict:

- Repository: https://github.com/terasum/medict
- Commit: [`04f572a6258997125d6382486598e4c7d5018ea7`](https://github.com/terasum/medict/commit/04f572a6258997125d6382486598e4c7d5018ea7)
- Copyright: Quan Chen, 2023
- License: GNU GPL v3 (notices retained in the Go source files)

See `internal/mdictcore/NOTICE.md`. Medict’s Vue UI is not used.

Direct Go modules used by the parser:

- [`github.com/c0mm4nd/go-ripemd`](https://github.com/c0mm4nd/go-ripemd) — RIPEMD
- [`github.com/rasky/go-lzo`](https://github.com/rasky/go-lzo) — LZO decompression
- [`github.com/op/go-logging`](https://github.com/op/go-logging) — logging
- [`golang.org/x/text`](https://pkg.go.dev/golang.org/x/text) — text encodings

### Application shell

- [Wails v2](https://github.com/wailsapp/wails) `v2.10.2` — Go desktop host and system webview
- [React](https://github.com/facebook/react) `19` + [react-dom](https://www.npmjs.com/package/react-dom) — application shell
- [Vite](https://github.com/vitejs/vite) `6` + [@vitejs/plugin-react](https://github.com/vitejs/vite-plugin-react) — frontend build
- [Newsreader](https://fonts.google.com/specimen/Newsreader) via [`@fontsource-variable/newsreader`](https://fontsource.org/fonts/newsreader) — UI typeface

### Test fixture (CC-CEDICT)

`pkg/dictionary/testdata/cc-cedict.mdx` is a small compatibility fixture derived from CC-CEDICT as carried in Medict’s test/preset data. It is not a bundled user dictionary.

- Title: CC-CEDICT
- Publisher: MDBG
- License: Creative Commons Attribution-ShareAlike 3.0 Unported
- Source: https://www.mdbg.net/chinese/dictionary?page=cc-cedict
- Local note: `pkg/dictionary/testdata/README.md`

## References

1. Chen Q. *Medict*. GitHub. https://github.com/terasum/medict. Commit `04f572a6258997125d6382486598e4c7d5018ea7`.
2. 上海卓越电子科技有限公司. *MDict* dictionary package format (MDX/MDD). https://www.mdict.cn
3. MDBG. *CC-CEDICT*. Creative Commons Attribution-ShareAlike 3.0 Unported. https://www.mdbg.net/chinese/dictionary?page=cc-cedict
4. Wails. *Wails v2*. https://wails.io. Module `github.com/wailsapp/wails/v2` v2.10.2.
5. Meta Platforms, Inc. *React*. https://react.dev. Version 19.
6. VoidZero / Vite contributors. *Vite*. https://vite.dev. Version 6.
7. Coles T, Reid E. *Newsreader* typeface. SIL Open Font License 1.1. https://fonts.google.com/specimen/Newsreader
8. Free Software Foundation. *GNU General Public License, version 3*. https://www.gnu.org/licenses/gpl-3.0.html
