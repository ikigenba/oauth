# oauth

A standalone, provider-agnostic OAuth 2.0 login CLI. It runs the full
authorization-code + PKCE handshake for any protocol-compliant OAuth service —
serving its own loopback callback, opening the browser, exchanging the code —
and writes the token endpoint's JSON response verbatim to stdout, with all
human-facing output on stderr. It holds no provider-specific knowledge;
consumers own credential storage, refresh, and claim extraction. Module path:
`github.com/ikigenba/oauth`.

    oauth <flags> > ~/.foo/auth.json

## How changes are made

Changes go through the spec under `project/`, not direct edits — settle the
spec, then let the build loop realize it. The spec itself is direction-gated:
`project/**` is written only inside an operator-invoked move (the `$open-spec`
→ `$grill-me` → `$seal-spec` arc, or the build loop's completion mutations).
In any other session `project/` is read-only reference — a stale or wrong spec
is a finding to report, not a license to edit, and a settled discussion is not
direction: say what should change and wait. Edit code directly only on
explicit operator instruction. See the `$ikispec` skill for the `project/`
spec contracts and `$ralph` for the unattended build workflow.

## Layout

- `cmd/oauth/` — the single binary: flag parsing, composition root.
- `internal/oauth/` — the provider-neutral authorization-code + PKCE flow.
- `internal/callback/` — the loopback HTTP endpoint that receives the redirect.
- `internal/browser/` — per-OS launchers that open the authorize URL.
- `project/` — the spec (product/design/plan) the build loop works from.

## Tests

- Unit: `go test ./...`
- Green bar (`make check`): `go build ./...`, `go vet ./...`, `go test ./...`
  all exit 0 and `gofmt -l .` prints nothing.
- Live login smoke test (real provider, manual):
  `go test -tags live ./cmd/oauth/ -run TestLiveLogin -v`

## Versioning

Versions are annotated git tags only, `vMAJOR.MINOR.PATCH` — no `VERSION`
file. `cmd/oauth` carries `var version = "dev"` and the build stamps the real
value via `-ldflags "-X main.version=<v>"`; `oauth -V` prints it. Cut a release
with `git tag -a vX.Y.Z -m "vX.Y.Z"` on `main` and push the tag, which triggers
`.github/workflows/release.yml` to run goreleaser and publish the GitHub
Release that `install.sh` pulls from. Latest is
`git tag --sort=-v:refname | head -1`.
