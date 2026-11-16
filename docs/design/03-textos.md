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
| finalidade `segmentacao` / `publicidade_terceiros` (v1.11) | Segmentação / Publicidade com terceiros | — |
| exportação `INCLUSAO` / `REMOCAO` (v1.11) | Inclusão / Remoção | neutral / warn |
| membro `REMOCAO_PENDENTE` (v1.11) | Remoção pendente | warn, triângulo |

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
| `arquivo_indisponivel` (v1.11) | Públicos, download | Este arquivo não está mais disponível: ele fica guardado por {horas} horas. Gere um novo, se precisar. |
| `dados_invalidos` (gerar público) (v1.11) | Públicos | Nenhum cliente entra neste público. Nada foi gerado. |
| `dados_invalidos` (gerar remoção) (v1.11) | Públicos | Não há remoção pendente neste segmento. Nada foi gerado. |
| `erro_interno` | Qualquer | Algo falhou do nosso lado. As etapas concluídas antes da falha ficaram salvas. Tente de novo; se continuar, informe o código {X-Request-ID} à equipe técnica. |
| `acesso_invalido` (v1.10) | Portal: abrir o link | Este link não vale mais. Peça um novo. |
| `transicao_invalida` (pedir eliminação, v1.10) | Portal | Para pedir a eliminação, retire a autorização antes. |
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
| Nome do cliente (confirmar eliminação) | igual ao nome do cadastro | O nome digitado não confere com o cadastro. |

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
| Conceder publicidade (v1.11) | Conceder publicidade com terceiros | A partir de agora, este cliente pode entrar nos públicos enviados a plataformas de anúncio, só pelo e-mail em hash, com o termo {versão}. A Segmentação não muda. | Conceder (primário) · Cancelar |
| Revogar publicidade (v1.11) | Revogar publicidade com terceiros | O cliente sai dos próximos públicos e fica pendente de remoção nos que já foram gerados. A Segmentação não muda. | Revogar publicidade (danger) · Cancelar (foco inicial) |
| Gerar público (v1.11) | Gerar arquivo do público | Vai ser gerado um arquivo com {n} clientes do segmento {segmento}, para subir na plataforma de anúncios. Ele leva só o e-mail em hash (SHA-256), que continua sendo dado pessoal, e fica disponível por {horas} horas. A exportação fica registrada com o seu login. | Gerar arquivo (primário) · Cancelar |
| Gerar remoção (v1.11) | Gerar arquivo de remoção | Vai ser gerado um arquivo com {n} clientes que precisam sair do público {segmento} na plataforma de anúncios. Suba o arquivo como remoção de público. | Gerar arquivo de remoção (primário) · Cancelar |
| Expirar | Expirar recomendações vencidas | Recomendações ativas com validade até agora passam a Expirada e deixam de ser exibidas. | Expirar agora (primário) · Cancelar |
| Desativar operador | Desativar {login} | O acesso ao painel acaba agora, inclusive nas sessões abertas. Você pode ativar de novo depois. | Desativar (danger) · Cancelar (foco inicial) |
| Salvar segmento (v1.6) | Prévia de {nome} | A prévia considera {n} clientes ativos com consentimento válido. Quem já foi segmentado muda só no próximo processamento. | Salvar alterações (primário) · Voltar ao formulário |
| Desativar segmento | Desativar {nome} | A mesma prévia, com a contagem sem o segmento. | Desativar (danger) · Cancelar |
| Desativar produto | Desativar {nome} | O produto sai da recomendação no próximo processamento de cada cliente. Você pode ativar de novo depois. | Desativar (danger) · Cancelar (foco inicial) |
| Eliminar (v1.9) | Eliminar dados do titular | Eventos, perfil e recomendações deste cliente são apagados e o cadastro é anonimizado. O histórico do consentimento fica guardado como prova, sem nome nem e-mail. Não pode ser desfeito. Digite o nome do cliente para confirmar: {nome}. | Eliminar dados (danger) · Cancelar (foco inicial) |
| Retenção (v1.9) | Aplicar retenção | Eventos, perfil e recomendações de quem revogou o consentimento há {N} dias ou mais são apagados. Cadastro e histórico do consentimento ficam. Não pode ser desfeito. | Aplicar retenção (danger) · Cancelar (foco inicial) |

## Toasts

"Cliente cadastrado" · "Consentimento concedido" · "Consentimento revogado" · "Arquivo gerado" (v1.11) · "Evento registrado" · "Processamento concluído" · "Processamento concluído sem recomendação" · "{n} recomendações expiradas" / "Nenhuma recomendação vencida" · "ID copiado" · "Operador criado" · "Papel alterado" · "Operador desativado" · "Operador ativado" · "Senha provisória definida" · "Senha trocada" (v1.5) · "Segmento salvo. Os clientes mudam de segmento no próximo processamento." · "Segmento ativado" · "Segmento desativado" · "Produto salvo" · "Produto ativado" · "Produto desativado" · "Nada mudou" (v1.6) · (v1.9) "Cadastro corrigido" · "Dados eliminados" · "Retenção aplicada a {n} clientes" / "Retenção aplicada a 1 cliente" / "Nenhum cliente com a retenção vencida".

## Direitos do titular (v1.9)

| Painel | Texto |
| --- | --- |
| Corrigir cadastro | Corrige nome e e-mail a pedido do titular. Não depende do consentimento. |
| Acesso e portabilidade | Tudo o que o Perfila guarda deste titular: cadastro, consentimento e histórico, eventos, perfil, segmento e recomendações, num arquivo JSON. Vale com qualquer estado do consentimento. |
| Eliminação | Apaga eventos, perfil e recomendações e anonimiza o cadastro. O histórico do consentimento fica guardado como prova. Não pode ser desfeito. Quem não é administrador vê, no lugar do botão: "Só o administrador elimina dados." (RN-11) |
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

Mensagem da caixa de saída simulada: assunto "Seu link de acesso ao Perfila"; texto "Olá. Recebemos um pedido de acesso aos dados do cadastro feito em {data do cadastro}. Use o link abaixo até {hora de expiração}; ele só pode ser usado uma vez. Se não foi você, ignore esta mensagem. {link}" (o link por último, sem pontuação depois dele)

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
