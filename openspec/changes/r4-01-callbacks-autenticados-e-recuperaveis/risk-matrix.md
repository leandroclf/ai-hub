# Matriz de risco

| Risco | Prioridade | Mitigação | Responsável | Evidência exigida |
|---|---|---|---|---|
| O caminho local agora vincula callback conhecido à conta durável por HMAC-SHA256 v1, atravessa a rota Kong `/callbacks/*` e não envia capability em URL. Endpoint público TLS, rotação de chaves, mTLS e autenticação de provedor comercial ainda não foram homologados; a prova local não autoriza promoção externa. | P0 | R4-CBK-01 | Integrações e Core | R4-CBK-01-S01/S02/S03 |
| O caminho legado ainda aceita capability somente junto da chave global de ingress para compatibilidade com callbacks antigos. A rota nova rejeita conta, assinatura ou timestamp ausentes/divergentes antes de acessar custódia; órfãos novos persistem conta, versão, timestamp e assinatura para reconciliação. | P0 | R4-CBK-02 | Integrações e Core | R4-CBK-02-S01/S02/S03 |
| ReconcileCallbackInbox só é chamado ao receber outro callback conhecido. A consulta percorre todos os RECEIVED associados, sem LIMIT, claim, lease ou escopo de célula. O cursor permanece aberto enquanto são feitas outras operações SQL e aplicação. O HTTP de uma operação pode depender do backlog/erro de outra, e sem novo callback não há gatilho autônomo. | P0 | R4-CBK-03 | Integrações e Core | R4-CBK-03-S01/S02/S03 |
| finalize valida OutputSchema e conserva OperationResult. ApplyExternalObservation transforma callback somente em detail, não chama a mesma validação e não confere provider_request_id contra correlação armazenada. ConserveObservation pode sobrescrever a correlação com qualquer ID não vazio recebido. Token de operação não torna o body semanticamente correto. | P0 | R4-CBK-04 | Integrações e Core | R4-CBK-04-S01/S02/S03 |

Probabilidade quantitativa não medida. Evidência estática não é incidente observado.
