# TelaClientes

A lista de clientes, ponto de entrada do painel: busca, filtro de consentimento e tabela paginada.

Composição: `SideNav` + `PageHeader` (ação primária "Cadastrar cliente") + um `Panel` com `pf-search`, `SegmentedControl` (Todos, Concedido, Não concedido, Revogado) e `DataTable` com Cliente, E-mail mascarado, Consentimento (`StatusTag`), Segmento e Cadastro. Comportamento e estados vazios: seção Telas e fluxos. Protótipo: `prototipo-manha.html`, rota `#/clientes`.
