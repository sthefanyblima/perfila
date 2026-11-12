# Alert

Aviso dentro da página, ligado ao contexto: bloqueio, falha, atenção ou informação.

**Variantes.** `--critical` para bloqueio e falha (sem consentimento, processamento interrompido, banco fora do ar); `--warn` para atenção sem bloqueio (sem produto elegível); `--ok` para confirmação persistente; `--info` em anil para orientação ("Segmentos entram por migração.").

**O que fornecer.** Ícone, título curto em `pf-alert__title`, texto que diz o que fazer em `pf-alert__text` e, se houver, uma ação à direita. `role="alert"` só para erros que acabaram de acontecer; avisos de estado ficam sem `role`.

**Não faça.** Usar Alert para confirmar ação concluída (isso é Toast). Empilhar mais de dois.
