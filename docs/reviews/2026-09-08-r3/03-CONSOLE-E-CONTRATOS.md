# Console administrativo e contratos de integração

O objetivo é operação ponta a ponta com dados persistidos e autorização. Não é aumentar a quantidade de menus.

| Jornada | Entrega mínima | Prova de aceite |
|---|---|---|
| Cliente/aplicação/oferta | Criar, listar, pesquisar, versionar e vincular contrato/perfil | Oferta publicada aparece elegível para token da aplicação correta, e não para outra. |
| Provedor/conta/binding | Auth condicional, referência de segredo, compartilhado ou dedicado, capacidade e saúde | Backend envia identidade correta; troca/revogação não contamina outro binding. |
| Serviço importado | Inventário, diagnóstico, adapter/capacidade e homologação | Não publica serviço inexequível; marcador real retorna ao cliente. |
| Produto | Editor de etapas/dependências, mapeamentos, parcialidade/compensação | DAG válido executa em paralelo limitado, retoma após falha e mostra timeline por etapa. |
| Perfil técnico | Contratos de entrada/saída, modalidades múltiplas e tempos efetivos | Cliente legado recebe formato contratado; GET e webhook iguais. |
| Protocolos | Filtros, paginação, detalhe, tentativa, evidência e disposição | Consulta final não acessa provedor; admin nominal entre tenants deixa trilha. |
| Entregas | Lista por delivery_id, tentativa, destino congelado e redelivery | Duas entregas do mesmo protocolo não se confundem; reenvio não refaz serviço. |
| SLA | Relógios cliente/provedor, violações, filtros, evidência | Endpoint real e painel consultam os mesmos dados autorizados. |
| Financeiro | Compra/venda, fatos, saldo, reserva, ajuste, disputa, período e exportação | Preparador/aprovador separados, valores exatos, tenant correto e fechamento auditável. |

## Correções imediatas de contrato UI/API
- OperationsPage: escolher delivery_id para entrega, não o primeiro protocol_id disponível.
- Implementar contrato efetivo da consulta SLA e da ação de reconciliação. O path sugerido pela UI não é prova de endpoint existente.
- FinancePage: remover campos de ator enviados pelo navegador; obter ator da identidade. Propagar tenant autorizado conforme API.
- Converter períodos para formato temporal aceito pelo backend com timezone/intervalo definidos (início inclusivo e fim exclusivo, se esse for o contrato publicado).
- Persistir idempotency key enquanto a mesma intenção estiver pendente; timeout ambíguo pede consulta/retry da mesma intenção, não nova chave.
- Não tratar JSON inválido em HTTP 200 como resposta válida tipada.
- Formulários devem preservar arrays de modes, campos OAuth/API Key/mTLS e versões de segredos por referência.
- Datas datetime-local precisam conversão de UTC para o fuso do editor e volta sem deslocar instantes.
- Lookups devem pesquisar/paginar no servidor e suportar mais de 100 registros.

## Ficha obrigatória por endpoint
Antes de implementar a jornada, registrar: método/path, propósito, DTO input/output, schema e versão,
escopos/papéis/MFA, tenant/application/cell, validação, status de erro, idempotência, paginação/filtro,
efeito durável, trilha de auditoria e cenários negativos. Gerar/validar OpenAPI quando aplicável.
Não inventar contratos privados divergentes entre TS e Go.

## Identidade do operador
Consumidor público consulta somente seu escopo. Operador de tenant precisa papel administrativo explícito.
Operador global de desenvolvimento é identidade nominal com MFA e leitura autorizada, auditada e mascarada.
A leitura global não concede alteração financeira, extração em massa irrestrita nem edição de segredos.
Esconder botão no frontend não é controle de acesso. Testar a API diretamente com tokens inadequados.

## Experiência e acessibilidade
Estados: carregando, vazio, erro, proibido, conflito de versão, validação, salvando e sucesso confirmado.
Preservar formulário após erro e contexto após navegação. Teclado, foco, labels e mensagens de campo fazem parte do aceite.
Browser E2E deve verificar persistência após reload e efeito posterior no runtime, além de capturas visuais.
