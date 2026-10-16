# Release checklist

A release is ready only when all of the following are true.

- [ ] `VERSION`, the CLI `-version` output and the release tag agree.
- [ ] `gofmt` produces no changes.
- [ ] `go vet ./...` passes.
- [ ] `go test -race ./...` passes.
- [ ] the Bafflewick acceptance case passes.
- [ ] `GOOS=linux GOARCH=arm64 go build ./cmd/...` passes.
- [ ] every PRD acceptance criterion is represented by a test or explicit acceptance check.
- [ ] `EDD.md` still describes the intended design.
- [ ] `generated/EDD.md` has been regenerated from the release candidate source commit.
- [ ] `generated/COMPARISON.md` contains no unexplained mismatch.
- [ ] known limitations remain explicit rather than being implied as implemented.
- [ ] `CHANGELOG.md` records the release.

Tag format: `v0.1.0`, `v0.2.0`, and so on.
