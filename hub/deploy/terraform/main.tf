# Topologia AWS de REFERÊNCIA (ppd/prd) do Hub de Interoperabilidade
# "Constelação" v4. Ver README.md deste diretório: nada aqui foi
# aplicado a uma conta real (P-01 pendente — DEC-02).
#
# Fontes normativas:
#   - openspec/changes/hub-interoperabilidade-v4/design.md
#   - .../specs/arquitetura-e-comunicacao/spec.md (ARQ-03/05/06)
#   - .../specs/decisoes-e-governanca/spec.md (DEC-02, P-01)
# Nomes exatos de filas/tópicos replicados de:
#   - hub/internal/queue/queue.go
#   - hub/cmd/{orbita,cometa,pulsar,libra}/main.go

data "aws_availability_zones" "available" {
  state = "available"
}

locals {
  azs = slice(data.aws_availability_zones.available.names, 0, var.az_count)

  # /20 por subnet a partir do /16 da VPC, públicas e privadas
  # intercaladas por índice de AZ.
  public_subnet_cidrs  = [for i in range(var.az_count) : cidrsubnet(var.vpc_cidr, 4, i)]
  private_subnet_cidrs = [for i in range(var.az_count) : cidrsubnet(var.vpc_cidr, 4, i + var.az_count)]

  name = var.project_name

  # Nomes exatos de filas/tópicos usados hoje pela implementação Go
  # (hub/internal/queue/queue.go, hub/cmd/*/main.go). Não são
  # namespaced por célula porque o código atual também não namespaces
  # — uma célula adicional exigiria decisão/alteração de código futura
  # (fora do escopo desta revisão de Terraform).
  queue_topology = {
    # fila de comandos ASYNC/AUTO de Órbita -> Cometa (COM-01, EXE-15);
    # nunca usada no despacho DIRECT de SYNC.
    "cometa-commands" = { topic = null }

    # fatos de operação (Cometa -> SNS hub-operation-facts), fan-out
    # para Órbita e Libra (custo).
    "orbita-operation-facts" = { topic = "hub-operation-facts" }
    "libra-cost-facts"       = { topic = "hub-operation-facts" }

    # fatos de protocolo (Órbita -> SNS hub-protocol-facts), fan-out
    # para Pulsar (webhook) e Libra (receita).
    "pulsar-protocol-facts" = { topic = "hub-protocol-facts" }
    "libra-revenue-facts"   = { topic = "hub-protocol-facts" }
  }

  topics = toset(["hub-operation-facts", "hub-protocol-facts"])
}

# =============================================================================
# Rede: VPC multi-AZ com subnets públicas/privadas (OPE-04, ARQ-03/06)
# =============================================================================

resource "aws_vpc" "this" {
  cidr_block           = var.vpc_cidr
  enable_dns_support   = true
  enable_dns_hostnames = true

  tags = {
    Name = "${local.name}-vpc"
  }
}

resource "aws_internet_gateway" "this" {
  vpc_id = aws_vpc.this.id

  tags = {
    Name = "${local.name}-igw"
  }
}

resource "aws_subnet" "public" {
  count                   = var.az_count
  vpc_id                  = aws_vpc.this.id
  cidr_block              = local.public_subnet_cidrs[count.index]
  availability_zone       = local.azs[count.index]
  map_public_ip_on_launch = true

  tags = {
    Name                                      = "${local.name}-public-${local.azs[count.index]}"
    "kubernetes.io/role/elb"                  = "1"
    "kubernetes.io/cluster/${local.name}-eks" = "shared"
  }
}

resource "aws_subnet" "private" {
  count             = var.az_count
  vpc_id            = aws_vpc.this.id
  cidr_block        = local.private_subnet_cidrs[count.index]
  availability_zone = local.azs[count.index]

  tags = {
    Name                                      = "${local.name}-private-${local.azs[count.index]}"
    "kubernetes.io/role/internal-elb"         = "1"
    "kubernetes.io/cluster/${local.name}-eks" = "shared"
    "karpenter.sh/discovery"                  = "${local.name}-eks"
  }
}

# Um NAT Gateway por AZ (referência de alta disponibilidade; um único
# NAT Gateway compartilhado seria um ponto único de falha entre AZs,
# incompatível com OPE-04/ARQ-03).
resource "aws_eip" "nat" {
  count  = var.az_count
  domain = "vpc"

  tags = {
    Name = "${local.name}-nat-eip-${local.azs[count.index]}"
  }
}

