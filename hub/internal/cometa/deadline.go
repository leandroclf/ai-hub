package cometa

import (
	"context"
	"time"
)

// withDeadline propaga o deadline absoluto do passo (EXE-14: "deadline
// propagado") ao contexto da chamada ao provedor, para que o
// cancelamento local ocorra no limite correto sem depender de timeout
// fixo por tentativa.
func withDeadline(ctx context.Context, deadline time.Time) (context.Context, func()) {
	return context.WithDeadline(ctx, deadline)
}
