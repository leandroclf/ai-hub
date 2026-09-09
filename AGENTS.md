# Instruções do repositório

## Identidade Git

- A conta GitHub responsável por este repositório é `leandroclf`.
- Use a identidade Git local `leandroclf` ao criar commits neste repositório.
- Nunca registre tokens, senhas ou chaves privadas neste arquivo ou no repositório.

## Conteúdo versionável

- Versione a documentação mantida em `docs/` e as instruções do repositório.
- Não versione configurações locais da IDE (`.idea/`) nem artefatos locais do Claude Code (`.claude/`).

## Ecossistema Compose local

- A máquina local deve manter no máximo um ecossistema de componentes do AI Hub ativo por vez. Não suba projetos Compose paralelos para a mesma qualificação.
- Antes de iniciar ou trocar o ambiente, inspecione `docker compose ls` e `docker ps -a` para identificar o projeto atualmente ativo e seus componentes.
- Use um nome de projeto Compose único e estável para o ambiente oficial local. Ao substituir a versão, pare o ecossistema anterior, confirme que os containers foram encerrados e só então suba o novo.
- É permitido remover containers, redes e imagens antigas do ecossistema local quando forem identificados como pertencentes ao AI Hub e não forem mais necessários. Preserve volumes e dados por padrão; remova volumes somente quando o trabalho autorizar explicitamente recriação dos dados locais.
- Não use `docker system prune`, curingas amplos ou comandos que possam afetar projetos fora do AI Hub. Não remova containers de terceiros nem o laboratório de outro trabalho sem identificar precisamente o alvo.
- Depois de uma troca, confirme com `docker compose ps` e `docker ps` que apenas o ecossistema escolhido está rodando. Containers parados antigos também devem ser limpos quando não forem necessários, para evitar confusão e consumo de disco.
- Se houver containers de uma execução interrompida pelo agente, o agente deve limpar somente esses recursos antes de encerrar ou registrar no checkpoint o motivo concreto para não fazê-lo.
