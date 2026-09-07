#!/bin/sh
set -eu

# O catálogo não é mais preenchido com fixtures de teste. Para importar a
# collection HivePlace local, execute no host:
#   hub/deploy/import_hiveplace_collection.sh
#
# Manter este job como no-op evita que uma subida/recriação da stack reintroduza
# APIs de teste depois da limpeza solicitada.
echo "seed de fixtures desabilitado; use deploy/import_hiveplace_collection.sh para o catálogo HivePlace"