resource "aws_nat_gateway" "this" {
  count         = var.az_count
  allocation_id = aws_eip.nat[count.index].id
  subnet_id     = aws_subnet.public[count.index].id

  tags = {
    Name = "${local.name}-nat-${local.azs[count.index]}"
  }

  depends_on = [aws_internet_gateway.this]
}

resource "aws_route_table" "public" {
  vpc_id = aws_vpc.this.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.this.id
  }

  tags = {
    Name = "${local.name}-public-rt"
  }
}

resource "aws_route_table_association" "public" {
  count          = var.az_count
  subnet_id      = aws_subnet.public[count.index].id
  route_table_id = aws_route_table.public.id
}

resource "aws_route_table" "private" {
  count  = var.az_count
  vpc_id = aws_vpc.this.id

  route {
    cidr_block     = "0.0.0.0/0"
    nat_gateway_id = aws_nat_gateway.this[count.index].id
  }

  tags = {
    Name = "${local.name}-private-rt-${local.azs[count.index]}"
  }
}

resource "aws_route_table_association" "private" {
  count          = var.az_count
  subnet_id      = aws_subnet.private[count.index].id
  route_table_id = aws_route_table.private[count.index].id
}

# =============================================================================
# KMS (Secrets Manager + KMS, ARQ-03/ARQ-05)
# =============================================================================

resource "aws_kms_key" "hub" {
  description             = "${local.name}: chave de criptografia de referência para RDS, Secrets Manager, S3 e SQS/SNS"
  deletion_window_in_days = 30
  enable_key_rotation     = true

  # Protegida contra exclusão acidental (ARQ-06: bancos, chaves e
  # buckets protegidos contra exclusão automática).
  lifecycle {
    prevent_destroy = true
  }
}

resource "aws_kms_alias" "hub" {
  name          = "alias/${local.name}"
  target_key_id = aws_kms_key.hub.key_id
}

# =============================================================================
# S3 (DAD-05): bucket de objetos, versionado e sem acesso público
# =============================================================================

resource "aws_s3_bucket" "objects" {
  bucket        = "${local.name}-objects-${var.environment}"
  force_destroy = var.objects_bucket_force_destroy

  lifecycle {
    prevent_destroy = true
  }

  tags = {
    Name = "${local.name}-objects-${var.environment}"
  }
}

resource "aws_s3_bucket_versioning" "objects" {
  bucket = aws_s3_bucket.objects.id

  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_bucket_server_side_encryption_configuration" "objects" {
  bucket = aws_s3_bucket.objects.id

  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm     = "aws:kms"
      kms_master_key_id = aws_kms_key.hub.arn
    }
    bucket_key_enabled = true
  }
}

