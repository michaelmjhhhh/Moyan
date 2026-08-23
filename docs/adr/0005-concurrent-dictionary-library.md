# Concurrent 词典包 in the 词库

**Status:** accepted

Moyan keeps multiple imported 词典包 open. Import appends; a failed import does not close or reorder existing packages. The 词库 is the ordered list of package paths, written to a durable file and restored on process start.

Lookup and candidates use `pkg/dictionary` `Open` / `Lookup` / `Candidates` per path. The 应用外壳 lists packages in the library sidebar; selecting a package shows that package’s 词条 or miss for the current query.

## Consequences

`App` holds one `dictionary.Reader` per imported path. Duplicate paths are not opened twice. CloseDictionary releases readers in memory and does not erase the persisted path list. Missing files on restore are skipped.

This supersedes ADR-0004.
