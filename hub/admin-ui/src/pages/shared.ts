import { AtlasApiError } from "../api/atlasClient";

export function describeError(err: unknown): string {
  if (err instanceof AtlasApiError) {
    return `Erro (${err.status} ${err.code}): ${err.message}`;
  }
  if (err instanceof Error) {
    return `Erro: ${err.message}`;
  }
  return "Erro desconhecido.";
}

export type Status = { kind: "ok" | "error"; text: string } | null;
