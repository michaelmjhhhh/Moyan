# Render dictionary content as untrusted data

**Status:** accepted

Dictionary HTML, CSS, and resources are untrusted user-supplied content. Moyan will render them in a scriptless, isolated surface with a restrictive CSP and a capability-authenticated resource handler; it will never expose Wails bindings, arbitrary filesystem paths, remote requests, or dictionary JavaScript to the application shell. This deliberately prefers a secure and mostly faithful reading surface over unrestricted compatibility.

## Consequences

Resource paths must be canonicalized and contained within the selected dictionary package, with MIME, size, decompression, and response limits. External links are disabled, malformed or over-limit packages are rejected with diagnostics, and any future compatibility mode requires a new security review rather than silently weakening the default boundary.
