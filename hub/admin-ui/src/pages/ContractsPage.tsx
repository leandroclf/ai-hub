import { FormEvent, useState } from "react";
import { Contract, createContract, getContract } from "../api/atlasClient";
import { Status, describeError } from "./shared";

const emptyForm = {
  tenant_id: "",
  plan: "",
  unit_price: "0",
  strict_balance: false,
  client_sla_seconds: "30",
  credential_mode_required: "SHARED_HUB",
};

export default function ContractsPage() {
  const [form, setForm] = useState(emptyForm);
  const [lookupTenant, setLookupTenant] = useState("");
  const [contracts, setContracts] = useState<Contract[]>([]);
  const [status, setStatus] = useState<Status>(null);
  const [busy, setBusy] = useState(false);

  function upsertLocal(c: Contract) {
    setContracts((prev) => [c, ...prev.filter((x) => x.tenant_id !== c.tenant_id)]);
  }

  async function handleCreate(e: FormEvent) {
    e.preventDefault();
    setBusy(true);
    setStatus(null);
    try {
      const created = await createContract({
        tenant_id: form.tenant_id.trim(),
        plan: form.plan.trim(),
        unit_price: Number(form.unit_price),
        strict_balance: form.strict_balance,
        client_sla_seconds: Number(form.client_sla_seconds),
        credential_mode_required: form.credential_mode_required,
      });
      upsertLocal(created);
      setStatus({ kind: "ok", text: `Contrato do tenant ${created.tenant_id} salvo.` });
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
      const c = await getContract(lookupTenant.trim());
      upsertLocal(c);
      setStatus({ kind: "ok", text: `Contrato do tenant ${c.tenant_id} encontrado.` });
    } catch (err) {
      setStatus({ kind: "error", text: describeError(err) });
    } finally {
      setBusy(false);
    }
  }

  return (
    <>
      <section className="panel">
        <h2>Cadastrar / atualizar contrato</h2>
        <p className="hint">POST /v1/contracts (FIN-02, CFG-04)</p>
        <form className="grid" onSubmit={handleCreate}>
          <label>
            Tenant
            <input
              type="text"
              required
              value={form.tenant_id}
              onChange={(e) => setForm({ ...form, tenant_id: e.target.value })}
            />
          </label>
          <label>
            Plano
            <input
              type="text"
              required
              value={form.plan}
              onChange={(e) => setForm({ ...form, plan: e.target.value })}
            />
          </label>
          <label>
            Preço unitário
            <input
              type="number"
              step="0.0001"
              min={0}
              value={form.unit_price}
              onChange={(e) => setForm({ ...form, unit_price: e.target.value })}
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
            Modalidade de credencial exigida
            <select
              value={form.credential_mode_required}
              onChange={(e) =>
                setForm({ ...form, credential_mode_required: e.target.value })
              }
            >
              <option value="SHARED_HUB">SHARED_HUB</option>
              <option value="TENANT_DEDICATED">TENANT_DEDICATED</option>
            </select>
          </label>
          <label>
            <input
              type="checkbox"
              checked={form.strict_balance}
              onChange={(e) => setForm({ ...form, strict_balance: e.target.checked })}
            />{" "}
            Saldo estrito (strict_balance)
          </label>
          <button className="primary" type="submit" disabled={busy}>
            Salvar contrato
          </button>
        </form>
      </section>

      <section className="panel">
        <h2>Consultar contrato por tenant</h2>
        <p className="hint">GET /v1/contracts/{"{tenantId}"}</p>
        <form className="inline-form" onSubmit={handleLookup}>
          <label>
            Tenant
            <input
              type="text"
              required
              value={lookupTenant}
              onChange={(e) => setLookupTenant(e.target.value)}
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
        <h2>Contratos vistos nesta sessão</h2>
        <table>
          <thead>
            <tr>
              <th>Tenant</th>
              <th>Plano</th>
              <th>Preço unitário</th>
              <th>SLA (s)</th>
              <th>Modalidade exigida</th>
              <th>Saldo estrito</th>
            </tr>
          </thead>
          <tbody>
            {contracts.length === 0 && (
              <tr>
                <td colSpan={6} className="empty-row">
                  Nenhum contrato cadastrado ou consultado ainda.
                </td>
              </tr>
            )}
            {contracts.map((c) => (
              <tr key={c.tenant_id}>
                <td>{c.tenant_id}</td>
                <td>{c.plan}</td>
                <td>{c.unit_price}</td>
                <td>{c.client_sla_seconds}</td>
                <td>{c.credential_mode_required}</td>
                <td>{c.strict_balance ? "sim" : "não"}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </section>
    </>
  );
}
