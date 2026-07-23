#!/bin/sh

set -eu

repo="https://github.com/ikigenba/login"
version="${OAUTH_LOGIN_VERSION:-latest}"
bindir="${BINDIR:-${PREFIX:-$HOME/.local}/bin}"

case "$(uname -s)" in
    Linux) os="linux" ;;
    Darwin) os="darwin" ;;
    *)
        echo "login: unsupported operating system: $(uname -s)" >&2
        exit 1
        ;;
esac

case "$(uname -m)" in
    x86_64 | amd64) arch="amd64" ;;
    arm64 | aarch64) arch="arm64" ;;
    *)
        echo "login: unsupported architecture: $(uname -m)" >&2
        exit 1
        ;;
esac

archive="login_${os}_${arch}.tar.gz"
if [ "$version" = "latest" ]; then
    url="$repo/releases/latest/download/$archive"
else
    url="$repo/releases/download/$version/$archive"
fi

tmpdir="$(mktemp -d)"
trap 'rm -rf "$tmpdir"' 0
trap 'exit 1' HUP INT TERM

curl -fsSL "$url" -o "$tmpdir/$archive"
tar -xzf "$tmpdir/$archive" -C "$tmpdir"
install -d "$bindir"
install -m 0755 "$tmpdir/login" "$bindir/login"

case ":${PATH:-}:" in
    *:"$bindir":*) ;;
    *) echo "login: warning: $bindir is not on PATH" >&2 ;;
esac

echo "login installed to $bindir/login" >&2
