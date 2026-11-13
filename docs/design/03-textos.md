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
| `sem_permissao` | Qualquer (403) | Seu perfil de acesso não permite abrir esta página. Se precisar dela, fale com quem administra o painel. |
| `login_em_uso` | Criar operador | Já existe um operador com esse login. Logins não se repetem, nem de operadores desativados. |
| `propria_conta` | Operador | Você não pode mudar o próprio papel nem se desativar. Peça a outra pessoa administradora. |
| `operador_nao_encontrado` | Usuários | Não encontramos esse operador. Volte para a lista de usuários. |
| `nao_autenticado` (Minha senha) | Minha senha | Senha atual incorreta. |
| sessão encerrada | Entrar | Sua sessão foi encerrada. Entre de novo. |

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

Segmentos e produtos (v1.6, RN-14): as mensagens vêm do domínio (`errs.ErroCampo`) e aparecem ao lado do campo.

| Campo | Regra | Mensagem |
| --- | --- | --- |
| Nome do segmento | 1 a 40, sem `;` nem `:`, único | Informe o nome do segmento. · Use até 40 caracteres no nome. · O nome não pode ter ponto e vírgula nem dois-pontos. · Já existe o segmento "{nome}". |
| Prioridade | 1 a 999, única entre os ativos, menor que a do padrão | A prioridade vai de 1 a 999. · A prioridade {n} já é do segmento "{nome}". · A prioridade precisa ser menor que a do segmento "padrao" ({n}). |
| Critério | gramática da RN-06; só o padrão é vazio | A condição "{condição}" está fora da gramática: use campo, operador e número com ponto. · Campo desconhecido na condição "{condição}": use frequenciaCompras, ticketMedio ou scoreEngajamento. · Condição vazia antes ou depois de "&&". · Complete a condição {n}: campo, operador e valor. · Só o segmento padrão pode ter o critério vazio. Monte pelo menos uma condição. |
| Situação | o padrão fica ativo | O segmento "padrao" não pode ser desativado. |
| Nome do produto | 1 a 100 | Informe o nome do produto. |
| Categoria | 1 a 40 | Informe a categoria. |
| Preço | 0 a 1.000.000 | O preço vai de 0 a 1.000.000. · Informe o preço com números, por exemplo 90,50. |

Valide no navegador (atributos HTML: `required`, `min`, `max`, `step`) para dar retorno imediato, e sempre de novo no servidor, que é quem decide.

## Confirmações

| Ação | Título do diálogo | Corpo | Botões |
| --- | --- | --- | --- |
| Conceder | Conceder consentimento | A partir de agora, os eventos deste cliente podem ser usados para perfil, segmento e recomendação, com a finalidade "{finalidade}" e o termo {versão}. | Conceder (primário) · Cancelar |
| Revogar | Revogar consentimento | O processamento dos dados deste cliente para imediatamente. Os dados já registrados continuam guardados. Você pode conceder de novo depois. | Revogar consentimento (danger) · Cancelar (foco inicial) |
| Expirar | Expirar recomendações vencidas | Recomendações ativas com validade até agora passam a Expirada e deixam de ser exibidas. | Expirar agora (primário) · Cancelar |
| Desativar operador | Desativar {login} | O acesso ao painel acaba agora, inclusive nas sessões abertas. Você pode ativar de novo depois. | Desativar (danger) · Cancelar (foco inicial) |
| Salvar segmento (v1.6) | Prévia de {nome} | A prévia considera {n} clientes ativos com consentimento válido. Quem já foi segmentado muda só no próximo processamento. | Salvar alterações (primário) · Voltar ao formulário |
| Desativar segmento | Desativar {nome} | A mesma prévia, com a contagem sem o segmento. | Desativar (danger) · Cancelar |
| Desativar produto | Desativar {nome} | O produto sai da recomendação no próximo processamento de cada cliente. Você pode ativar de novo depois. | Desativar (danger) · Cancelar (foco inicial) |

## Toasts

"Cliente cadastrado" · "Consentimento concedido" · "Consentimento revogado" · "Evento registrado" · "Processamento concluído" · "Processamento concluído sem recomendação" · "{n} recomendações expiradas" / "Nenhuma recomendação vencida" · "ID copiado" · "Operador criado" · "Papel alterado" · "Operador desativado" · "Operador ativado" · "Senha provisória definida" · "Senha trocada" (v1.5) · "Segmento salvo. Os clientes mudam de segmento no próximo processamento." · "Segmento ativado" · "Segmento desativado" · "Produto salvo" · "Produto ativado" · "Produto desativado" · "Nada mudou" (v1.6).

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
