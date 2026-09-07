import { useState } from "react";
import ServicesPage from "./pages/ServicesPage";
import ProviderAccountsPage from "./pages/ProviderAccountsPage";
import CredentialBindingsPage from "./pages/CredentialBindingsPage";
import ContractsPage from "./pages/ContractsPage";

type Tab = "services" | "provider-accounts" | "credential-bindings" | "contracts";

const TABS: { id: Tab; label: string }[] = [
  { id: "services", label: "Catálogo de serviços" },
  { id: "provider-accounts", label: "Contas de provedor" },
  { id: "credential-bindings", label: "Vínculos de credencial" },
  { id: "contracts", label: "Contratos" },
];

export default function App() {
  const [tab, setTab] = useState<Tab>("services");

  return (
    <div>
      <header className="app-header">
        <h1>Atlas — Console administrativo</h1>
        <p className="subtitle">
          Plano de controle do Hub de Interoperabilidade (ARQ-04) — catálogo,
          contas de provedor, vínculos de credencial e contratos.
        </p>
      </header>
      <nav className="app-nav">
        {TABS.map((t) => (
          <button
            key={t.id}
            className={tab === t.id ? "active" : ""}
            onClick={() => setTab(t.id)}
            type="button"
          >
            {t.label}
          </button>
        ))}
      </nav>
      <main className="app-main">
        <div className="banner-warning">
          Autenticação real (SEG-01, OIDC) NÃO está implementada nesta
          interface — este é apenas um placeholder de scaffold administrativo
          local. Não exponha esta UI fora de um ambiente de desenvolvimento
          confiável sem antes resolver SEG-01. Veja o README para detalhes.
        </div>
        {tab === "services" && <ServicesPage />}
        {tab === "provider-accounts" && <ProviderAccountsPage />}
        {tab === "credential-bindings" && <CredentialBindingsPage />}
        {tab === "contracts" && <ContractsPage />}
      </main>
    </div>
  );
}
