import { FormEvent, useState } from "react";
import {
  CredentialBinding,
  createCredentialBinding,
  resolveCredential,
} from "../api/atlasClient";
import { Status, describeError } from "./shared";

const emptyForm = {
  binding_id: "",
  credential_mode: "SHARED_HUB",
  tenant_id: "",
  provider_account_id: "",
  secret_ref: "",
  settlement_party: "",
  state: "ATIVO",
};

const emptyResolveForm = { tenant_id: "", provider_account_id: "" };

export default function CredentialBindingsPage() {
  const [form, setForm] = useState(emptyForm);
  const [resolveForm, setResolveForm] = useState(emptyResolveForm);
  const [bindings, setBindings] = useState<CredentialBinding[]>([]);
  const [resolved, setResolved] = useState<CredentialBinding | null>(null);
  const [status, setStatus] = useState<Status>(null);
  const [busy, setBusy] = useState(false);

  const isDedicated = form.credential_mode === "TENANT_DEDICATED";

  async function handleCreate(e: FormEvent) {
    e.preventDefault();
    setBusy(true);
    setStatus(null);
    try {
      const created = await createCredentialBinding({
        ...form,
        // SHARED_HUB nao tem tenant_id (o backend grava NULL nesse caso).
        tenant_id: isDedicated ? form.tenant_id.trim() : "",
      });
      setBindings((prev) => [
        created,
        ...prev.filter((b) => b.binding_id !== created.binding_id),
      ]);
      setStatus({ kind: "ok", text: `Vínculo ${created.binding_id} cadastrado.` });
    } catch (err) {
      setStatus({ kind: "error", text: describeError(err) });
    } finally {
      setBusy(false);
    }
  }

  async function handleResolve(e: FormEvent) {
    e.preventDefault();
    setBusy(true);
    setStatus(null);
    setResolved(null);
    try {
      const cb = await resolveCredential(
        resolveForm.tenant_id.trim(),
        resolveForm.provider_account_id.trim(),
      );
      setResolved(cb);
      setStatus({ kind: "ok", text: `Vínculo resolvido: ${cb.binding_id}.` });
    } catch (err) {
      setStatus({ kind: "error", text: describeError(err) });
    } finally {
      setBusy(false);
    }
  }

  return (
    <>
      <section className="panel">
        <h2>Cadastrar vínculo de credencial</h2>
        <p className="hint">POST /v1/credential-bindings (CFG-05, DAD-02)</p>
        <form className="grid" onSubmit={handleCreate}>
          <label>
            ID do vínculo
            <input
              type="text"
              required
              value={form.binding_id}
              onChange={(e) => setForm({ ...form, binding_id: e.target.value })}
            />
          </label>
          <label>
            Modalidade de credencial
            <select
              value={form.credential_mode}
              onChange={(e) => setForm({ ...form, credential_mode: e.target.value })}
            >
              <option value="SHARED_HUB">SHARED_HUB</option>
              <option value="TENANT_DEDICATED">TENANT_DEDICATED</option>
            </select>
          </label>
          <label>
            Tenant (só p/ TENANT_DEDICATED)
            <input
              type="text"
              disabled={!isDedicated}
              value={isDedicated ? form.tenant_id : ""}
              onChange={(e) => setForm({ ...form, tenant_id: e.target.value })}
            />
          </label>
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
            Parte de liquidação (settlement_party)
            <input
              type="text"
              value={form.settlement_party}
              onChange={(e) => setForm({ ...form, settlement_party: e.target.value })}
            />
          </label>
          <label>
            Estado
            <input
              type="text"
              placeholder="ATIVO (default do backend se vazio)"
              value={form.state}
              onChange={(e) => setForm({ ...form, state: e.target.value })}
            />
          </label>
          <label className="span-2">
            Referência do segredo no cofre (secret_ref)
            <input
              type="text"
              required
              placeholder="ex.: vault://provider/foo/secret-123"
              value={form.secret_ref}
              onChange={(e) => setForm({ ...form, secret_ref: e.target.value })}
            />
          </label>
          {/*
            CFG-05: esta tela NUNCA tem um campo para o segredo em si (ex.:
            client_secret, api_key em claro). O campo acima e apenas a
            REFERENCIA/alias/caminho do segredo no cofre — o segredo real e
            escrito no cofre por um fluxo restrito e separado, e nao existe
            leitura posterior em claro por esta interface administrativa.
            Nao adicione aqui um campo de segredo em claro.
          */}
          <p className="secret-note span-2">
            O segredo em si nunca é digitado ou exibido nesta tela (CFG-05).
            Apenas a referência do cofre (secret_ref) é cadastrada aqui; a
            escrita do segredo real acontece em um fluxo restrito separado.
          </p>
          <button className="primary" type="submit" disabled={busy}>
            Cadastrar vínculo
          </button>
        </form>
      </section>

      <section className="panel">
        <h2>Resolver credencial (SEG-05)</h2>
        <p className="hint">
          GET /v1/credentials/resolve?tenant_id=&amp;provider_account_id= — resolução
          determinística pela modalidade exigida pelo contrato do tenant, sem
          fallback implícito entre modos.
        </p>
        <form className="inline-form" onSubmit={handleResolve}>
          <label>
            Tenant
            <input
              type="text"
              required
              value={resolveForm.tenant_id}
              onChange={(e) =>
                setResolveForm({ ...resolveForm, tenant_id: e.target.value })
              }
            />
          </label>
          <label>
            ID da conta de provedor
            <input
              type="text"
              required
              value={resolveForm.provider_account_id}
              onChange={(e) =>
                setResolveForm({ ...resolveForm, provider_account_id: e.target.value })
              }
            />
          </label>
          <button type="submit" disabled={busy}>
            Resolver
          </button>
        </form>
        {resolved && (
          <table>
            <tbody>
              <tr>
                <th>Binding ID</th>
                <td>{resolved.binding_id}</td>
              </tr>
              <tr>
                <th>Modalidade</th>
                <td>{resolved.credential_mode}</td>
              </tr>
              <tr>
                <th>Tenant</th>
                <td>{resolved.tenant_id || "(SHARED_HUB)"}</td>
              </tr>
              <tr>
                <th>Conta de provedor</th>
                <td>{resolved.provider_account_id}</td>
              </tr>
              <tr>
                <th>secret_ref (referência, não o segredo)</th>
                <td>{resolved.secret_ref}</td>
              </tr>
              <tr>
                <th>Settlement party</th>
                <td>{resolved.settlement_party}</td>
              </tr>
              <tr>
                <th>Estado</th>
                <td>{resolved.state}</td>
              </tr>
            </tbody>
          </table>
        )}
      </section>

      {status && (
        <p className={status.kind === "ok" ? "status-ok" : "status-error"}>{status.text}</p>
      )}

      <section className="panel">
        <h2>Vínculos cadastrados nesta sessão</h2>
        <table>
          <thead>
            <tr>
              <th>Binding ID</th>
              <th>Modalidade</th>
              <th>Tenant</th>
              <th>Conta</th>
              <th>Settlement</th>
              <th>Estado</th>
            </tr>
          </thead>
          <tbody>
            {bindings.length === 0 && (
              <tr>
                <td colSpan={6} className="empty-row">
                  Nenhum vínculo cadastrado ainda.
                </td>
              </tr>
            )}
            {bindings.map((b) => (
              <tr key={b.binding_id}>
                <td>{b.binding_id}</td>
                <td>{b.credential_mode}</td>
                <td>{b.tenant_id || "(SHARED_HUB)"}</td>
                <td>{b.provider_account_id}</td>
                <td>{b.settlement_party}</td>
                <td>{b.state}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </section>
    </>
  );
}
