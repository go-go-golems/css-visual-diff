# Tasks

## Done

- [x] Create docmgr ticket workspace for opaque Go-backed JavaScript type review.
- [x] Scan the repository for raw JavaScript object / `map[string]any` Go boundaries.
- [x] Review existing Go-backed Proxy handles/builders and strict unwrapping APIs.
- [x] Write intern-facing design and implementation guide.
- [x] Write investigation diary.
- [x] Relate key source files to the design document.
- [x] Validate the ticket with `docmgr doctor`.
- [x] Upload the ticket bundle to reMarkable.

## Follow-up implementation tasks

- [ ] Add tests for current legacy raw object compatibility and strict raw-object rejection.
- [ ] Refactor `jsapi` to use per-module/per-runtime `ProxyRegistry` state.
- [ ] Add `cvd.prepare.*` builders and support them in `page.prepare(...)` and `target.prepare(...)`.
- [ ] Allow `page.preflight(...)`, `page.inspect(...)`, and `page.inspectAll(...)` to accept `cvd.probe(...)` builders.
- [ ] Add `cvd.inspectOptions()` builder.
- [ ] Add `cvd.catalogTarget(...)` or equivalent catalog target builder support.
- [ ] Improve probe/extractor builders with `.optional()`, `.exists()`, `.visible()`, `.text(options)`, and `.extract(...)`.
