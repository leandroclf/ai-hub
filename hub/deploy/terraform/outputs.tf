# Saídas de referência. Nenhum destes valores existe de fato — só têm
# significado depois de um `terraform apply` real, que não foi (e não
# pode ser) executado nesta revisão (P-01 pendente).

output "vpc_id" {
  description = "ID da VPC de referência"
  value       = aws_vpc.this.id
}

output "availability_zones" {
  description = "Zonas de disponibilidade usadas (multi-AZ, OPE-04)"
  value       = local.azs
}

output "public_subnet_ids" {
  value = aws_subnet.public[*].id
}

output "private_subnet_ids" {
  value = aws_subnet.private[*].id
}

output "eks_cluster_name" {
  value = aws_eks_cluster.this.name
}

output "eks_cluster_endpoint" {
  value = aws_eks_cluster.this.endpoint
}

output "eks_cluster_oidc_issuer_url" {
  value = aws_eks_cluster.this.identity[0].oidc[0].issuer
}

output "eks_oidc_provider_arn" {
  value = aws_iam_openid_connect_provider.eks.arn
}

output "karpenter_node_instance_profile_name" {
  value = aws_iam_instance_profile.karpenter_node.name
}

output "karpenter_controller_role_arn" {
  value = aws_iam_role.karpenter_controller.arn
}

output "keda_operator_role_arn" {
  value = aws_iam_role.keda_operator.arn
}

output "crossplane_provider_aws_role_arn" {
  value = aws_iam_role.crossplane_provider_aws.arn
}

output "db_control_endpoint" {
  description = "Endpoint do PostgreSQL hub_control (Atlas — global, fora de célula)"
  value       = aws_db_instance.control.endpoint
}

output "db_control_secret_arn" {
  value = aws_secretsmanager_secret.control_db.arn
}

output "cell_core_db_endpoints" {
  description = "Endpoint do PostgreSQL hub_core por célula"
  value       = { for k, v in aws_db_instance.cell_core : k => v.endpoint }
}

output "cell_finance_db_endpoints" {
  description = "Endpoint do PostgreSQL hub_finance por célula"
  value       = { for k, v in aws_db_instance.cell_finance : k => v.endpoint }
}

output "cell_core_db_secret_arns" {
  value = { for k, v in aws_secretsmanager_secret.cell_core_db : k => v.arn }
}

output "cell_finance_db_secret_arns" {
  value = { for k, v in aws_secretsmanager_secret.cell_finance_db : k => v.arn }
}

output "objects_bucket_name" {
  value = aws_s3_bucket.objects.bucket
}

output "objects_bucket_arn" {
  value = aws_s3_bucket.objects.arn
}

output "kms_key_arn" {
  value = aws_kms_key.hub.arn
}

output "sns_topic_arns" {
  description = "ARNs dos tópicos SNS de fatos (hub-operation-facts, hub-protocol-facts)"
  value       = { for k, v in aws_sns_topic.facts : k => v.arn }
}

output "sqs_queue_urls" {
  description = "URLs das filas SQS Standard (nomes replicados de hub/internal/queue/queue.go e hub/cmd/*/main.go)"
  value       = { for k, v in aws_sqs_queue.this : k => v.id }
}

output "sqs_dlq_urls" {
  value = { for k, v in aws_sqs_queue.dlq : k => v.id }
}

output "pending_decision_notice" {
  description = "Lembrete explícito: nenhum destes recursos foi aplicado. P-01 (conta/região/orçamento) segue pendente."
  value       = "NENHUM RECURSO APLICADO — P-01 pendente (ver openspec/changes/hub-interoperabilidade-v4/specs/decisoes-e-governanca/spec.md). Este Terraform é referência de sintaxe/arquitetura, não infraestrutura provisionada."
}
