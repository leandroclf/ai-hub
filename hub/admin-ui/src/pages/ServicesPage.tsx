import { FormEvent, useState } from "react";
import { Service, createService, getService } from "../api/atlasClient";
import { Status, describeError } from "./shared";

const emptyForm = {
  code: "",
  version: "1",
  description: "",
  modes: "SYNC",
  client_sla_seconds: "30",
  retry_ttl_seconds: "300",
  published: false,
};

// Nao existe um GET "listar todos os servicos" no Atlas hoje (so
// GET /v1/services/{code}/{version} por chave exata) — entao esta lista e
// um historico local (nesta sessao do navegador) dos servicos que foram
// publicados ou consultados aqui, nao uma fonte de verdade persistida.
export default function ServicesPage() {
  const [form, setForm] = useState(emptyForm);
  const [lookup, setLookup] = useState({ code: "", version: "1" });
  const [services, setServices] = useState<Service[]>([]);
  const [status, setStatus] = useState<Status>(null);
  const [busy, setBusy] = useState(false);

  function upsertLocal(svc: Service) {
    setServices((prev) => {
      const withoutSame = prev.filter(
        (s) => !(s.code === svc.code && s.version === svc.version),
      );
      return [svc, ...withoutSame];
    });
  }

  async function handlePublish(e: FormEvent) {
    e.preventDefault();
    setBusy(true);
    setStatus(null);
    try {
      const svc: Service = {
        code: form.code.trim(),
        version: Number(form.version),
        description: form.description.trim(),
        modes: form.modes
          .split(",")
          .map((m) => m.trim())
          .filter(Boolean),
        client_sla_seconds: Number(form.client_sla_seconds),
        retry_ttl_seconds: Number(form.retry_ttl_seconds),
        published: form.published,
      };
      const created = await createService(svc);
      upsertLocal(created);
      setStatus({ kind: "ok", text: `Serviço ${created.code} v${created.version} publicado.` });
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
      const svc = await getService(lookup.code.trim(), Number(lookup.version));
      upsertLocal(svc);
      setStatus({ kind: "ok", text: `Serviço ${svc.code} v${svc.version} encontrado.` });
    } catch (err) {
      setStatus({ kind: "error", text: describeError(err) });
    } finally {
      setBusy(false);
    }
  }

  return (
    <>
      <section className="panel">
        <h2>Publicar novo serviço</h2>
        <p className="hint">POST /v1/services — versões publicadas são imutáveis (CAT-01).</p>
        <form className="grid" onSubmit={handlePublish}>
          <label>
            Código
            <input
              type="text"
              required
              value={form.code}
              onChange={(e) => setForm({ ...form, code: e.target.value })}
            />
          </label>
          <label>
            Versão
            <input
              type="number"
              required
              min={1}
              value={form.version}
              onChange={(e) => setForm({ ...form, version: e.target.value })}
            />
          </label>
          <label className="span-2">
            Descrição
            <input
              type="text"
              value={form.description}
              onChange={(e) => setForm({ ...form, description: e.target.value })}
            />
          </label>
          <label className="span-2">
            Modos (separados por vírgula — ex.: SYNC, ASYNC)
            <input
              type="text"
              value={form.modes}
              onChange={(e) => setForm({ ...form, modes: e.target.value })}
            />
          </label>
          <label>
            SLA do cliente (segundos)
            <input
              type="number"
              min={0}
              value={form.client_sla_seconds}
              onChange={(e) => setForm({ ...form, client_sla_seconds: e.target.value })}
            />
          </label>
          <label>
            TTL de retry (segundos)
            <input
              type="number"
              min={0}
              value={form.retry_ttl_seconds}
              onChange={(e) => setForm({ ...form, retry_ttl_seconds: e.target.value })}
            />
          </label>
          <label>
            <input
              type="checkbox"
              checked={form.published}
              onChange={(e) => setForm({ ...form, published: e.target.checked })}
            />{" "}
            Publicado
          </label>
          <button className="primary" type="submit" disabled={busy}>
            Publicar serviço
          </button>
        </form>
      </section>

      <section className="panel">
        <h2>Consultar serviço por código/versão</h2>
        <p className="hint">GET /v1/services/{"{code}"}/{"{version}"}</p>
        <form className="inline-form" onSubmit={handleLookup}>
          <label>
            Código
            <input
              type="text"
              required
              value={lookup.code}
              onChange={(e) => setLookup({ ...lookup, code: e.target.value })}
            />
          </label>
          <label>
            Versão
            <input
              type="number"
              required
              min={1}
              value={lookup.version}
              onChange={(e) => setLookup({ ...lookup, version: e.target.value })}
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
        <h2>Serviços vistos nesta sessão</h2>
        <p className="hint">
          Histórico local desta aba — o Atlas não expõe um endpoint de listagem
          geral, apenas consulta por chave exata.
        </p>
        <table>
          <thead>
            <tr>
              <th>Código</th>
              <th>Versão</th>
              <th>Descrição</th>
              <th>Modos</th>
              <th>SLA (s)</th>
              <th>Retry TTL (s)</th>
              <th>Publicado</th>
            </tr>
          </thead>
          <tbody>
            {services.length === 0 && (
              <tr>
                <td colSpan={7} className="empty-row">
                  Nenhum serviço publicado ou consultado ainda.
                </td>
              </tr>
            )}
            {services.map((s) => (
              <tr key={`${s.code}:${s.version}`}>
                <td>{s.code}</td>
                <td>{s.version}</td>
                <td>{s.description}</td>
                <td>{s.modes.join(", ")}</td>
                <td>{s.client_sla_seconds}</td>
                <td>{s.retry_ttl_seconds}</td>
                <td>{s.published ? "sim" : "não"}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </section>
    </>
  );
}
