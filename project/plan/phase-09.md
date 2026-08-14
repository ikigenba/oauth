# Phase 09 — xAI worked example in `--help`

*Realizes design Decision 6 (CLI surface), the xAI help-example slice
(R-VS94-25B6).*

`cmd/oauth`'s `--help` text carries two labeled worked examples. The OpenAI
example is unchanged. After it, help prints the xAI invocation:

```
oauth \
  --auth-url  https://auth.x.ai/oauth2/authorize \
  --token-url https://auth.x.ai/oauth2/token \
  --client-id b1a00492-073a-47ea-816f-4c329264a828 \
  --scope "openid profile email offline_access grok-cli:access api:access" \
  --callback-host 127.0.0.1 \
  --port 56121 \
  --callback-path /callback \
  > x-ai-auth.json
```

The Basic authentication snippet still follows both examples. No flag, default,
or login behavior changes.

## Done when

The id below is covered by a clearly-named test in `cmd/oauth` (`*_test.go`),
and the suite is green (`project/design/README.md` → Conventions):

- R-VS94-25B6 — `--help` exits 0 and its output contains the xAI worked
  example's authorize URL, token URL, client id, `--callback-host 127.0.0.1`,
  `--port 56121`, `--callback-path /callback`, `grok-cli:access`,
  `api:access`, and `> x-ai-auth.json`.
