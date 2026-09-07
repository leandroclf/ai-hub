// Cliente HTTP tipado do Atlas (plano de controle do Hub de
// Interoperabilidade — CFG-01, CFG-05, FIN-02).
//
// As interfaces abaixo espelham EXATAMENTE os structs Go expostos em
// hub/internal/atlas/store.go e as rotas registradas em
// hub/internal/atlas/handlers.go — os nomes de campo estao em snake_case
// porque o backend Go usa essas tags `json:"..."` e nao faz nenhuma
// transformacao de case. Nao invente campos nem rotas aqui: qualquer
// mudanca de contrato deve ser refletida primeiro no Go e so entao aqui.
//
// Todas as chamadas usam o prefixo relativo "/api" (nunca uma URL
// absoluta tipo "http://localhost:8081"), para funcionar tanto via proxy
// do Vite em dev (ver vite.config.ts) quanto atras de um gateway comum
// (ex.: Kong) em outros ambientes.

const BASE = "/api";

export class AtlasApiError extends Error {
  status: number;
  code: string;

  constructor(status: number, code: string, message: string) {
    super(message);
    this.name = "AtlasApiError";
    this.status = status;
    this.code = code;
  }
}

interface ErrorBody {
  error: string;
  message: string;
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(`${BASE}${path}`, {
    headers: { "Content-Type": "application/json" },
    ...init,
  });

  if (!res.ok) {
    let body: ErrorBody | undefined;
    try {
      body = (await res.json()) as ErrorBody;
    } catch {
      // corpo de erro nao era JSON; segue com status bruto abaixo.
    }
    throw new AtlasApiError(
      res.status,
      body?.error ?? "unknown_error",
      body?.message ?? res.statusText,
    );
  }

  // Respostas 204 (nao usadas hoje pelo Atlas, mas defensivo) nao tem corpo.
  if (res.status === 204) {
    return undefined as T;
  }
  return (await res.json()) as T;
}

function postJSON<T>(path: string, body: unknown): Promise<T> {
  return request<T>(path, { method: "POST", body: JSON.stringify(body) });
}

function getJSON<T>(path: string): Promise<T> {
  return request<T>(path, { method: "GET" });
}

// ---------------------------------------------------------------------
// Servicos do catalogo (CAT-01) — POST /v1/services, GET /v1/services/{code}/{version}
// ---------------------------------------------------------------------

export interface Service {
  code: string;
  version: number;
  description: string;
  modes: string[];
  client_sla_seconds: number;
  retry_ttl_seconds: number;
  published: boolean;
}

export function createService(svc: Service): Promise<Service> {
  return postJSON<Service>("/v1/services", svc);
}

export function getService(code: string, version: number): Promise<Service> {
  return getJSON<Service>(
    `/v1/services/${encodeURIComponent(code)}/${encodeURIComponent(String(version))}`,
  );
}

// ---------------------------------------------------------------------
// Contas de provedor (CAT-06, DAD-02) — POST /v1/provider-accounts,
// GET /v1/provider-accounts/{id}
// ---------------------------------------------------------------------

export interface ProviderAccount {
  provider_account_id: string;
  provider_id: string;
  environment: string;
  base_url: string;
  provider_mode: string;
  auth_type: "NONE" | "BASIC" | "OAUTH_CLIENT_CREDENTIALS" | "MTLS_OAUTH" | string;
  auth_username: string;
  auth_secret_ref: string;
  oauth_token_url: string;
  oauth_client_id: string;
  oauth_client_secret_ref: string;
  mtls_certificate_ref: string;
  token_ttl_seconds: number;
}

export function createProviderAccount(pa: ProviderAccount): Promise<ProviderAccount> {
  return postJSON<ProviderAccount>("/v1/provider-accounts", pa);
}

export function getProviderAccount(id: string): Promise<ProviderAccount> {
  return getJSON<ProviderAccount>(`/v1/provider-accounts/${encodeURIComponent(id)}`);
}

// ---------------------------------------------------------------------
// Vinculos de credencial (CFG-05, DAD-02) — POST /v1/credential-bindings,
// GET /v1/credentials/resolve?tenant_id=&provider_account_id=
//
// IMPORTANTE (CFG-05): o campo `secret_ref` e apenas uma REFERENCIA ao
// cofre (ex.: caminho/alias de secret), nunca o segredo em si. Esta
// interface administrativa NUNCA tem um campo para digitar/visualizar o
// segredo em claro — o segredo e escrito no cofre por um fluxo restrito
// e separado, sem leitura posterior em claro pela interface. Se algum dia
// alguem for adicionar um campo "secret"/"password" nesta tela, pare: isso
// violaria CFG-05.
// ---------------------------------------------------------------------

export interface CredentialBinding {
  binding_id: string;
  credential_mode: "SHARED_HUB" | "TENANT_DEDICATED" | string;
  tenant_id: string; // vazio quando credential_mode = SHARED_HUB
  provider_account_id: string;
  secret_ref: string; // referencia ao cofre — nunca o segredo em si (CFG-05)
  settlement_party: string;
  state: string; // ex.: "ATIVO"; o backend aplica esse default se vazio
}

export function createCredentialBinding(
  cb: CredentialBinding,
): Promise<CredentialBinding> {
  return postJSON<CredentialBinding>("/v1/credential-bindings", cb);
}

export function resolveCredential(
  tenantId: string,
  providerAccountId: string,
): Promise<CredentialBinding> {
  const params = new URLSearchParams({
    tenant_id: tenantId,
    provider_account_id: providerAccountId,
  });
  return getJSON<CredentialBinding>(`/v1/credentials/resolve?${params.toString()}`);
}

// ---------------------------------------------------------------------
// Contratos (FIN-02, CFG-04) — POST /v1/contracts, GET /v1/contracts/{tenantId}
// ---------------------------------------------------------------------

export interface Contract {
  tenant_id: string;
  plan: string;
  unit_price: number;
  strict_balance: boolean;
  client_sla_seconds: number;
  credential_mode_required: string; // default do backend: "SHARED_HUB" se vazio
}

export function createContract(c: Contract): Promise<Contract> {
  return postJSON<Contract>("/v1/contracts", c);
}

export function getContract(tenantId: string): Promise<Contract> {
  return getJSON<Contract>(`/v1/contracts/${encodeURIComponent(tenantId)}`);
}
