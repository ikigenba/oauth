#!/bin/sh

set -eu

repo="https://github.com/ikigenba/oauth"
version="${OAUTH_VERSION:-latest}"
bindir="${BINDIR:-${PREFIX:-$HOME/.local}/bin}"

case "$(uname -s)" in
    Linux) os="linux" ;;
    Darwin) os="darwin" ;;
    *)
        echo "oauth: unsupported operating system: $(uname -s)" >&2
        exit 1
        ;;
esac

case "$(uname -m)" in
    x86_64 | amd64) arch="amd64" ;;
    arm64 | aarch64) arch="arm64" ;;
    *)
        echo "oauth: unsupported architecture: $(uname -m)" >&2
        exit 1
        ;;
esac

archive="oauth_${os}_${arch}.tar.gz"
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
install -m 0755 "$tmpdir/oauth" "$bindir/oauth"

case ":${PATH:-}:" in
    *:"$bindir":*) ;;
    *) echo "oauth: warning: $bindir is not on PATH" >&2 ;;
esac

echo "oauth installed to $bindir/oauth" >&2
