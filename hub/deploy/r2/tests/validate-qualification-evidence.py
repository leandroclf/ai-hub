#!/usr/bin/env python3
"""Valida um manifesto de evidência antes de promover uma qualificação.

O manifesto é deliberadamente pequeno e explícito: um resultado positivo só
é aceito quando a evidência registra o que era esperado e o que foi observado,
além de comando, ambiente, SHA e um artefato existente que identifica o
cenário. Assim, um HTTP 200 isolado não pode ser promovido como prova integral.
"""

from __future__ import annotations

import argparse
import json
import re
import sys
from pathlib import Path
from typing import Any


QUALIFIED_PREFIXES = ("PASS", "IMPLEMENTADO")
SOURCE_SHA = re.compile(r"^[0-9a-f]{40}$")


def fail(message: str) -> int:
    print(f"QUALIFICATION_EVIDENCE=BLOCK error={message}", file=sys.stderr)
    return 1


def read_manifest(path: Path) -> list[dict[str, Any]]:
    try:
        value = json.loads(path.read_text(encoding="utf-8"))
    except (OSError, json.JSONDecodeError) as error:
        raise ValueError(f"manifesto inválido: {error}") from error
    entries = value if isinstance(value, list) else [value]
    if not entries or not all(isinstance(entry, dict) for entry in entries):
        raise ValueError("manifesto deve ser um objeto ou uma lista de objetos")
    return entries


def resolve_evidence(root: Path, reference: str) -> Path:
    path = Path(reference)
    candidate = (path if path.is_absolute() else root / path).resolve()
    try:
        candidate.relative_to(root)
    except ValueError as error:
        raise ValueError(f"evidência fora da raiz do projeto: {reference}") from error
    if not candidate.is_file():
        raise ValueError(f"artefato de evidência inexistente: {reference}")
    return candidate


def validate_entry(entry: dict[str, Any], root: Path, require_pass: bool) -> None:
    required = {"scenario_id", "status", "expected", "observed", "command", "environment", "source_sha", "evidence"}
    missing = sorted(required - entry.keys())
    if missing:
        raise ValueError(f"campos obrigatórios ausentes: {','.join(missing)}")

    scenario_id = entry["scenario_id"]
    status = entry["status"]
    expected = entry["expected"]
    observed = entry["observed"]
    if not isinstance(scenario_id, str) or not scenario_id.strip():
        raise ValueError("scenario_id vazio")
    if not isinstance(status, str) or not status.strip():
        raise ValueError(f"{scenario_id}: status vazio")
    if not isinstance(expected, list) or not expected or not all(isinstance(item, str) and item.strip() for item in expected):
        raise ValueError(f"{scenario_id}: expected deve conter oráculos não vazios")
    if not isinstance(observed, list) or not all(isinstance(item, str) and item.strip() for item in observed):
        raise ValueError(f"{scenario_id}: observed inválido")
    missing_oracles = sorted(set(expected) - set(observed))
    qualified = status.startswith(QUALIFIED_PREFIXES)
    if qualified and missing_oracles:
        raise ValueError(f"{scenario_id}: evidência incompatível; oráculos ausentes={','.join(missing_oracles)}")
    if require_pass and not qualified:
        raise ValueError(f"{scenario_id}: status não qualificável para promoção={status}")

    for field in ("command", "environment", "evidence"):
        if not isinstance(entry[field], str) or not entry[field].strip():
            raise ValueError(f"{scenario_id}: {field} vazio")
    source_sha = entry["source_sha"]
    if not isinstance(source_sha, str) or not SOURCE_SHA.fullmatch(source_sha):
        raise ValueError(f"{scenario_id}: source_sha deve ser um SHA-1 hexadecimal de 40 caracteres")

    artifact = resolve_evidence(root, entry["evidence"])
    contents = artifact.read_text(encoding="utf-8", errors="replace")
    if scenario_id not in contents:
        raise ValueError(f"{scenario_id}: artefato não identifica o cenário")


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--manifest", required=True, type=Path)
    parser.add_argument("--root", default=Path.cwd(), type=Path)
    parser.add_argument("--require-pass", action="store_true")
    args = parser.parse_args()
    root = args.root.resolve()
    try:
        entries = read_manifest(args.manifest.resolve())
        seen: set[str] = set()
        for entry in entries:
            scenario_id = entry.get("scenario_id")
            if scenario_id in seen:
                raise ValueError(f"cenário duplicado: {scenario_id}")
            seen.add(scenario_id)
            validate_entry(entry, root, args.require_pass)
    except ValueError as error:
        return fail(str(error))
    print(f"QUALIFICATION_EVIDENCE=PASS scenarios={len(entries)}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
