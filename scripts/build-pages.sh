#!/bin/sh
set -eu

repo_root="$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)"
cd "$repo_root"

if [ ! -f site/index.html ] || [ ! -f site/app.js ] || [ ! -f site/app.css ]; then
	printf 'site source files are incomplete\n' >&2
	exit 1
fi

mkdir -p docs
cp site/index.html site/app.js site/app.css site/favicon.svg docs/
printf '%s\n' 'pybot.0sample.com' > docs/CNAME
: > docs/.nojekyll

GOOS=js GOARCH=wasm go build -trimpath -ldflags='-s -w' -o docs/pybot.wasm ./cmd/pybotwasm

goroot="$(go env GOROOT)"
if [ -f "$goroot/lib/wasm/wasm_exec.js" ]; then
	cp "$goroot/lib/wasm/wasm_exec.js" docs/wasm_exec.js
elif [ -f "$goroot/misc/wasm/wasm_exec.js" ]; then
	cp "$goroot/misc/wasm/wasm_exec.js" docs/wasm_exec.js
else
	printf 'could not find wasm_exec.js in GOROOT %s\n' "$goroot" >&2
	exit 1
fi

chmod 644 docs/pybot.wasm docs/wasm_exec.js

printf 'GitHub Pages site built in docs/ using %s\n' "$(go version)"
