#!/usr/bin/env python3
"""Gera a matriz de requisitos e cenários diretamente das specs OpenSpec.

O artefato produzido é um inventário de fonte, não uma aprovação de
qualificação. Resultados de testes permanecem em matrizes de evidência
separadas para não confundir existência de cenário com PASS.
"""

from __future__ import annotations

import argparse
import csv
import hashlib
import re
import sys
from dataclasses import dataclass
from pathlib import Path


REQUIREMENT_RE = re.compile(r"^### Requirement:\s*(.+?)\s*$")
SCENARIO_RE = re.compile(r"^#### Scenario:\s*(.+?)\s*$")
ID_RE = re.compile(r"^[A-Za-z0-9]+(?:-[A-Za-z0-9]+)*")


@dataclass(frozen=True)
class Requirement:
    change: str
    requirement_id: str
    title: str
    path: Path
    line: int


@dataclass(frozen=True)
class Scenario:
    key: str
    scenario_id: str
    change: str
    requirement_id: str
    requirement_title: str
    title: str
    path: Path
    line: int


def split_identifier(label: str) -> tuple[str, str]:
    """Separa o identificador opcional do título do heading."""

    for separator in (" — ", " - "):
        if separator in label:
            candidate, title = label.split(separator, 1)
            if ID_RE.fullmatch(candidate.strip()):
                return candidate.strip(), title.strip()
    return "", label.strip()


def source_digest(paths: list[Path], root: Path) -> str:
    digest = hashlib.sha256()
    for path in paths:
        relative = path.relative_to(root).as_posix().encode()
        digest.update(len(relative).to_bytes(8, "big"))
        digest.update(relative)
        content = path.read_bytes()
        digest.update(len(content).to_bytes(8, "big"))
        digest.update(content)
    return digest.hexdigest()


def collect_specs(root: Path) -> tuple[list[Requirement], list[Scenario], str]:
    specs = sorted((root / "openspec" / "changes").glob("*/specs/*/spec.md"))
    if not specs:
        raise RuntimeError("nenhuma spec encontrada em openspec/changes/*/specs/*/spec.md")

    requirements: list[Requirement] = []
    scenarios: list[Scenario] = []
    for path in specs:
        change = path.relative_to(root / "openspec" / "changes").parts[0]
        current: Requirement | None = None
        ordinal = 0
        for line_number, line in enumerate(path.read_text(encoding="utf-8-sig").splitlines(), 1):
            requirement_match = REQUIREMENT_RE.match(line)
            if requirement_match:
                requirement_id, title = split_identifier(requirement_match.group(1))
                if not requirement_id:
                    raise RuntimeError(f"requisito sem ID: {path}:{line_number}")
                current = Requirement(change, requirement_id, title, path, line_number)
                requirements.append(current)
                ordinal = 0
                continue

            scenario_match = SCENARIO_RE.match(line)
            if not scenario_match:
                continue
            if current is None:
                raise RuntimeError(f"cenário fora de requisito: {path}:{line_number}")
            ordinal += 1
            explicit_id, title = split_identifier(scenario_match.group(1))
            scenario_id = explicit_id or f"{current.requirement_id}:S{ordinal:03d}"
            key = f"{change}:{scenario_id}"
            scenarios.append(
                Scenario(
                    key,
                    scenario_id,
                    change,
                    current.requirement_id,
                    current.title,
                    title,
                    path,
                    line_number,
                )
            )

    return requirements, scenarios, source_digest(specs, root)


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument(
        "--root",
        type=Path,
        default=Path(__file__).resolve().parents[4],
        help="raiz do repositório (padrão: descoberta a partir deste script)",
    )
    parser.add_argument(
        "--output",
        type=Path,
        default=None,
        help="CSV de saída; sem opção, imprime somente o resumo",
    )
    parser.add_argument("--expected-requirements", type=int, default=201)
    parser.add_argument("--expected-scenarios", type=int, default=732)
    args = parser.parse_args()
    root = args.root.resolve()

    try:
        requirements, scenarios, digest = collect_specs(root)
    except (OSError, RuntimeError) as error:
        print(f"INVENTORY=FAIL error={error}", file=sys.stderr)
        return 1

    requirement_keys = [(item.change, item.requirement_id) for item in requirements]
    scenario_keys = [item.key for item in scenarios]
    duplicates = sorted({key for key in scenario_keys if scenario_keys.count(key) > 1})
    if duplicates:
        print(f"INVENTORY=FAIL duplicate_scenario_keys={','.join(duplicates)}", file=sys.stderr)
        return 1
    if args.expected_requirements >= 0 and len(requirements) != args.expected_requirements:
        print(
            f"INVENTORY=FAIL requirements={len(requirements)} expected={args.expected_requirements}",
            file=sys.stderr,
        )
        return 1
    if args.expected_scenarios >= 0 and len(scenarios) != args.expected_scenarios:
        print(
            f"INVENTORY=FAIL scenarios={len(scenarios)} expected={args.expected_scenarios}",
            file=sys.stderr,
        )
        return 1

    if args.output is not None:
        output = args.output if args.output.is_absolute() else root / args.output
        output.parent.mkdir(parents=True, exist_ok=True)
        with output.open("w", encoding="utf-8", newline="") as handle:
            writer = csv.writer(handle, lineterminator="\n")
            writer.writerow(
                [
                    "scenario_key",
                    "scenario_id",
                    "change",
                    "requirement_id",
                    "requirement_title",
                    "scenario_title",
                    "spec",
                    "line",
                    "source_digest",
                ]
            )
            for item in scenarios:
                writer.writerow(
                    [
                        item.key,
                        item.scenario_id,
                        item.change,
                        item.requirement_id,
                        item.requirement_title,
                        item.title,
                        item.path.relative_to(root).as_posix(),
                        item.line,
                        digest,
                    ]
                )

    print(
        f"INVENTORY=PASS requirements={len(requirements)} scenarios={len(scenarios)} "
        f"source_digest=sha256:{digest}"
    )
    if args.output is not None:
        print(f"output={output.relative_to(root).as_posix()}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
