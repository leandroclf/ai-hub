import { FormEvent, useEffect, useState } from "react";
import { api, command, Principal } from "../api/admin";
import { Status, describeError } from "./shared";

type Destination = {
  id: string;
  version: number;
  application_id: string;
  url: string;
  state: string;
  max_attempts: number;
  timeout_seconds: number;
};

const emptyForm = {
  id: "",
  version: "1",
  application_id: "",
  url: "",
  secret_ref: "vault://webhooks/fixture",
  secret_version: "v1",
  state: "ACTIVE",
  max_attempts: "3",
  timeout_seconds: "5",
  reason: "publicação de destino versionado",
};

export default function DestinationsPage({ tenant, principal }: { tenant: string; principal: Principal }) {
  const [form, setForm] = useState(emptyForm);
  const [items, setItems] = useState<Destination[]>([]);
  const [status, setStatus] = useState<Status>(null);
  const [busy, setBusy] = useState(false);
  const canWrite = principal.scopes.includes("deliveries:write");

  async function load() {
    setBusy(true);
    try {
      const page = await api<{ items: Destination[] }>(
        `/admin/v1/destinations?tenant_id=${encodeURIComponent(tenant)}`,
      );
      setItems(page.items || []);
    } catch (error) {
      setStatus({ kind: "error", text: describeError(error) });
    } finally {
      setBusy(false);
    }
  }

  useEffect(() => {
    void load();
  }, [tenant]);

  async function publish(event: FormEvent) {
    event.preventDefault();
    setBusy(true);
    setStatus(null);
    try {
      const result = await command<Destination>(
        `/admin/v1/destinations?tenant_id=${encodeURIComponent(tenant)}`,
        {
          id: form.id.trim() || undefined,
          version: Number(form.version),
          application_id: form.application_id.trim(),
          url: form.url.trim(),
          secret_ref: form.secret_ref.trim(),
          secret_version: form.secret_version.trim(),
          state: form.state,
          max_attempts: Number(form.max_attempts),
          timeout_seconds: Number(form.timeout_seconds),
          reason: form.reason.trim(),
        },
      );
      setStatus({ kind: "ok", text: `Destino ${result.id} v${result.version} publicado.` });
      await load();
    } catch (error) {
      setStatus({ kind: "error", text: describeError(error) });
    } finally {
      setBusy(false);
    }
  }

  return (
    <>
      <section className="panel">
        <h2>Destinos de webhook versionados</h2>
        <p className="hint">
          O aceite congela URL, aplicação, referência do segredo e política de
          tentativas. A tela nunca recebe o segredo em claro.
        </p>
        {!canWrite && <p role="note">Seu perfil pode consultar destinos autorizados, mas não pode publicar novas versões.</p>}
        {canWrite && <form className="grid" onSubmit={publish}>
          <label>
            ID (vazio gera UUID)
            <input value={form.id} onChange={(e) => setForm({ ...form, id: e.target.value })} />
          </label>
          <label>
            Versão
            <input type="number" min={1} required value={form.version} onChange={(e) => setForm({ ...form, version: e.target.value })} />
          </label>
          <label>
            Aplicação (vazio = tenant)
            <input value={form.application_id} onChange={(e) => setForm({ ...form, application_id: e.target.value })} />
          </label>
          <label>
            Estado
            <select value={form.state} onChange={(e) => setForm({ ...form, state: e.target.value })}>
              <option value="ACTIVE">ACTIVE</option>
              <option value="SUSPENDED">SUSPENDED</option>
            </select>
          </label>
          <label className="span-2">
            URL autorizada
            <input type="url" required value={form.url} onChange={(e) => setForm({ ...form, url: e.target.value })} />
          </label>
          <label>
            Referência do segredo
            <input required value={form.secret_ref} onChange={(e) => setForm({ ...form, secret_ref: e.target.value })} />
          </label>
          <label>
            Versão do segredo
            <input required value={form.secret_version} onChange={(e) => setForm({ ...form, secret_version: e.target.value })} />
          </label>
          <label>
            Máximo de tentativas
            <input type="number" min={1} max={20} required value={form.max_attempts} onChange={(e) => setForm({ ...form, max_attempts: e.target.value })} />
          </label>
          <label>
            Timeout (segundos)
            <input type="number" min={1} max={15} required value={form.timeout_seconds} onChange={(e) => setForm({ ...form, timeout_seconds: e.target.value })} />
          </label>
          <label className="span-2">
            Justificativa auditável
            <input minLength={8} required value={form.reason} onChange={(e) => setForm({ ...form, reason: e.target.value })} />
          </label>
          <button className="primary" type="submit" disabled={busy}>Publicar versão</button>
        </form>}
      </section>
      {status && <p className={status.kind === "ok" ? "status-ok" : "status-error"}>{status.text}</p>}
      <section className="panel">
        <div className="page-title"><h2>Destinos persistidos</h2><button onClick={() => void load()} disabled={busy}>Atualizar</button></div>
        <div className="table-scroll" tabIndex={0}>
          <table>
            <thead><tr><th>ID</th><th>Versão</th><th>Aplicação</th><th>URL</th><th>Estado</th><th>Tentativas</th><th>Timeout</th></tr></thead>
            <tbody>{items.length === 0 ? <tr><td colSpan={7} className="empty-row">Nenhum destino versionado no tenant.</td></tr> : items.map((item) => <tr key={`${item.id}:${item.version}`}><td>{item.id}</td><td>{item.version}</td><td>{item.application_id || "tenant"}</td><td>{item.url}</td><td>{item.state}</td><td>{item.max_attempts}</td><td>{item.timeout_seconds}s</td></tr>)}</tbody>
          </table>
        </div>
      </section>
    </>
  );
}
