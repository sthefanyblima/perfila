# Textos da interface

## Estados do domínio

| Valor no sistema | Na tela | Etiqueta (`StatusTag`) |
| --- | --- | --- |
| `NAO_CONCEDIDO` | Não concedido | neutral, círculo tracejado |
| `CONCEDIDO` | Concedido | ok, círculo com visto |
| `REVOGADO` | Revogado | critical, círculo com barra |
| `GERADA` | Gerada | warn, triângulo (raramente visível: o sistema já grava como Ativa) |
| `ATIVA` | Ativa | ok, círculo com visto |
| `EXPIRADA` | Expirada | neutral, círculo vazado |
| cliente `Ativo = false` | Inativo | neutral, círculo vazado |
| tipo `compra` / `navegacao` / `interacao` | Compra / Navegação / Interação | — |

## Códigos de erro da API → mensagem

O código estável (`errs.Codigo`) define a mensagem; o texto técnico do servidor nunca aparece.

| Código | Onde aparece | Mensagem |
| --- | --- | --- |
| `nao_autenticado` | Entrar | Usuário ou senha incorretos. |
| `cliente_nao_encontrado` | Busca, ficha | Não encontramos esse cliente. Confira o ID ou busque pelo nome. |
| `cliente_inativo` | Ficha | Este cadastro está inativo. Eventos e processamento ficam bloqueados enquanto ele estiver assim. |
| `consentimento_invalido` | Ficha, eventos, processamento | Sem consentimento válido. Registre o consentimento do titular para usar os dados de comportamento. |
| `transicao_invalida` (conceder) | Consentimento | O consentimento já está concedido. Para mudar a versão do termo, revogue e conceda de novo. |
| `transicao_invalida` (revogar) | Consentimento | Não há consentimento concedido para revogar. |
| `dados_invalidos` | Formulários | Mensagem por campo (abaixo). |
| `perfil_nao_encontrado` | Visão geral | Este cliente ainda não foi processado. Use "Processar agora". |
| `sem_produto_elegivel` | Recomendação (não é erro) | Nenhum produto ativo na categoria preferida ({categoria}). |
| `erro_interno` | Qualquer | Algo falhou do nosso lado. As etapas concluídas antes da falha ficaram salvas. Tente de novo; se continuar, informe o código {X-Request-ID} à equipe técnica. |
| 503 `indisponivel` | Qualquer | O banco de dados não está respondendo. Nada foi alterado. Tente de novo em instantes. |

## Validação de campos (RN-04 e cadastro)

| Campo | Regra | Mensagem |
| --- | --- | --- |
| Nome | obrigatório | Informe o nome do cliente. |
| E-mail | obrigatório | Informe o e-mail do cliente. |
| Finalidade | obrigatória | Informe para que os dados serão usados. |
| Versão do termo | obrigatória | Informe a versão do termo que o titular aceitou. |
| Tipo do evento | um dos três | Escolha Compra, Navegação ou Interação. |
| Origem | obrigatória | Informe de onde veio o evento (por exemplo, loja on-line). |
| Valor | ≥ 0 e ≤ R$ 1.000.000,00 | O valor deve estar entre R$ 0,00 e R$ 1.000.000,00. |
| Valor (compra) | > 0 | Uma compra precisa ter valor maior que zero. |
| Data e hora | até 2 min no futuro | A data não pode estar no futuro. |

Valide no navegador (atributos HTML: `required`, `min`, `max`, `step`) para dar retorno imediato, e sempre de novo no servidor, que é quem decide.

## Importação de eventos (v1.8, RN-16)

Cada linha não aceita mostra a mensagem pelo código do erro; em `dados_invalidos`, pelo motivo.

