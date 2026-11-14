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
| cliente com `DataEliminacao` (v1.9) | Eliminado | critical, círculo com barra |
| pedido `ABERTO` (v1.10) | Aberto | warn, triângulo |
| pedido `ATENDIDO` (v1.10) | Atendido | ok, círculo com visto |
| tipo `compra` / `navegacao` / `interacao` | Compra / Navegação / Interação | — |

## Códigos de erro da API → mensagem

O código estável (`errs.Codigo`) define a mensagem; o texto técnico do servidor nunca aparece.

| Código | Onde aparece | Mensagem |
| --- | --- | --- |
| `nao_autenticado` | Entrar | Usuário ou senha incorretos. |
| `cliente_nao_encontrado` | Busca, ficha | Não encontramos esse cliente. Confira o ID ou busque pelo nome. |
| `cliente_inativo` | Ficha | Este cadastro está inativo. Eventos e processamento ficam bloqueados enquanto ele estiver assim. |
| `cliente_eliminado` (v1.9) | Ficha, qualquer ação | Os dados deste titular foram eliminados. Nada mais pode ser alterado. |
| `transicao_invalida` (eliminar) | Direitos do titular | Revogue o consentimento antes de eliminar os dados. |
| `consentimento_invalido` | Ficha, eventos, processamento | Sem consentimento válido. Registre o consentimento do titular para usar os dados de comportamento. |
| `transicao_invalida` (conceder) | Consentimento | O consentimento já está concedido. Para mudar a versão do termo, revogue e conceda de novo. |
| `transicao_invalida` (revogar) | Consentimento | Não há consentimento concedido para revogar. |
| `dados_invalidos` | Formulários | Mensagem por campo (abaixo). |
| `perfil_nao_encontrado` | Visão geral | Este cliente ainda não foi processado. Use "Processar agora". |
| `sem_produto_elegivel` | Recomendação (não é erro) | Nenhum produto ativo na categoria preferida ({categoria}). |
| `erro_interno` | Qualquer | Algo falhou do nosso lado. As etapas concluídas antes da falha ficaram salvas. Tente de novo; se continuar, informe o código {X-Request-ID} à equipe técnica. |
| `acesso_invalido` (v1.10) | Portal: abrir o link | Este link não vale mais. Peça um novo. |
| `transicao_invalida` (pedir eliminação, v1.10) | Portal | Para pedir a eliminação, retire a autorização antes. |
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
| Nome do cliente (confirmar eliminação) | igual ao nome do cadastro | O nome digitado não confere com o cadastro. |

Valide no navegador (atributos HTML: `required`, `min`, `max`, `step`) para dar retorno imediato, e sempre de novo no servidor, que é quem decide.

## Confirmações

| Ação | Título do diálogo | Corpo | Botões |
| --- | --- | --- | --- |
| Conceder | Conceder consentimento | A partir de agora, os eventos deste cliente podem ser usados para perfil, segmento e recomendação, com a finalidade "{finalidade}" e o termo {versão}. | Conceder (primário) · Cancelar |
| Revogar | Revogar consentimento | O processamento dos dados deste cliente para imediatamente. Os dados já registrados continuam guardados. Você pode conceder de novo depois. | Revogar consentimento (danger) · Cancelar (foco inicial) |
| Expirar | Expirar recomendações vencidas | Recomendações ativas com validade até agora passam a Expirada e deixam de ser exibidas. | Expirar agora (primário) · Cancelar |
| Eliminar (v1.9) | Eliminar dados do titular | Eventos, perfil e recomendações deste cliente são apagados e o cadastro é anonimizado. O histórico do consentimento fica guardado como prova, sem nome nem e-mail. Não pode ser desfeito. Digite o nome do cliente para confirmar: {nome}. | Eliminar dados (danger) · Cancelar (foco inicial) |
| Retenção (v1.9) | Aplicar retenção | Eventos, perfil e recomendações de quem revogou o consentimento há {N} dias ou mais são apagados. Cadastro e histórico do consentimento ficam. Não pode ser desfeito. | Aplicar retenção (danger) · Cancelar (foco inicial) |

## Toasts

"Cliente cadastrado" · "Consentimento concedido" · "Consentimento revogado" · "Evento registrado" · "Processamento concluído" · "Processamento concluído sem recomendação" · "{n} recomendações expiradas" / "Nenhuma recomendação vencida" · "ID copiado" · (v1.9) "Cadastro corrigido" · "Dados eliminados" · "Retenção aplicada a {n} clientes" / "Retenção aplicada a 1 cliente" / "Nenhum cliente com a retenção vencida".

