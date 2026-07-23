# Phase 08 — Build, release, and install tooling

*Realizes design Decision 8 (build/release/install tooling; structural, no
ids). Depends on Phase 07.*

Four artifacts land at the repository root, standardized against the sibling
repositories:

- `Makefile` — the existing targets, with the `build` recipe now stamping
  `-ldflags "-X main.version=$(shell git describe --tags --always --dirty)"`.
  `check` and the `$(wildcard go.sum)` prerequisite are retained.
- `install.sh` — the `curl … | sh` installer that fetches
  `oauth-login_<os>_<arch>.tar.gz` from GitHub releases (`latest` or
  `$OAUTH_LOGIN_VERSION`) and installs it under
  `${BINDIR:-${PREFIX:-$HOME/.local}/bin}`.
- `.goreleaser.yml` — goreleaser v2 building linux/darwin × amd64/arm64
  archives with versionless names, stamping `-X main.version={{ .Tag }}`.
- `.github/workflows/release.yml` — on a pushed `v*` tag, runs
  `goreleaser release --clean`.

Depends on Phase 07 because the end-to-end done-bar drives the stamped binary
through `-V`.

## Done when

Structural — deterministic exit conditions, no id tags:

- `go build ./...`, `go vet ./...`, `go test ./...` exit 0 and `gofmt -l .`
  prints nothing (the green bar; equivalently `make check`).
- The four files exist: `Makefile`, `install.sh`, `.goreleaser.yml`,
  `.github/workflows/release.yml`.
- `sh -n install.sh` exits 0 (the installer parses as valid POSIX `sh`).
- `make build` produces `bin/oauth-login`, and its stamped version matches the
  checkout: `test "$(bin/oauth-login -V)" = "$(git describe --tags --always --dirty)"`
  exits 0 — proving the `Makefile` ldflags stamp and Phase 07's `-V` action
  agree end to end.
- The `Makefile` `build` recipe contains `-X main.version=` (grep the
  `Makefile`, outside `project/`).