resource "aws_s3_bucket_public_access_block" "objects" {
  bucket = aws_s3_bucket.objects.id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

resource "aws_s3_bucket_lifecycle_configuration" "objects" {
  bucket = aws_s3_bucket.objects.id

  rule {
    id     = "abort-incomplete-multipart"
    status = "Enabled"

    filter {
      prefix = ""
    }

    abort_incomplete_multipart_upload {
      days_after_initiation = 7
    }
  }
}

# =============================================================================
# Secrets Manager: apenas REFERÊNCIAS de segredo (não valores fixos).
# O valor é gerado aleatoriamente por random_password no próprio
# apply — nunca commitado, nunca um segredo real de conta aprovada.
# =============================================================================

resource "random_password" "control_db" {
  length  = 32
  special = false
}

resource "aws_secretsmanager_secret" "control_db" {
  name       = "${local.name}/${var.environment}/hub_control/db"
  kms_key_id = aws_kms_key.hub.arn

  tags = {
    Domain = "hub_control"
  }
}

resource "aws_secretsmanager_secret_version" "control_db" {
  secret_id = aws_secretsmanager_secret.control_db.id
  secret_string = jsonencode({
    engine   = "postgres"
    username = "hub_control_app"
    password = random_password.control_db.result
    dbname   = "hub_control"
  })
}

resource "random_password" "cell_core_db" {
  for_each = toset(var.cells)
  length   = 32
  special  = false
}

resource "aws_secretsmanager_secret" "cell_core_db" {
  for_each   = toset(var.cells)
  name       = "${local.name}/${var.environment}/${each.key}/hub_core/db"
  kms_key_id = aws_kms_key.hub.arn

  tags = {
    Domain = "hub_core"
    Cell   = each.key
  }
}

resource "aws_secretsmanager_secret_version" "cell_core_db" {
  for_each  = toset(var.cells)
  secret_id = aws_secretsmanager_secret.cell_core_db[each.key].id
  secret_string = jsonencode({
    engine   = "postgres"
    username = "hub_core_app"
    password = random_password.cell_core_db[each.key].result
    dbname   = "hub_core"
  })
}

resource "random_password" "cell_finance_db" {
  for_each = toset(var.cells)
  length   = 32
  special  = false
}

resource "aws_secretsmanager_secret" "cell_finance_db" {
  for_each   = toset(var.cells)
  name       = "${local.name}/${var.environment}/${each.key}/hub_finance/db"
  kms_key_id = aws_kms_key.hub.arn

  tags = {
    Domain = "hub_finance"
    Cell   = each.key
  }
}

resource "aws_secretsmanager_secret_version" "cell_finance_db" {
  for_each  = toset(var.cells)
  secret_id = aws_secretsmanager_secret.cell_finance_db[each.key].id
  secret_string = jsonencode({
    engine   = "postgres"
    username = "hub_finance_app"
    password = random_password.cell_finance_db[each.key].result
    dbname   = "hub_finance"
  })
}

# =============================================================================
# PostgreSQL (DAD-01/ADR-03, DAD-07): hub_control é único/global;
# hub_core e hub_finance são um par por célula, sem cluster global
# (ARQ-05: separar banco por célula — servidor compartilhado não é
# isolamento de I/O suficiente).
# =============================================================================

resource "aws_db_subnet_group" "this" {
  name       = "${local.name}-db"
  subnet_ids = aws_subnet.private[*].id

  tags = {
    Name = "${local.name}-db-subnet-group"
  }
}

resource "aws_security_group" "db" {
  name        = "${local.name}-db-sg"
  description = "Acesso PostgreSQL somente dos workloads internos do EKS (referência; refinar por security group do node/pod real)"
  vpc_id      = aws_vpc.this.id

  ingress {
    description     = "PostgreSQL a partir dos nós/pods do EKS"
    from_port       = 5432
    to_port         = 5432
    protocol        = "tcp"
    security_groups = [aws_security_group.eks_nodes.id]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = "${local.name}-db-sg"
  }
}

# hub_control (Atlas) — banco de controle único, global, independente
# de célula.
resource "aws_db_instance" "control" {
  identifier     = "${local.name}-hub-control"
  engine         = "postgres"
  engine_version = var.db_engine_version
  instance_class = var.db_instance_class

  allocated_storage     = var.db_allocated_storage_gb
  max_allocated_storage = var.db_max_allocated_storage_gb
  storage_type          = "gp3"
  storage_encrypted     = true
  kms_key_id            = aws_kms_key.hub.arn

  db_name  = "hub_control"
  username = "hub_control_app"
  password = random_password.control_db.result

  multi_az               = true
  db_subnet_group_name   = aws_db_subnet_group.this.name
  vpc_security_group_ids = [aws_security_group.db.id]

  backup_retention_period = var.db_backup_retention_days
  backup_window           = "03:00-04:00"
  maintenance_window      = "sun:04:30-sun:05:30"

  deletion_protection       = true
  skip_final_snapshot       = false
  final_snapshot_identifier = "${local.name}-hub-control-final"
  copy_tags_to_snapshot     = true

  enabled_cloudwatch_logs_exports = ["postgresql", "upgrade"]

  lifecycle {
    prevent_destroy = true
  }

  tags = {
    Domain = "hub_control"
    Cell   = "global"
  }
}

# hub_core e hub_finance por célula (uma instância cada). Modelado
# aqui para "cell-01" via var.cells; uma nova célula soma outro par de
# instâncias (ver módulo Crossplane em ../crossplane para o
# equivalente declarativo de auto-atendimento).
resource "aws_db_instance" "cell_core" {
  for_each = toset(var.cells)

  identifier     = "${local.name}-${each.key}-hub-core"
  engine         = "postgres"
  engine_version = var.db_engine_version
  instance_class = var.db_instance_class

  allocated_storage     = var.db_allocated_storage_gb
  max_allocated_storage = var.db_max_allocated_storage_gb
  storage_type          = "gp3"
  storage_encrypted     = true
  kms_key_id            = aws_kms_key.hub.arn

  db_name  = "hub_core"
  username = "hub_core_app"
  password = random_password.cell_core_db[each.key].result

  multi_az               = true
  db_subnet_group_name   = aws_db_subnet_group.this.name
  vpc_security_group_ids = [aws_security_group.db.id]

  backup_retention_period = var.db_backup_retention_days
  backup_window           = "03:00-04:00"
  maintenance_window      = "sun:04:30-sun:05:30"

  deletion_protection       = true
  skip_final_snapshot       = false
  final_snapshot_identifier = "${local.name}-${each.key}-hub-core-final"
  copy_tags_to_snapshot     = true

  enabled_cloudwatch_logs_exports = ["postgresql", "upgrade"]

  lifecycle {
    prevent_destroy = true
  }

  tags = {
    Domain = "hub_core"
    Cell   = each.key
  }
}

resource "aws_db_instance" "cell_finance" {
  for_each = toset(var.cells)

  identifier     = "${local.name}-${each.key}-hub-finance"
  engine         = "postgres"
  engine_version = var.db_engine_version
  instance_class = var.db_instance_class

  allocated_storage     = var.db_allocated_storage_gb
  max_allocated_storage = var.db_max_allocated_storage_gb
  storage_type          = "gp3"
  storage_encrypted     = true
  kms_key_id            = aws_kms_key.hub.arn

  db_name  = "hub_finance"
  username = "hub_finance_app"
  password = random_password.cell_finance_db[each.key].result

  multi_az               = true
  db_subnet_group_name   = aws_db_subnet_group.this.name
  vpc_security_group_ids = [aws_security_group.db.id]

  backup_retention_period = var.db_backup_retention_days
  backup_window           = "03:00-04:00"
  maintenance_window      = "sun:04:30-sun:05:30"

  # hub_finance é a autoridade financeira única do contrato de saldo
  # estrito (Decision 2 do design.md) — proteção reforçada.
  deletion_protection       = true
  skip_final_snapshot       = false
  final_snapshot_identifier = "${local.name}-${each.key}-hub-finance-final"
  copy_tags_to_snapshot     = true

  enabled_cloudwatch_logs_exports = ["postgresql", "upgrade"]

  lifecycle {
    prevent_destroy = true
  }

  tags = {
    Domain = "hub_finance"
    Cell   = each.key
  }
}

# =============================================================================
# SNS + SQS Standard (ARQ-03/ARQ-05, COM-01/COM-03): fila por
# consumidor e domínio de isolamento, sem presumir ordem. Nomes
# replicados literalmente de hub/internal/queue/queue.go e dos
# hub/cmd/*/main.go.
# =============================================================================

resource "aws_sns_topic" "facts" {
  for_each          = local.topics
  name              = each.key
  kms_master_key_id = aws_kms_key.hub.id

  tags = {
    Purpose = "hub-fact-fanout"
  }
}

# Dead-letter queue por fila (OPE-06: alerta de "DLQ com item").
resource "aws_sqs_queue" "dlq" {
  for_each                  = local.queue_topology
  name                      = "${each.key}-dlq"
  message_retention_seconds = 1209600 # 14 dias
  kms_master_key_id         = aws_kms_key.hub.id
  sqs_managed_sse_enabled   = false

  tags = {
    Purpose = "dlq"
    Queue   = each.key
  }
}

resource "aws_sqs_queue" "this" {
  for_each = local.queue_topology

  name                       = each.key
  visibility_timeout_seconds = 60
  message_retention_seconds  = 345600 # 4 dias
  kms_master_key_id          = aws_kms_key.hub.id
  sqs_managed_sse_enabled    = false

  redrive_policy = jsonencode({
    deadLetterTargetArn = aws_sqs_queue.dlq[each.key].arn
    maxReceiveCount     = 5
  })

  tags = {
    Purpose = "hub-queue"
  }
}

# Política de fila: somente o(s) tópico(s) SNS correspondente(s) pode(m)
# publicar (aws:SourceArn), nunca "Principal: *" — diferente do atalho
# de bootstrap local em queue.go (AllowSNSDelivery), que é
# deliberadamente permissivo apenas para LocalStack.
resource "aws_sqs_queue_policy" "sns_delivery" {
  for_each = { for name, cfg in local.queue_topology : name => cfg if cfg.topic != null }

  queue_url = aws_sqs_queue.this[each.key].id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Sid       = "AllowSNSDelivery"
      Effect    = "Allow"
      Principal = { Service = "sns.amazonaws.com" }
      Action    = "sqs:SendMessage"
      Resource  = aws_sqs_queue.this[each.key].arn
      Condition = {
        ArnEquals = {
          "aws:SourceArn" = aws_sns_topic.facts[each.value.topic].arn
        }
      }
    }]
  })
}

