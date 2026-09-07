import { FormEvent, useState } from "react";
import {
  ProviderAccount,
  createProviderAccount,
  getProviderAccount,
} from "../api/atlasClient";
import { Status, describeError } from "./shared";

const emptyForm = {
  provider_account_id: "",
  provider_id: "",
  environment: "sandbox",
  base_url: "",
  provider_mode: "",
  auth_type: "NONE",
  auth_username: "",
  auth_secret_ref: "",
  oauth_token_url: "",
  oauth_client_id: "",
  oauth_client_secret_ref: "",
  mtls_certificate_ref: "",
  token_ttl_seconds: "300",
};

export default function ProviderAccountsPage() {
  const [form, setForm] = useState(emptyForm);
  const [lookupId, setLookupId] = useState("");
  const [accounts, setAccounts] = useState<ProviderAccount[]>([]);
  const [status, setStatus] = useState<Status>(null);
  const [busy, setBusy] = useState(false);

  function upsertLocal(pa: ProviderAccount) {
    setAccounts((prev) => [
      pa,
      ...prev.filter((p) => p.provider_account_id !== pa.provider_account_id),
    ]);
  }

  async function handleCreate(e: FormEvent) {
    e.preventDefault();
    setBusy(true);
    setStatus(null);
    try {
      const created = await createProviderAccount({
        ...form,
        token_ttl_seconds: Number(form.token_ttl_seconds),
      });
      upsertLocal(created);
      setStatus({ kind: "ok", text: `Conta ${created.provider_account_id} cadastrada.` });
    } catch (err) {
      setStatus({ kind: "error", text: describeError(err) });
    } finally {
      setBusy(false);
    }
  }

  async function handleLookup(e: FormEvent) {
    e.preventDefault();
    setBusy(true);
    setStatus(null);
    try {
      const pa = await getProviderAccount(lookupId.trim());
      upsertLocal(pa);
      setStatus({ kind: "ok", text: `Conta ${pa.provider_account_id} encontrada.` });
    } catch (err) {
      setStatus({ kind: "error", text: describeError(err) });
    } finally {
      setBusy(false);
    }
  }

  return (
    <>
      <section className="panel">
        <h2>Cadastrar conta de provedor</h2>
        <p className="hint">POST /v1/provider-accounts (CAT-06, DAD-02)</p>
        <form className="grid" onSubmit={handleCreate}>
          <label>
            ID da conta de provedor
            <input
              type="text"
              required
              value={form.provider_account_id}
              onChange={(e) => setForm({ ...form, provider_account_id: e.target.value })}
            />
          </label>
          <label>
            ID do provedor
            <input
              type="text"
              required
              value={form.provider_id}
              onChange={(e) => setForm({ ...form, provider_id: e.target.value })}
            />
          </label>
          <label>
            Ambiente
            <input
              type="text"
              placeholder="sandbox | production"
              value={form.environment}
              onChange={(e) => setForm({ ...form, environment: e.target.value })}
            />
          </label>
          <label>
            Modo do provedor
            <input
              type="text"
              value={form.provider_mode}
              onChange={(e) => setForm({ ...form, provider_mode: e.target.value })}
            />
          </label>
          <label className="span-2">
            Base URL
            <input
              type="text"
              required
              value={form.base_url}
              onChange={(e) => setForm({ ...form, base_url: e.target.value })}
            />
          </label>
          <label>
            Autenticação
            <select value={form.auth_type} onChange={(e) => setForm({ ...form, auth_type: e.target.value })}>
              <option value="NONE">NONE</option>
              <option value="BASIC">BASIC</option>
              <option value="OAUTH_CLIENT_CREDENTIALS">OAUTH Client Credentials</option>
              <option value="MTLS_OAUTH">mTLS + OAuth</option>
            </select>
          </label>
          <label>
            TTL do token (s)
            <input type="number" min={1} value={form.token_ttl_seconds} onChange={(e) => setForm({ ...form, token_ttl_seconds: e.target.value })} />
          </label>
          <label>
            Usuário Basic
            <input value={form.auth_username} onChange={(e) => setForm({ ...form, auth_username: e.target.value })} />
          </label>
          <label>
            Referência do segredo
            <input placeholder="vault://..." value={form.auth_secret_ref} onChange={(e) => setForm({ ...form, auth_secret_ref: e.target.value })} />
          </label>
          <label className="span-2">
            URL do token OAuth
            <input placeholder="https://idp.example/oauth/token" value={form.oauth_token_url} onChange={(e) => setForm({ ...form, oauth_token_url: e.target.value })} />
          </label>
          <label>
            OAuth client ID
            <input value={form.oauth_client_id} onChange={(e) => setForm({ ...form, oauth_client_id: e.target.value })} />
          </label>
          <label>
            Referência do client secret
            <input placeholder="vault://..." value={form.oauth_client_secret_ref} onChange={(e) => setForm({ ...form, oauth_client_secret_ref: e.target.value })} />
          </label>
          <label className="span-2">
            Referência do certificado mTLS
            <input placeholder="vault://..." value={form.mtls_certificate_ref} onChange={(e) => setForm({ ...form, mtls_certificate_ref: e.target.value })} />
          </label>
          <button className="primary" type="submit" disabled={busy}>
            Cadastrar conta
          </button>
        </form>
      </section>

      <section className="panel">
        <h2>Consultar conta por ID</h2>
        <p className="hint">GET /v1/provider-accounts/{"{id}"}</p>
        <form className="inline-form" onSubmit={handleLookup}>
          <label>
            ID da conta
            <input
              type="text"
              required
              value={lookupId}
              onChange={(e) => setLookupId(e.target.value)}
            />
          </label>
          <button type="submit" disabled={busy}>
            Buscar
          </button>
        </form>
      </section>

      {status && (
        <p className={status.kind === "ok" ? "status-ok" : "status-error"}>{status.text}</p>
      )}

      <section className="panel">
        <h2>Contas vistas nesta sessão</h2>
        <p className="hint">
          Histórico local desta aba — o Atlas não expõe listagem geral de contas.
        </p>
        <table>
          <thead>
            <tr>
              <th>ID da conta</th>
              <th>Provedor</th>
              <th>Ambiente</th>
              <th>Base URL</th>
              <th>Modo</th>
              <th>Autenticação</th>
              <th>TTL (s)</th>
            </tr>
          </thead>
          <tbody>
            {accounts.length === 0 && (
              <tr>
                <td colSpan={7} className="empty-row">
                  Nenhuma conta cadastrada ou consultada ainda.
                </td>
              </tr>
            )}
            {accounts.map((a) => (
              <tr key={a.provider_account_id}>
                <td>{a.provider_account_id}</td>
                <td>{a.provider_id}</td>
                <td>{a.environment}</td>
                <td>{a.base_url}</td>
                <td>{a.provider_mode}</td>
                <td>{a.auth_type}</td>
                <td>{a.token_ttl_seconds}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </section>
    </>
  );
}
