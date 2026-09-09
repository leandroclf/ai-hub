# Risk matrix

| Risco | Prioridade | Mitigação | Prova | Responsável |
|---|---|---|---|---|
| O executor recusa AdapterID diferente de synthetic-provider. Configurar endpoint não cria integração executável. Config não recebe APIKeyHeader, embora Apply o exija. | P0 | R3-EXE-01 | R3-EXE-01-S01/S02/S03 | Core e Integrações |
| Callback chama ApplyExternalObservation sem retorno de erro e responde 200; operação desconhecida, erro de consulta e observação após terminal não têm confirmação de recibo nesse caminho. A rota está sob JWT interno do Hub, não sob autenticação específica da conta provedora. | P0 | R3-EXE-02 | R3-EXE-02-S01/S02/S03 | Core e Integrações |
| Após SUBMITTING, replay pode devolver UNKNOWN durável sem recuperador geral desse estado. Envio não transmite chave idempotente homologada. Consulta inicial de replay antecede validação completa de hash/aplicação. | P0 | R3-EXE-03 | R3-EXE-03-S01/S02/S03 | Core e Integrações |
| RetryDeadline deriva do aceite e limita polling; provider_mode mantém polling/callback exclusivos. ADR R2 registra contraexemplo de commit posterior ao deadline e requisito não qualificado. | P0 | R3-EXE-04 | R3-EXE-04-S01/S02/S03 | Core e Integrações |
| Bootstraps independentes não estabelecem barreira de todas as assinaturas obrigatórias antes do relay. Quarentena financeira conserva hash/motivo sem payload para replay e então confirma a mensagem. | P0 | R3-EXE-05 | R3-EXE-05-S01/S02/S03 | Core e Integrações |

P0 bloqueia qualificação de uso; P1 bloqueia conclusão do escopo. Probabilidade quantitativa não foi medida.
