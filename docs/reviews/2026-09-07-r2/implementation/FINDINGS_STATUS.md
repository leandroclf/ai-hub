# Evolução dos achados R2

Registro em andamento. Auditoria original preservada; nenhum achado integralmente encerrado por inferência de testes parciais.

| Achado | Título | Situação |
|---|---|---|
| F-01 | Identidade do cliente controlada pelo próprio chamador | Revalidação e qualificação integrada pendentes |
| F-02 | APIs internas e administrativas sem identidade de workload | Revalidação e qualificação integrada pendentes |
| F-03 | Callback confirma recebimento sem custódia comprovada | Revalidação e qualificação integrada pendentes |
| F-04 | 202 após falha de envio de comando, sem retomada de despacho | Revalidação e qualificação integrada pendentes |
| F-05 | Consumidores removem mensagens mesmo quando o efeito falha | Revalidação e qualificação integrada pendentes |
| F-06 | Resultado retornado não é o resultado do provedor | Revalidação e qualificação integrada pendentes |
| F-07 | Concorrência pode disparar mais de uma operação externa | Revalidação e qualificação integrada pendentes |
| F-08 | Estado externo e fato são gravados em transações separadas | Revalidação e qualificação integrada pendentes |
| F-09 | Corrida temporal permite sucesso depois do deadline | Revalidação e qualificação integrada pendentes |
| F-10 | TTL de retry configurado não governa execução | Revalidação e qualificação integrada pendentes |
| F-11 | SYNC, AUTO e UUID em erros não cumprem todo o contrato | Revalidação e qualificação integrada pendentes |
| F-12 | Credencial dedicada resolvida não é a usada na chamada | Revalidação e qualificação integrada pendentes |
| F-13 | Autenticação do simulador não equivale a integração segura real | Revalidação e qualificação integrada pendentes |
| F-14 | Redis e broker bloqueiam inclusive reinício do caminho SYNC | Revalidação e qualificação integrada pendentes |
| F-15 | Polling sem autenticação, lease e orçamento configurável | Revalidação e qualificação integrada pendentes |
| F-16 | Não há amortecimento adaptativo nem isolamento de carga | Revalidação e qualificação integrada pendentes |
| F-17 | Catálogo publicado pode ser sobrescrito e não governa admissão | Revalidação e qualificação integrada pendentes |
| F-18 | Produtos, DAG e contratos legados ainda não existem | Revalidação e qualificação integrada pendentes |
| F-19 | Importação não ativa adaptadores e substitui configurações locais | Revalidação e qualificação integrada pendentes |
| F-20 | Console limitado a quatro formulários e histórico efêmero | Revalidação e qualificação integrada pendentes |
| F-21 | Faltam jornadas administrativas de produto, operação e financeiro | Revalidação e qualificação integrada pendentes |
| F-22 | Formulários expõem detalhes técnicos e estados pouco guiados | Revalidação e qualificação integrada pendentes |
| F-23 | Custo e receita não usam contratos econômicos congelados | Revalidação e qualificação integrada pendentes |
| F-24 | Saldo estrito não contabiliza consumo já capturado | Revalidação e qualificação integrada pendentes |
| F-25 | Ledger, precisão monetária e fechamento não estão implementados | Revalidação e qualificação integrada pendentes |
| F-26 | Webhook usa URL como identidade do segredo | Revalidação e qualificação integrada pendentes |
| F-27 | Webhook sem claim, recibos completos e política por cliente | Revalidação e qualificação integrada pendentes |
| F-28 | Resultado final não é materializado como representação imutável completa | Revalidação e qualificação integrada pendentes |
| F-29 | Arquivos grandes não participam dos fluxos reais | Revalidação e qualificação integrada pendentes |
| F-30 | Persistência sem isolamento de papéis, expurgo e auditoria durável | Revalidação e qualificação integrada pendentes |
| F-31 | Docker local não inclui toda a plataforma e não garante persistência na recriação | Revalidação e qualificação integrada pendentes |
| F-32 | Kubernetes é referência incompleta, sem os cinco ambientes | Revalidação e qualificação integrada pendentes |
| F-33 | Autoscaling de células e reconciliação de IaC não fecham o circuito | Revalidação e qualificação integrada pendentes |
| F-34 | Clientes AWS e bootstrap são fixos ao ambiente local | Revalidação e qualificação integrada pendentes |
| F-35 | Probes não acompanham capacidades e não há drenagem graciosa | Revalidação e qualificação integrada pendentes |
| F-36 | Observabilidade não mede SLA, pressão nem custódia | Revalidação e qualificação integrada pendentes |
| F-37 | Ensaios de HA, isolamento e recuperação não foram demonstrados | Revalidação e qualificação integrada pendentes |
| F-38 | Cobertura existente não prova as garantias que os nomes dos testes sugerem | Revalidação e qualificação integrada pendentes |
| F-39 | Rastreabilidade histórica contém conclusões que não refletem o snapshot | Revalidação e qualificação integrada pendentes |
| F-40 | Versões e contratos de ferramenta carecem de qualificação operacional | Revalidação e qualificação integrada pendentes |
| F-41 | Destinos configuráveis não possuem defesa SSRF | Revalidação e qualificação integrada pendentes |
| F-42 | Recepção e pools não limitam custo por tenant | Revalidação e qualificação integrada pendentes |