resource "aws_sns_topic_subscription" "fanout" {
  for_each = { for name, cfg in local.queue_topology : name => cfg if cfg.topic != null }

  topic_arn = aws_sns_topic.facts[each.value.topic].arn
  protocol  = "sqs"
  endpoint  = aws_sqs_queue.this[each.key].arn
}

# =============================================================================
# EKS multi-AZ (ARQ-03: Kubernetes, referência EKS multi-AZ)
# =============================================================================

resource "aws_security_group" "eks_cluster" {
  name        = "${local.name}-eks-cluster-sg"
  description = "Security group do control plane EKS"
  vpc_id      = aws_vpc.this.id

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = "${local.name}-eks-cluster-sg"
  }
}

resource "aws_security_group" "eks_nodes" {
  name        = "${local.name}-eks-nodes-sg"
  description = "Security group dos nós/workers EKS"
  vpc_id      = aws_vpc.this.id

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name                                      = "${local.name}-eks-nodes-sg"
    "kubernetes.io/cluster/${local.name}-eks" = "owned"
  }
}

resource "aws_security_group_rule" "nodes_from_cluster" {
  type                     = "ingress"
  from_port                = 0
  to_port                  = 65535
  protocol                 = "-1"
  security_group_id        = aws_security_group.eks_nodes.id
  source_security_group_id = aws_security_group.eks_cluster.id
}

