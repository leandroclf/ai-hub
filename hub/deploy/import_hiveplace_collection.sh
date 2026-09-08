#!/bin/sh
set -eu
# R2: only stage a collection in Atlas. No SQL deletion or implicit activation.
SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
COLLECTION=${1:-"$SCRIPT_DIR/../../postman/collections/Hive.postman_collection.json"}
ATLAS_URL=${ATLAS_URL:-"http://localhost:8081"}
: "${ATLAS_TOKEN_FILE:?Informe ATLAS_TOKEN_FILE com token administrativo de ensaio}"
test -r "$COLLECTION"
test -r "$ATLAS_TOKEN_FILE"
command -v jq >/dev/null
WORK_DIR=$(mktemp -d)
trap 'rm -rf "$WORK_DIR"' EXIT HUP INT TERM
chmod 700 "$WORK_DIR"
jq -n --slurpfile collection "$COLLECTION" '{collection:$collection[0]}' > "$WORK_DIR/request.json"
printf 'Authorization: Bearer %s\n' "$(cat "$ATLAS_TOKEN_FILE")" > "$WORK_DIR/headers"
curl --fail-with-body -sS -X POST "$ATLAS_URL/admin/v1/imports" -H @"$WORK_DIR/headers" -H 'Content-Type: application/json' --data-binary @"$WORK_DIR/request.json"
