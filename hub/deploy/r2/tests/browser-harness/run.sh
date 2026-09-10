#!/usr/bin/env bash
set -euo pipefail

if ! command -v browser-harness >/dev/null 2>&1; then
  echo 'browser-harness não está instalado; consulte README.md para instalação opcional.' >&2
  exit 2
fi

if ! curl -fsS "${R2_ADMIN_URL:-http://localhost:13000}/services" >/dev/null; then
  echo "console indisponível em ${R2_ADMIN_URL:-http://localhost:13000}" >&2
  exit 3
fi

browser-harness < "$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)/scenarios/admin-console.py"
