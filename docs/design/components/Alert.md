# Alert

Aviso dentro da página, ligado ao contexto: bloqueio, falha, atenção ou informação.

**Variantes.** `pf-alert` (padrão, crítico) para bloqueio e falha (sem consentimento, processamento interrompido, banco fora do ar); `--warn` para atenção sem bloqueio (sem produto elegível, "Revogue o consentimento antes de eliminar os dados"); `--ok` para confirmação persistente ("Importação concluída"); `--info` (violeta suave) para orientação ("Quem entra no público"); `--neutral` para estado encerrado ("Sem consentimento válido" antes da primeira concessão, "Dados eliminados").

**O que fornecer.** Título curto em `pf-alert__title` (display, na cor do estado), texto que diz o que fazer em `pf-alert__text` e, se houver, a ação logo abaixo (botão pequeno). Sem ícone: a cor do título e do fundo carrega o estado, e a palavra do título diz qual é. `role="alert"` só para erros que acabaram de acontecer; avisos de estado ficam com `role="status"`.

**Forma.** Raio `radius-alert`, borda de 1px em 25% da cor do estado, fundo `*-tint`. Dois avisos, ou um aviso e um painel, empilhados recebem `space-5` entre si.

**Não faça.** Usar Alert para confirmar ação concluída (isso é Toast). Empilhar mais de dois.
