# Referências verificadas em 2026-09-08

A conclusão sobre o código deriva do SHA fixado e das fontes por achado, não de documentação genérica.
As referências abaixo sustentam somente decisões técnicas correspondentes; não qualificam o Hub.

- A entrega SNS possui política finita e precisa de tratamento de mensagens não entregues: [AWS SNS delivery retries](https://docs.aws.amazon.com/sns/latest/dg/sns-message-delivery-retries.html).
- RLS deve ser qualificada com o papel efetivo, observando privilégios que a contornam: [PostgreSQL row security](https://www.postgresql.org/docs/current/ddl-rowsecurity.html).
- HPA controla réplicas; não é por si só estratégia completa de nós/dados: [Kubernetes HPA](https://kubernetes.io/docs/concepts/workloads/autoscaling/horizontal-pod-autoscale/).
- Decoder.UseNumber evita conversão automática para float64; precisão requer também validação e aritmética adequadas: [Go encoding/json](https://pkg.go.dev/encoding/json#Decoder.UseNumber).
