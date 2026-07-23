# Phase 07 — The `-V` version flag

*Realizes design Decision 6 (CLI surface), the version slice
(R-9CGI-VDPQ, R-9DOF-95GF, R-9EWB-MX74).*

`cmd/oauth-login` gains a package-level `var version = "dev"` and a `-V`
terminal action. `-V` short-circuits before any login: it writes `version`
(with a trailing newline) to stdout and exits 0, binding no listener and
opening no browser. `--help` gains a `-V` line describing it alongside `-h`.
The version symbol is the one the build stamps via
`-ldflags "-X main.version=<v>"`; nothing in this phase stamps it, so an
ordinary `go test`/`go run` build reports `dev`.

## Done when

Every id below is covered by a clearly-named test in `cmd/oauth-login`
(`*_test.go`), and the suite is green (`project/design/README.md` →
Conventions):

- R-9CGI-VDPQ — In an unstamped build, `-V` writes exactly `dev\n` to stdout,
  exits 0, and touches neither listener nor browser.
- R-9DOF-95GF — A binary compiled with `-ldflags "-X main.version=<sentinel>"`
  prints that sentinel (not `dev`) for `-V`. The test builds a real stamped
  binary and runs it, exercising the linker path end to end.
- R-9EWB-MX74 — `--help` exits 0 and its output names the `-V` flag together
  with its description.
