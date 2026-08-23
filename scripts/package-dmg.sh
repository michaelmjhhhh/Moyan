#!/usr/bin/env bash
set -euo pipefail

root="$(cd "$(dirname "$0")/.." && pwd)"
app="$root/build/bin/Moyan.app"
if [[ ! -d "$app" ]]; then
  echo "missing $app" >&2
  exit 1
fi

version="${GITHUB_REF_NAME:-local}"
out_dir="$root/build/releases"
mkdir -p "$out_dir"
stage="$(mktemp -d)"
trap 'rm -rf "$stage"' EXIT

cp -R "$app" "$stage/Moyan.app"
ln -s /Applications "$stage/Applications"

dmg="$out_dir/Moyan-${version}-macos.dmg"
rm -f "$dmg"
hdiutil create \
  -volname "Moyan" \
  -srcfolder "$stage" \
  -ov \
  -format UDZO \
  "$dmg"
echo "wrote $dmg"
