# Prova estática de alteração de migração
Anterior SHA: a4a876a9f8e875db882f7ca45cf7dece24d57aee
Atual SHA: b9d0f90ce02aa0c27cad546745153d160ff5867f
Arquivo: hub/migrations/control/0002_provider_auth.sql
SHA-256 anterior: 37241e604376471efd7de394e9394673feaec0a32476a517c38363890dad1541
SHA-256 atual: f4174fbf3e0157b42c8d8fbed0d92405337a9a3bd00647af2f9a52f4c4390fbb

O runner hub/deploy/r2/scripts/migrate.sh compara checksum de schema_migrations e encerra com código 3
em mismatch. Se o banco foi migrado com a versão anterior, o upgrade atinge esse ramo antes da 0004.
A condição é demonstrável pelo conteúdo; não foi executado PostgreSQL nesta sessão.
O plano de correção deve considerar também bancos criados com o hash atual; não editar ledger cegamente.
