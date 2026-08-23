# Open one 词典包 at a time

**Status:** accepted

Moyan currently opens a single user-supplied 词典包. Importing another package replaces the open one. Lookup returns at most one 词条.

This supersedes the “supports multiple dictionaries” clause of ADR-0001. Concurrent 词典包 will be revisited only after one-package import, lookup, candidates, and rendering are reliable.

## Consequences

The Wails `App` holds one `dictionary.Reader`. The 应用外壳 does not present enable/disable or multi-package result lists. A failed import must leave the current 词典包 open.
