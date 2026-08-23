# macOS install (unsigned)

Releases ship a DMG. The app is not notarized.

1. Open the DMG and drag `Moyan.app` to Applications.
2. Clear Gatekeeper quarantine:

```bash
xattr -cr /Applications/Moyan.app
```

3. Open Moyan from Applications. If macOS still blocks it, Control-click the app and choose Open.

To cut a release: push a tag `v*` (example `v0.1.0`). GitHub Actions runs tests, builds a universal Darwin binary, and attaches the DMG to the GitHub Release.
