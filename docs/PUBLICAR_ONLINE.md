# Publicar a demonstração online (Render, plano gratuito)

Modo memória com dados sintéticos (D-23): sem banco, sem custo. Cada reinício volta ao estado inicial.

1. No Render, **New → Blueprint**, conectar este repositório (o `render.yaml` da raiz descreve o serviço).
2. Em **Environment**, preencher os cinco segredos (`AUTH_TOKEN`, `PAINEL_SENHA_HASH`, `PAINEL_CHAVE_SESSAO`,
   `PORTAL_CHAVE_SESSAO`, `PORTAL_URL_BASE`). Os valores ficam fora do repositório. Para gerar a senha do
   administrador: `printf '%s' 'SUA-SENHA-COM-12-OU-MAIS' | go run ./cmd/omega-server hash-senha`.
3. `PORTAL_URL_BASE` é o endereço público que o Render atribuir (`https://….onrender.com`, sem barra no fim);
   depois de definir, o serviço reinicia sozinho.
4. O plano gratuito dorme após ~15 minutos sem acesso e a primeira visita leva cerca de 1 minuto: abra a página
   alguns minutos antes da apresentação.

Atenção: no modo memória entram também os operadores de exemplo `analista` e `atendimento` (Q-41), com as senhas
de desenvolvimento do `Makefile`, que são públicas. Aceitável só porque os dados são sintéticos.