resource "aws_security_group_rule" "cluster_from_nodes" {
  type                     = "ingress"
  from_port                = 443
  to_port                  = 443
  protocol                 = "tcp"
  security_group_id        = aws_security_group.eks_cluster.id
  source_security_group_id = aws_security_group.eks_nodes.id
}

resource "aws_iam_role" "eks_cluster" {
  name = "${local.name}-eks-cluster-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect    = "Allow"
      Principal = { Service = "eks.amazonaws.com" }
      Action    = "sts:AssumeRole"
    }]
  })
}

resource "aws_iam_role_policy_attachment" "eks_cluster_policy" {
  role       = aws_iam_role.eks_cluster.name
  policy_arn = "arn:aws:iam::aws:policy/AmazonEKSClusterPolicy"
}

resource "aws_eks_cluster" "this" {
  name     = "${local.name}-eks"
  role_arn = aws_iam_role.eks_cluster.arn
  version  = var.eks_cluster_version

  vpc_config {
    subnet_ids              = concat(aws_subnet.private[*].id, aws_subnet.public[*].id)
    security_group_ids      = [aws_security_group.eks_cluster.id]
    endpoint_private_access = true
    # Acesso público existe apenas nesta referência para permitir
    # kubectl/CI sem bastion; em prd real avaliar restringir por CIDR
    # (decisão de rede, fora do escopo desta revisão — ver P-01).
    endpoint_public_access = true
  }

  enabled_cluster_log_types = ["api", "audit", "authenticator", "controllerManager", "scheduler"]

  depends_on = [aws_iam_role_policy_attachment.eks_cluster_policy]

  tags = {
    Name = "${local.name}-eks"
  }
}

# OIDC provider do cluster, base do IRSA (IAM Roles for Service
# Accounts) usado por Karpenter, KEDA e Crossplane (ARQ-06).
data "tls_certificate" "eks_oidc" {
  url = aws_eks_cluster.this.identity[0].oidc[0].issuer
}

resource "aws_iam_openid_connect_provider" "eks" {
  url             = aws_eks_cluster.this.identity[0].oidc[0].issuer
  client_id_list  = ["sts.amazonaws.com"]
  thumbprint_list = [data.tls_certificate.eks_oidc.certificates[0].sha1_fingerprint]
}

# Node group gerenciado de base: piso mínimo aquecido (ARQ-06: KEDA
# mantém mínimo aquecido; Karpenter cobre o restante da elasticidade
# via NodePools configurados no cluster, fora deste Terraform).
resource "aws_iam_role" "eks_node" {
  name = "${local.name}-eks-node-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect    = "Allow"
      Principal = { Service = "ec2.amazonaws.com" }
      Action    = "sts:AssumeRole"
    }]
  })
}

