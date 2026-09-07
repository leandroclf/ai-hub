// Package idgen gera identificadores UUIDv7 (RFC 9562) usados como
// protocol_id, command_id, operation_id, attempt_id, event_id, etc,
// conforme EXE-16 da especificacao (docs/04_EXECUCAO_E_INTEGRACOES.md).
package idgen

import "github.com/google/uuid"

// New retorna um novo UUIDv7 como string. UUIDv7 e usado em toda
// admissao duravel (SYNC, ASYNC ou AUTO), inclusive protocolos que
// terminem em falha/expiracao. O componente temporal do UUIDv7 nunca
// deve ser tratado como prova de autorizacao ou de commit/deadline.
func New() string {
	id, err := uuid.NewV7()
	if err != nil {
		// uuid.NewV7 so falha se o gerador de aleatoriedade do SO falhar;
		// nesse caso nao ha admissao segura possivel.
		panic("idgen: falha ao gerar UUIDv7: " + err.Error())
	}
	return id.String()
}
