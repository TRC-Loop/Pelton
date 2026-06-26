#!/usr/bin/env bash
# collects license texts of all direct go + npm deps into licenses/
set -euo pipefail

root="$(cd "$(dirname "$0")/.." && pwd)"
cd "$root"

find_license() {
  # prints first license-like filename in dir $1, empty if none
  ls "$1" 2>/dev/null | grep -iE '^(LICENSE|LICENCE|COPYING|UNLICENSE)' | head -1 || true
}

echo "==> go modules"
rm -rf licenses/go && mkdir -p licenses/go
go list -m -f '{{if and (not .Indirect) (not .Main)}}{{.Path}}|{{.Dir}}{{end}}' all \
  | grep -v '^$' \
  | while IFS='|' read -r path dir; do
      [ -d "$dir" ] || continue
      lic="$(find_license "$dir")"
      safe="$(printf '%s' "$path" | tr '/' '_')"
      if [ -n "$lic" ]; then
        cp "$dir/$lic" "licenses/go/${safe}.txt"
        echo "  ok   $path"
      else
        echo "  miss $path"
      fi
    done

echo "==> npm packages"
rm -rf licenses/npm && mkdir -p licenses/npm
nm="frontend/node_modules"
if [ -d "$nm" ]; then
  pkgs="$(node -e 'const p=require("./frontend/package.json");console.log(Object.keys({...p.dependencies,...p.devDependencies}).join("\n"))')"
  printf '%s\n' "$pkgs" | grep -v '^$' | while read -r pkg; do
    dir="$nm/$pkg"
    [ -d "$dir" ] || continue
    lic="$(find_license "$dir")"
    safe="$(printf '%s' "$pkg" | sed 's#[@/]#_#g; s#^_##')"
    if [ -n "$lic" ]; then
      cp "$dir/$lic" "licenses/npm/${safe}.txt"
      echo "  ok   $pkg"
    else
      echo "  miss $pkg"
    fi
  done
else
  echo "  (frontend/node_modules missing, run: cd frontend && npm install)"
fi

echo "done."
