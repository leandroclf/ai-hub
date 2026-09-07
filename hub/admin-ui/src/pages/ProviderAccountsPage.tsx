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
      const created = await createProviderAccount({ ...form });
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
            </tr>
          </thead>
          <tbody>
            {accounts.length === 0 && (
              <tr>
                <td colSpan={5} className="empty-row">
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
              </tr>
            ))}
          </tbody>
        </table>
      </section>
    </>
  );
}