resource "aws_iam_role_policy_attachment" "eks_node_worker" {
  role       = aws_iam_role.eks_node.name
  policy_arn = "arn:aws:iam::aws:policy/AmazonEKSWorkerNodePolicy"
}

resource "aws_iam_role_policy_attachment" "eks_node_cni" {
  role       = aws_iam_role.eks_node.name
  policy_arn = "arn:aws:iam::aws:policy/AmazonEKS_CNI_Policy"
}

resource "aws_iam_role_policy_attachment" "eks_node_ecr" {
  role       = aws_iam_role.eks_node.name
  policy_arn = "arn:aws:iam::aws:policy/AmazonEC2ContainerRegistryReadOnly"
}

resource "aws_eks_node_group" "baseline" {
  cluster_name    = aws_eks_cluster.this.name
  node_group_name = "${local.name}-baseline"
  node_role_arn   = aws_iam_role.eks_node.arn
  subnet_ids      = aws_subnet.private[*].id

  instance_types = var.eks_node_instance_types

  scaling_config {
    desired_size = var.eks_node_group_desired_size
    min_size     = var.eks_node_group_min_size
    max_size     = var.eks_node_group_max_size
  }

  update_config {
    max_unavailable = 1
  }

  labels = {
    "hub.constelacao/pool" = "baseline"
  }

  depends_on = [
    aws_iam_role_policy_attachment.eks_node_worker,
    aws_iam_role_policy_attachment.eks_node_cni,
    aws_iam_role_policy_attachment.eks_node_ecr,
  ]

  tags = {
    Name = "${local.name}-baseline-ng"
  }
}

# ---------------------------------------------------------------------------
# Karpenter (ARQ-06): role de nó dedicada aos nós provisionados
# dinamicamente pelos NodePools do Karpenter (configuração dos
# NodePools/EC2NodeClass é manifesto Kubernetes, fora do escopo desta
# revisão de Terraform), mais a role IRSA do controller.
# ---------------------------------------------------------------------------

resource "aws_iam_role" "karpenter_node" {
  name = "${local.name}-karpenter-node-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect    = "Allow"
      Principal = { Service = "ec2.amazonaws.com" }
      Action    = "sts:AssumeRole"
    }]
  })
}

resource "aws_iam_role_policy_attachment" "karpenter_node_worker" {
  role       = aws_iam_role.karpenter_node.name
  policy_arn = "arn:aws:iam::aws:policy/AmazonEKSWorkerNodePolicy"
}

resource "aws_iam_role_policy_attachment" "karpenter_node_cni" {
  role       = aws_iam_role.karpenter_node.name
  policy_arn = "arn:aws:iam::aws:policy/AmazonEKS_CNI_Policy"
}

resource "aws_iam_role_policy_attachment" "karpenter_node_ecr" {
  role       = aws_iam_role.karpenter_node.name
  policy_arn = "arn:aws:iam::aws:policy/AmazonEC2ContainerRegistryReadOnly"
}

resource "aws_iam_role_policy_attachment" "karpenter_node_ssm" {
  role       = aws_iam_role.karpenter_node.name
  policy_arn = "arn:aws:iam::aws:policy/AmazonSSMManagedInstanceCore"
}

resource "aws_iam_instance_profile" "karpenter_node" {
  name = "${local.name}-karpenter-node-profile"
  role = aws_iam_role.karpenter_node.name
}

resource "aws_iam_role" "karpenter_controller" {
  name = "${local.name}-karpenter-controller-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect = "Allow"
      Principal = {
        Federated = aws_iam_openid_connect_provider.eks.arn
      }
      Action = "sts:AssumeRoleWithWebIdentity"
      Condition = {
        StringEquals = {
          "${replace(aws_iam_openid_connect_provider.eks.url, "https://", "")}:aud" = "sts.amazonaws.com"
          "${replace(aws_iam_openid_connect_provider.eks.url, "https://", "")}:sub" = "system:serviceaccount:karpenter:karpenter"
        }
      }
    }]
  })
}

