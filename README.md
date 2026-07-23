# oauth

A standalone, provider-agnostic OAuth 2.0 login CLI. It runs the full
authorization-code + PKCE handshake against any protocol-compliant service:
it serves its own loopback callback endpoint, opens your browser, exchanges
the code, and writes the token endpoint's JSON response **verbatim to stdout**.

It holds no provider-specific knowledge — a service is described entirely by
flags. Everything meant for a human (the authorize URL, progress, errors) goes
to stderr, so stdout can be redirected straight to a file. A failed login
writes nothing to stdout and exits non-zero.

## Install

Prebuilt binary (no Go toolchain needed):

    curl -fsSL https://raw.githubusercontent.com/ikigenba/oauth/main/install.sh | sh

Installs the latest release to `~/.local/bin` (override with `BINDIR=/path` or
`PREFIX=/path`; pin a version with `OAUTH_VERSION=v0.1.0`). Make sure the
target dir is on your `PATH`.

From source (requires Go):

    make build                 # -> ./bin/oauth
    make install               # -> ~/.local/bin/oauth
    make install PREFIX=/usr/local

Check the installed version with `oauth -V`.

## Usage

    oauth \
      --auth-url  https://auth.example.com/oauth/authorize \
      --token-url https://auth.example.com/oauth/token \
      --client-id your-client-id \
      --scope "openid profile offline_access" \
      > auth.json

The callback address must match what the provider has registered for your
client. Use `--port` and `--callback-path` to match it exactly, and
`--callback-host` if the registration uses `127.0.0.1` rather than `localhost`.

For confidential clients, pass `--client-secret`, or supply an `Authorization`
header with `--token-header`. Providers that require extra parameters are
handled with the repeatable `--auth-param` and `--token-param` escape hatches.

Run `oauth --help` for the full flag list and a worked example.
