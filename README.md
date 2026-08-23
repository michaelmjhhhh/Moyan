# Moyan

Moyan is an offline desktop dictionary reader. It opens dictionary packages you already have on disk, looks up headwords, and shows the matching entry. It does not ship any dictionary content of its own.

A dictionary package is an `.mdx` file plus any sibling resources that package needs: an optional `.mdd`, CSS, images, and audio. Moyan references those files in place. It does not copy them into an internal store.

Supported platforms are Windows and macOS. The application is GPLv3.

Lookup, indexing, rendering, and resource access stay on-device. The only network use is an update check, and only when the user starts it.

## Behavior

Import appends a package path to the library. The library is the ordered list of imported paths. A failed import does not change existing records. Duplicate paths are not opened twice. On start, Moyan restores the list from `<user-config-dir>/Moyan/library.json`; missing files on restore are skipped.

A query may produce one entry per open package. The sidebar selects which package’s entry the reading pane shows. Lookup is exact, prefix, and lightweight tolerant match. It is not full-text search.

Entry HTML, CSS, and package resources are untrusted user content. They render in a scriptless surface. Dictionary-authored scripts are not executed. External links are not followed. Protected (encrypted) packages and MDX/MDD v3 are rejected.

Out of scope: bundled or downloaded dictionary content, global lookup, bookmarks, export, and sync.

The MDX/MDD parser in `internal/mdictcore` is a source-donor copy. Parser hardening, bounded allocation, checksum validation, and fixture coverage are still required before arbitrary user-supplied packages should be treated as safe to open.

## Implementation

The desktop host is Wails v2 (Go plus the system webview). The application shell is React, TypeScript, and Vite.

| Path | Role |
| --- | --- |
| `pkg/dictionary` | Public reader: `Open`, `Lookup`, `Candidates` |
| `internal/mdictcore` | MDX/MDD parser |
| `app.go` | Wails bindings and library persistence |
| `frontend/` | Application shell |

## Build

Requires Go 1.25, Node 22, and Wails v2.10.2.

```bash
npm --prefix frontend ci
go test ./...
cd frontend && npm test && cd ..
wails build
```

Pushing a `v*` tag runs `.github/workflows/release.yml`, which tests, builds a universal Darwin binary, and attaches an unsigned DMG to the GitHub Release. Install steps for that artifact are in `docs/macos-install.md`.

## Third-party sources

Moyan reads the MDict package layout (`.mdx` headwords and records, optional `.mdd` binary resources). That format comes from MDict, published by 上海卓越电子科技有限公司. This repository does not include MDict.

The parser in `internal/mdictcore` is derived from Medict (Quan Chen, GPLv3) at commit [`04f572a6258997125d6382486598e4c7d5018ea7`](https://github.com/terasum/medict/commit/04f572a6258997125d6382486598e4c7d5018ea7). Original notices remain in the Go files; see `internal/mdictcore/NOTICE.md`. Medict’s UI is not used.

`pkg/dictionary/testdata/cc-cedict.mdx` is a small compatibility fixture derived from [CC-CEDICT](https://www.mdbg.net/chinese/dictionary?page=cc-cedict) (MDBG, Creative Commons Attribution-ShareAlike 3.0), as carried in Medict’s test data. It is not a bundled user dictionary. See `pkg/dictionary/testdata/README.md`.