# Política mínima de referência (Karpenter real reduz ainda mais o
# escopo por tag/condição — revisar antes de produção, ARQ-05).
resource "aws_iam_role_policy" "karpenter_controller" {
  name = "${local.name}-karpenter-controller-policy"
  role = aws_iam_role.karpenter_controller.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect = "Allow"
      Action = [
        "ec2:RunInstances",
        "ec2:TerminateInstances",
        "ec2:DescribeInstances",
        "ec2:DescribeInstanceTypes",
        "ec2:DescribeLaunchTemplates",
        "ec2:DescribeSubnets",
        "ec2:DescribeSecurityGroups",
        "ec2:DescribeAvailabilityZones",
        "ec2:CreateLaunchTemplate",
        "ec2:CreateTags",
        "eks:DescribeCluster",
        "iam:PassRole",
        "pricing:GetProducts",
        "ssm:GetParameter",
      ]
      Resource = "*"
    }]
  })
}

# ---------------------------------------------------------------------------
# KEDA (ARQ-06): role IRSA para o scaler ler backlog/idade das filas
# SQS deste hub (métrica de escala dos workers).
# ---------------------------------------------------------------------------

resource "aws_iam_role" "keda_operator" {
  name = "${local.name}-keda-operator-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect = "Allow"
      Principal = {
        Federated = aws_iam_openid_connect_provider.eks.arn
      }
      Action = "sts:AssumeRoleWithWebIdentity"
      Condition = {
        StringEquals = {
          "${replace(aws_iam_openid_connect_provider.eks.url, "https://", "")}:aud" = "sts.amazonaws.com"
          "${replace(aws_iam_openid_connect_provider.eks.url, "https://", "")}:sub" = "system:serviceaccount:keda:keda-operator"
        }
      }
    }]
  })
}

resource "aws_iam_role_policy" "keda_operator" {
  name = "${local.name}-keda-operator-policy"
  role = aws_iam_role.keda_operator.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect = "Allow"
      Action = [
        "sqs:GetQueueAttributes",
        "sqs:GetQueueUrl",
      ]
      Resource = [for q in aws_sqs_queue.this : q.arn]
    }]
  })
}

# ---------------------------------------------------------------------------
# Crossplane (ARQ-06, CFG-06, OPE-12): role IRSA de referência do
# provider-aws do Crossplane, que reconcilia os recursos desejados de
# uma nova célula (RDS/SQS/SNS) publicados por Atlas — a decisão de
# QUANDO expandir continua em Atlas/política de capacidade, nunca no
# próprio Crossplane (ARQ-06). Ver módulo em ../crossplane.
# ---------------------------------------------------------------------------

resource "aws_iam_role" "crossplane_provider_aws" {
  name = "${local.name}-crossplane-provider-aws-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect = "Allow"
      Principal = {
        Federated = aws_iam_openid_connect_provider.eks.arn
      }
      Action = "sts:AssumeRoleWithWebIdentity"
      Condition = {
        StringEquals = {
          "${replace(aws_iam_openid_connect_provider.eks.url, "https://", "")}:aud" = "sts.amazonaws.com"
          "${replace(aws_iam_openid_connect_provider.eks.url, "https://", "")}:sub" = "system:serviceaccount:crossplane-system:provider-aws"
        }
      }
    }]
  })
}

# Escopo amplo de referência apenas para o plano de gestão de células
# (rds/sqs/sns); NÃO SHALL ser usado como modelo final de produção sem
# reduzir por tag/condição/limite de conta (ARQ-05: dependência de
# produção exige revisão de alternativa/escopo antes de liberar).
resource "aws_iam_role_policy" "crossplane_provider_aws" {
  name = "${local.name}-crossplane-provider-aws-policy"
  role = aws_iam_role.crossplane_provider_aws.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect = "Allow"
      Action = [
        "rds:CreateDBInstance",
        "rds:DescribeDBInstances",
        "rds:ModifyDBInstance",
        "rds:DeleteDBInstance",
        "rds:AddTagsToResource",
        "sqs:CreateQueue",
        "sqs:GetQueueAttributes",
        "sqs:SetQueueAttributes",
        "sqs:DeleteQueue",
        "sns:CreateTopic",
        "sns:GetTopicAttributes",
        "sns:Subscribe",
        "sns:DeleteTopic",
        "secretsmanager:CreateSecret",
        "secretsmanager:GetSecretValue",
        "secretsmanager:PutSecretValue",
      ]
      Resource = "*"
    }]
  })
}
