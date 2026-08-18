# Moyan MVP boundaries

**Status:** accepted

Moyan is a GPLv3, offline-first desktop reader for user-supplied MDX/MDD v1/v2 dictionary packages on Windows and macOS. The MVP references source files in place, supports multiple dictionaries with exact/prefix/lightweight tolerant lookup, recent history, bounded local static resources and click-to-play audio, but does not distribute dictionary content, support protected/v3 files, full-text search, global lookup, bookmarks, export, sync, or dictionary-authored scripts. This keeps copyright ownership separate from the reader and makes a small, testable product possible.

## Consequences

The compatibility corpus consists of redistributable fixtures plus private, lawfully obtained local dictionaries. The app may make a network request only after the user explicitly requests an update check; dictionary lookup and rendering never connect to the network.
