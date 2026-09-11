# Identidade visual HIVEPlace

Esta pasta guarda as referências visuais fornecidas para a marca HIVEPlace. A aplicação do sistema visual no portal administrativo usa tokens CSS e o símbolo hexagonal em SVG, sem depender de imagens de campanha durante o carregamento da interface.

## Sistema visual aplicado

| Elemento | Definição | Uso principal |
| --- | --- | --- |
| Amarelo HIVE | `#F5B719` | Marca, ação primária, foco e destaque |
| Amarelo profundo | `#B97900` | Links e texto de destaque sobre fundo claro |
| Preto HIVE | `#0A0A0A` | Cabeçalho, contraste e áreas de marca |
| Tinta | `#262124` | Títulos e conteúdo principal |
| Papel | `#FFFEFA` | Cartões, formulários e superfícies de trabalho |
| Fundo quente | `#F5F4F1` | Fundo geral da aplicação |

### Tipografia

- **Manrope** para marca, títulos e hierarquia editorial.
- **DM Sans** para navegação, formulários, dados e textos corridos.
- A pilha de fallback mantém `Avenir Next`, `Avenir` e `Segoe UI` para ambientes sem carregamento da fonte web.

### Princípios de interface

1. Fundo preto e amarelo são reservados para marca, navegação ativa e ações de alta prioridade.
2. Cartões usam branco quente, borda discreta e sombra baixa; a interface não depende de sombras fortes para indicar agrupamento.
3. Estados de sucesso, erro e aviso mantêm semântica própria e não usam amarelo como substituto de erro.
4. O hexágono é usado como símbolo de marca e como motivo decorativo; nunca substitui texto acessível.
5. Foco visível, contraste e reflow móvel fazem parte do sistema, não são ajustes posteriores.

## Arquivos de referência

- `Capa Linkedin - HIVE fundo branco.png`
- `Capa Linkedin - HIVE fundo branco 2.png`
- `Capa Linkedin - HIVE fundo preto.png`
- `Capa Linkedin - HIVE fundo preto 2.png`

## Implementação

A base está em `hub/admin-ui/src/index.css`; o símbolo reutilizável está em `hub/admin-ui/src/components/BrandMark.tsx`. Novas páginas devem consumir os tokens `--hive-*` em vez de introduzir cores ad hoc.