| Código / motivo | Mensagem |
| --- | --- |
| `cliente_nao_encontrado` | Cliente não encontrado. Confira o ID. |
| `cliente_inativo` | Cadastro inativo. Eventos ficam bloqueados. |
| `consentimento_invalido` | Sem consentimento válido. O evento não pode ser usado. |
| `dados_invalidos` / `cliente_id` | Informe o ID do cliente no formato UUID. |
| `dados_invalidos` / `tipo` | Use compra, navegacao ou interacao. |
| `dados_invalidos` / `origem` | Informe de onde veio o evento. |
| `dados_invalidos` / `timestamp` | Informe data, hora e fuso, por exemplo 2026-11-10T14:30:00-03:00. |
| `dados_invalidos` / `timestamp_futuro` | A data não pode estar no futuro. |
| `dados_invalidos` / `valor` | O valor deve estar entre 0 e 1000000, com ponto decimal. |
| `dados_invalidos` / `valor_compra` | Uma compra precisa ter valor maior que zero. |
| `dados_invalidos` / `colunas` | A linha não tem as mesmas colunas do cabeçalho. |
| duplicada | Já registrado: mesmo cliente, tipo, data e valor. |

Arquivo recusado inteiro:

| Motivo | Mensagem |
| --- | --- |
| `arquivo_vazio` | O arquivo não tem nenhuma linha de evento. Confira se ele tem o cabeçalho e pelo menos uma linha. |
| `arquivo_grande` | O arquivo passa de 256 KiB. Divida em arquivos menores. |
| `linhas_demais` | O arquivo passa de 1.000 linhas. Divida em arquivos menores. |
| `codificacao_invalida` | O arquivo não está em UTF-8. Salve de novo como "CSV UTF-8". |
| `cabecalho_invalido` | O cabeçalho deve ter as colunas cliente_id, tipo, origem, timestamp e valor (categoria é opcional), separadas por vírgula. |
| `formato_invalido` | O arquivo tem aspas mal fechadas. Confira o arquivo e envie de novo. |

## Confirmações

| Ação | Título do diálogo | Corpo | Botões |
| --- | --- | --- | --- |
| Conceder | Conceder consentimento | A partir de agora, os eventos deste cliente podem ser usados para perfil, segmento e recomendação, com a finalidade "{finalidade}" e o termo {versão}. | Conceder (primário) · Cancelar |
| Revogar | Revogar consentimento | O processamento dos dados deste cliente para imediatamente. Os dados já registrados continuam guardados. Você pode conceder de novo depois. | Revogar consentimento (danger) · Cancelar (foco inicial) |
| Importar eventos | (sem diálogo: a prévia já é a confirmação) | {n} eventos prontos para importar. Linhas duplicadas e com erro ficam de fora. | Importar {n} eventos (primário) · Enviar outro arquivo |
| Expirar | Expirar recomendações vencidas | Recomendações ativas com validade até agora passam a Expirada e deixam de ser exibidas. | Expirar agora (primário) · Cancelar |

## Toasts

"Cliente cadastrado" · "Consentimento concedido" · "Consentimento revogado" · "Evento registrado" · "Processamento concluído" · "Processamento concluído sem recomendação" · "{n} recomendações expiradas" / "Nenhuma recomendação vencida" · "ID copiado".

## Formatos

| Dado | Formato | Exemplo |
| --- | --- | --- |
| Data e hora | `dd/mm/aaaa HH:mm`, fuso America/Sao_Paulo | 11/11/2026 22:07 |
| Só data | `dd/mm/aaaa` | 11/12/2026 |
| Relativo (metadado) | "há 3 minutos", "ontem às 14:10"; o absoluto vai no `title` | — |
| Dinheiro | `R$ 1.234,56` | R$ 610,00 |
| Score | duas casas com vírgula + explicação | 0,40 · 4 de 10 eventos em 30 dias |
| ID | 8 primeiros + "…" + 4 últimos, em mono, com cópia | 6200ae2f…a4c9 |
| E-mail mascarado | primeira letra + `***` + domínio | c***@exemplo.test |
| Critério | campos com nome legível e operador matemático | Ticket médio ≥ 500 · Compras ≥ 3 |

Nomes legíveis dos campos de critério: `frequenciaCompras` = Compras, `ticketMedio` = Ticket médio, `scoreEngajamento` = Engajamento.
