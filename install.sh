#!/bin/sh
# Install shipquiz into ~/.shipquiz/bin and optionally run setup.
# Usage: curl -fsSL https://shipquiz.com/install.sh | sh
#        sh install.sh --no-setup
set -eu

no_setup=0
for arg in "$@"; do
  if [ "$arg" = "--no-setup" ]; then
    no_setup=1
  fi
done

os=$(uname -s)
case "$os" in
  Darwin) os=darwin ;;
  Linux) os=linux ;;
  *)
    echo "shipquiz: unsupported operating system: $os" >&2
    exit 1
    ;;
esac

arch=$(uname -m)
case "$arch" in
  x86_64|amd64) arch=amd64 ;;
  arm64|aarch64) arch=arm64 ;;
  *)
    echo "shipquiz: unsupported architecture: $arch" >&2
    exit 1
    ;;
esac

repo="${SHIPQUIZ_REPO:-pyroblue/shipquiz}"
version="${SHIPQUIZ_VERSION:-latest}"
asset="shipquiz_${os}_${arch}.tar.gz"
if [ "$version" = "latest" ]; then
  url="https://github.com/${repo}/releases/latest/download/${asset}"
else
  url="https://github.com/${repo}/releases/download/${version}/${asset}"
fi

dest="${SHIPQUIZ_HOME:-$HOME/.shipquiz}"
mkdir -p "$dest/bin"

tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

echo "shipquiz: downloading ${asset}"
if ! curl -fsSL "$url" -o "$tmp/shipquiz.tar.gz"; then
  echo "shipquiz: download failed: $url" >&2
  echo "shipquiz: no release is published yet, or this platform is missing." >&2
  echo "Build from source instead:" >&2
  echo "  go build -o \"\$HOME/.shipquiz/bin/shipquiz\" ./cmd/shipquiz" >&2
  exit 1
fi

tar -xzf "$tmp/shipquiz.tar.gz" -C "$tmp"
if [ ! -f "$tmp/shipquiz" ]; then
  echo "shipquiz: archive did not contain a shipquiz binary" >&2
  exit 1
fi

install -m 755 "$tmp/shipquiz" "$dest/bin/shipquiz"

path_line='export PATH="$HOME/.shipquiz/bin:$PATH"'
updated=0
for rc in "$HOME/.zshrc" "$HOME/.bashrc"; do
  if [ -f "$rc" ]; then
    if ! grep -F '.shipquiz/bin' "$rc" >/dev/null 2>&1; then
      printf '\n# shipquiz\n%s\n' "$path_line" >> "$rc"
    fi
    updated=1
  fi
done
if [ "$updated" -eq 0 ]; then
  case "${SHELL:-}" in
    *zsh) rc="$HOME/.zshrc" ;;
    *) rc="$HOME/.bashrc" ;;
  esac
  if [ ! -f "$rc" ] || ! grep -F '.shipquiz/bin' "$rc" >/dev/null 2>&1; then
    printf '\n# shipquiz\n%s\n' "$path_line" >> "$rc"
  fi
fi

echo "shipquiz: installed to $dest/bin/shipquiz"
echo "This shell will not see it until you open a new terminal, or run:"
echo "  export PATH=\"\$HOME/.shipquiz/bin:\$PATH\""

if [ "$no_setup" -eq 1 ] || [ -n "${CI:-}" ]; then
  exit 0
fi

if [ -e /dev/tty ]; then
  "$dest/bin/shipquiz" setup </dev/tty
fi
