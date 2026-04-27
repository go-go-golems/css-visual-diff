# Changelog

## 2026-04-27

- Initial workspace created


## 2026-04-27

Created local verb repository config support design guide and investigation diary.

### Related Files

- /home/manuel/code/wesen/corporate-headquarters/css-visual-diff/ttmp/2026/04/27/CSSVD-LOCAL-VERB-REPOS--add-local-verb-repository-config-support/design-doc/01-local-verb-repository-config-analysis-design-and-implementation-guide.md — Primary design deliverable
- /home/manuel/code/wesen/corporate-headquarters/css-visual-diff/ttmp/2026/04/27/CSSVD-LOCAL-VERB-REPOS--add-local-verb-repository-config-support/reference/01-investigation-diary.md — Chronological investigation record


## 2026-04-27

Validated ticket with docmgr doctor and uploaded design/diary bundle to reMarkable.

### Related Files

- /home/manuel/code/wesen/corporate-headquarters/css-visual-diff/ttmp/2026/04/27/CSSVD-LOCAL-VERB-REPOS--add-local-verb-repository-config-support/reference/01-investigation-diary.md — Validation and upload evidence


## 2026-04-27

Implemented local verb repository config discovery and committed code as 5fd1c68519662dafbccf1dc34cb05e90298eba32.

### Related Files

- /home/manuel/code/wesen/corporate-headquarters/css-visual-diff/internal/cssvisualdiff/verbcli/bootstrap.go — Local config discovery implementation
- /home/manuel/code/wesen/corporate-headquarters/css-visual-diff/internal/cssvisualdiff/verbcli/bootstrap_test.go — Local config discovery tests
- /home/manuel/code/wesen/corporate-headquarters/css-visual-diff/internal/cssvisualdiff/verbcli/command_test.go — Lazy command smoke test


## 2026-04-27

Fixed repository lint issues so make lint passes; committed as 2ff8de408c9812c9006ef8600d20f4f82fdde8f2.

### Related Files

- /home/manuel/code/wesen/corporate-headquarters/css-visual-diff/internal/cssvisualdiff/modes/html_report.go — Converted linted string formatting patterns
- /home/manuel/code/wesen/corporate-headquarters/css-visual-diff/internal/cssvisualdiff/modes/inspect.go — Documented retained legacy helpers with targeted nolint comments
- /home/manuel/code/wesen/corporate-headquarters/css-visual-diff/internal/cssvisualdiff/service/catalog_service.go — Converted linted markdown formatting patterns


## 2026-04-27

Adjusted CI unit tests to run packages serially with GOWORK=off go test -p 1 ./... after GitHub Actions Chrome websocket timeout.

### Related Files

- /home/manuel/code/wesen/corporate-headquarters/css-visual-diff/.github/workflows/push.yml — Serializes package tests to reduce CI Chrome startup contention
- /home/manuel/code/wesen/corporate-headquarters/css-visual-diff/ttmp/2026/04/27/CSSVD-LOCAL-VERB-REPOS--add-local-verb-repository-config-support/reference/01-investigation-diary.md — Records CI failure analysis and validation

