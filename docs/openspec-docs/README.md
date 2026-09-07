# Spec-Driven Development com OpenSpec — Pacote de Prompts

Este pacote contém prompts e checklists para transformar ideias, produtos, mudanças técnicas ou evoluções de prompts em artefatos versionáveis no estilo OpenSpec.

## Arquivos

- `01-prompt-final-completo.md`  
  Prompt principal, completo, com fluxo de exploração crítica, proposal, specs, design, tasks e artefatos auxiliares.

- `02-prompt-curto.md`  
  Versão reduzida para uso rápido em agentes de IA.

- `03-template-openspec-change.md`  
  Template estrutural dos artefatos OpenSpec: `proposal.md`, specs delta, `design.md` e `tasks.md`.

- `04-checklist-avaliador.md`  
  Checklist para humano ou agente avaliador validar uma implementação contra os artefatos.

- `05-referencias-metodologia.md`  
  Referências metodológicas e justificativas de design do prompt.

## Uso recomendado

1. Cole o conteúdo de `01-prompt-final-completo.md` em um agente de IA quando a mudança for relevante, ambígua, arquitetural, brownfield ou envolver riscos de dados, APIs, segurança ou operação.
2. Cole o conteúdo de `02-prompt-curto.md` quando precisar iniciar rapidamente uma especificação.
3. Use `03-template-openspec-change.md` como estrutura de arquivos dentro do repositório.
4. Use `04-checklist-avaliador.md` depois da implementação para validar aderência.
5. Consulte `05-referencias-metodologia.md` para entender os fundamentos usados na revisão.

## Estrutura sugerida no repositório

```text
openspec/
  specs/
    <capability>/
      spec.md
  changes/
    <change-id>/
      proposal.md
      design.md
      tasks.md
      specs/
        <capability>/
          spec.md
```

## Observação importante

Este fluxo não substitui revisão humana. Ele reduz ambiguidade, força verificabilidade e melhora a qualidade das instruções para agentes de IA, mas decisões de produto, arquitetura, segurança e compliance ainda precisam de validação responsável.
