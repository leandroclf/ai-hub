#!/usr/bin/env python3
"""Gera uma matriz completa de resultados sem promover cenários sem prova.

O inventário de fonte define o universo de cenários. A matriz de evidências
existente pode cobrir apenas uma fatia; todos os demais cenários recebem um
estado explícito de não qualificação nesta rodada. Isso torna lacunas
observáveis e impede que um relatório parcial pareça uma aprovação integral.
"""

from __future__ import annotations

import argparse
import csv
import sys
from pathlib import Path


DEFAULT_INVENTORY = Path("docs/reviews/2026-09-09-r4/implementation/INVENTORY-732-CENARIOS.csv")
DEFAULT_RESULTS = Path("docs/reviews/2026-09-09-r4/implementation/SCENARIO_RESULTS.csv")
DEFAULT_OUTPUT = Path("docs/reviews/2026-09-09-r4/implementation/RESULT-MATRIX-732-CENARIOS.csv")
UNQUALIFIED = "NAO_QUALIFICADO_NESTA_RODADA"
UNQUALIFIED_EVIDENCE = "sem execução/evidência vinculada nesta rodada"


def display_path(path: Path, root: Path) -> str:
    try:
        return path.relative_to(root).as_posix()
    except ValueError:
        return path.as_posix()


def read_inventory(path: Path) -> list[dict[str, str]]:
    with path.open(encoding="utf-8-sig", newline="") as handle:
        rows = list(csv.DictReader(handle))
    required = {"scenario_key", "scenario_id", "change", "requirement_id", "source_digest"}
    if not rows or not required.issubset(rows[0]):
        raise ValueError(f"inventário inválido ou incompleto: {path}")
    keys = [row["scenario_key"] for row in rows]
    duplicates = sorted({key for key in keys if keys.count(key) > 1})
    if duplicates:
        raise ValueError(f"chaves duplicadas no inventário: {','.join(duplicates)}")
    return rows


def read_results(path: Path) -> dict[str, tuple[str, str]]:
    with path.open(encoding="utf-8-sig", newline="") as handle:
        reader = csv.DictReader(handle, delimiter=";")
        if not reader.fieldnames or not {"cenario", "resultado", "evidencia"}.issubset(reader.fieldnames):
            raise ValueError(f"matriz de resultados inválida ou incompleta: {path}")
        output: dict[str, tuple[str, str]] = {}
        for row in reader:
            scenario_id = row["cenario"].strip()
            if not scenario_id:
                raise ValueError(f"cenário vazio na matriz: {path}")
            if scenario_id in output:
                raise ValueError(f"cenário duplicado na matriz: {scenario_id}")
            output[scenario_id] = (row["resultado"].strip(), row["evidencia"].strip())
    return output


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--root", type=Path, default=Path(__file__).resolve().parents[4])
    parser.add_argument("--inventory", type=Path, default=DEFAULT_INVENTORY)
    parser.add_argument("--results", type=Path, default=DEFAULT_RESULTS)
    parser.add_argument("--output", type=Path, default=DEFAULT_OUTPUT)
    parser.add_argument("--expected-scenarios", type=int, default=732)
    args = parser.parse_args()
    root = args.root.resolve()

    def resolve(path: Path) -> Path:
        return path if path.is_absolute() else root / path

    try:
        inventory = read_inventory(resolve(args.inventory))
        results = read_results(resolve(args.results))
        known_ids = {row["scenario_id"] for row in inventory}
        unknown = sorted(set(results) - known_ids)
        if unknown:
            raise ValueError(f"resultados sem cenário no inventário: {','.join(unknown)}")
        if args.expected_scenarios >= 0 and len(inventory) != args.expected_scenarios:
            raise ValueError(f"cenários={len(inventory)} esperado={args.expected_scenarios}")

        output = resolve(args.output)
        output.parent.mkdir(parents=True, exist_ok=True)
        fields = [
            "scenario_key",
            "scenario_id",
            "change",
            "requirement_id",
            "requirement_title",
            "scenario_title",
            "status",
            "evidence",
            "spec",
            "line",
            "source_digest",
        ]
        with output.open("w", encoding="utf-8", newline="") as handle:
            writer = csv.DictWriter(handle, fieldnames=fields, lineterminator="\n")
            writer.writeheader()
            for row in inventory:
                status, evidence = results.get(row["scenario_id"], (UNQUALIFIED, UNQUALIFIED_EVIDENCE))
                writer.writerow(
                    {
                        **{field: row.get(field, "") for field in fields},
                        "status": status,
                        "evidence": evidence,
                    }
                )
    except (OSError, ValueError) as error:
        print(f"RESULT_MATRIX=FAIL error={error}", file=sys.stderr)
        return 1

    qualified = sum(row["scenario_id"] in results for row in inventory)
    print(
        f"RESULT_MATRIX=PASS scenarios={len(inventory)} qualified_rows={qualified} "
        f"unqualified_rows={len(inventory) - qualified} source_digest=sha256:{inventory[0]['source_digest']}"
    )
    print(f"output={display_path(output, root)}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
