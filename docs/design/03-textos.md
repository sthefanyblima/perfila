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
| finalidade `segmentacao` / `publicidade_terceiros` (v1.9) | Segmentação / Publicidade com terceiros | — |
| exportação `INCLUSAO` / `REMOCAO` (v1.9) | Inclusão / Remoção | neutral / warn |
| membro `REMOCAO_PENDENTE` (v1.9) | Remoção pendente | warn, triângulo |

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
| `arquivo_indisponivel` (v1.9) | Públicos, download | Este arquivo não está mais disponível: ele fica guardado por {horas} horas. Gere um novo, se precisar. |
| `dados_invalidos` (gerar público) (v1.9) | Públicos | Nenhum cliente entra neste público. Nada foi gerado. |
| `dados_invalidos` (gerar remoção) (v1.9) | Públicos | Não há remoção pendente neste segmento. Nada foi gerado. |
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

## Confirmações

| Ação | Título do diálogo | Corpo | Botões |
| --- | --- | --- | --- |
| Conceder | Conceder consentimento | A partir de agora, os eventos deste cliente podem ser usados para perfil, segmento e recomendação, com a finalidade "{finalidade}" e o termo {versão}. | Conceder (primário) · Cancelar |
| Revogar | Revogar consentimento | O processamento dos dados deste cliente para imediatamente. Os dados já registrados continuam guardados. Você pode conceder de novo depois. | Revogar consentimento (danger) · Cancelar (foco inicial) |
| Conceder publicidade (v1.9) | Conceder publicidade com terceiros | A partir de agora, este cliente pode entrar nos públicos enviados a plataformas de anúncio, só pelo e-mail em hash, com o termo {versão}. A Segmentação não muda. | Conceder (primário) · Cancelar |
| Revogar publicidade (v1.9) | Revogar publicidade com terceiros | O cliente sai dos próximos públicos e fica pendente de remoção nos que já foram gerados. A Segmentação não muda. | Revogar publicidade (danger) · Cancelar (foco inicial) |
| Gerar público (v1.9) | Gerar arquivo do público | Vai ser gerado um arquivo com {n} clientes do segmento {segmento}, para subir na plataforma de anúncios. Ele leva só o e-mail em hash (SHA-256), que continua sendo dado pessoal, e fica disponível por {horas} horas. A exportação fica registrada com o seu login. | Gerar arquivo (primário) · Cancelar |
| Gerar remoção (v1.9) | Gerar arquivo de remoção | Vai ser gerado um arquivo com {n} clientes que precisam sair do público {segmento} na plataforma de anúncios. Suba o arquivo como remoção de público. | Gerar arquivo de remoção (primário) · Cancelar |
| Expirar | Expirar recomendações vencidas | Recomendações ativas com validade até agora passam a Expirada e deixam de ser exibidas. | Expirar agora (primário) · Cancelar |

## Toasts

"Cliente cadastrado" · "Consentimento concedido" · "Consentimento revogado" · "Arquivo gerado" (v1.9) · "Evento registrado" · "Processamento concluído" · "Processamento concluído sem recomendação" · "{n} recomendações expiradas" / "Nenhuma recomendação vencida" · "ID copiado".

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