## Direitos do titular (v1.9)

| Painel | Texto |
| --- | --- |
| Corrigir cadastro | Corrige nome e e-mail a pedido do titular. Não depende do consentimento. |
| Acesso e portabilidade | Tudo o que o Perfila guarda deste titular: cadastro, consentimento e histórico, eventos, perfil, segmento e recomendações, num arquivo JSON. Vale com qualquer estado do consentimento. |
| Eliminação | Apaga eventos, perfil e recomendações e anonimiza o cadastro. O histórico do consentimento fica guardado como prova. Não pode ser desfeito. |
| Cliente eliminado | Dados eliminados em {data}. O cadastro foi anonimizado; resta o histórico do consentimento, guardado como prova. |
| Retenção (Operação) | Apaga eventos, perfil e recomendações de quem revogou o consentimento há {N} dias ou mais. Cadastro e histórico do consentimento ficam. É o mesmo efeito do comando `aplicar-retencao`. |

## Portal do titular (v1.10)

Fala com a própria pessoa, em segunda pessoa e sem jargão: "autorização" nas ações, "consentimento" só no nome do estado.

| Onde | Texto |
| --- | --- |
| Entrar (título) | Acesse seus dados |
| Entrar (explicação) | Informe o e-mail do seu cadastro. Enviamos um link que vale por 30 minutos e só pode ser usado uma vez. |
| E-mail vazio ou sem arroba | Informe o e-mail do seu cadastro. |
| Depois de enviar (sempre igual) | Se o e-mail estiver cadastrado, enviamos um link de acesso. Ele vale por 30 minutos e só pode ser usado uma vez. |
| Muitas tentativas | Muitas tentativas. Aguarde 15 minutos. |
| Abrir o link | Para entrar, use o botão abaixo. O link vale uma vez só. · Entrar no portal |
| Sessão terminou | Sua sessão terminou. Peça um novo link para continuar. |
| Finalidade `segmentacao` | segmentação: entender seu comportamento de compra para recomendar produtos |
| Origem `portal` na linha do tempo | por você, no portal |
| Com Concedido, sem eliminar | Para pedir a eliminação, retire a autorização antes. |
| Pedido aberto | Você pediu a eliminação em {data}. A equipe vai atender o pedido. |
| Baixar dados | Uma cópia de tudo o que o Perfila guarda sobre você, num arquivo JSON. |
| Pedir eliminação (painel) | Peça para apagarmos seus dados de comportamento e anonimizarmos seu cadastro. O registro das suas autorizações fica guardado como prova. |

| Ação | Título | Corpo | Botões |
| --- | --- | --- | --- |
| Retirar autorização | Retirar sua autorização | A partir de agora, seus dados de comportamento deixam de ser usados. Os dados já registrados continuam guardados até você pedir a eliminação ou até o prazo de retenção. Você pode autorizar de novo depois. | Retirar autorização (danger) · Cancelar (foco inicial) |
| Autorizar de novo | Autorizar o uso dos seus dados | Você autoriza o uso dos seus dados de comportamento para {finalidade em palavras}, conforme o termo {versão}. Você pode retirar a autorização quando quiser. | Autorizar (primário) · Cancelar |
| Pedir eliminação | Pedir a eliminação dos seus dados | A equipe vai apagar seus eventos, seu perfil e suas recomendações e anonimizar seu cadastro. O registro das suas autorizações fica guardado como prova, sem seu nome nem seu e-mail. Depois de atendido, não pode ser desfeito. | Pedir eliminação (danger) · Cancelar (foco inicial) |

Toasts do portal: "Autorização retirada" · "Autorização registrada" · "Pedido de eliminação registrado".

Mensagem da caixa de saída simulada: assunto "Seu link de acesso ao Perfila"; texto "Olá. Recebemos um pedido de acesso aos dados do cadastro feito em {data do cadastro}. Use este link até {hora de expiração}: {link}. Ele só pode ser usado uma vez. Se não foi você, ignore esta mensagem."

No painel: "O titular pediu a eliminação pelo portal em {data}." (aba Direitos do titular) · "Nenhum pedido do titular." · "Nenhuma mensagem nas últimas 24 horas." · "Simulação para a demonstração: nenhum e-mail sai do sistema. Mensagens com mais de 24 horas são apagadas."

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
