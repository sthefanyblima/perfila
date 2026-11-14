#!/usr/bin/env sh
# Prepara os segredos da implantação (Q-14), sem sobrescrever o que já existe:
#   deploy/.env        OMEGA_DB_SENHA (lida pelo docker compose)
#   deploy/config.env  config.env do UML8, com AUTH_TOKEN aleatório
#   painel (Q-28)      PAINEL_SENHA aleatória em deploy/.env; hash bcrypt (omega-server hash-senha)
#                      e PAINEL_CHAVE_SESSAO em deploy/config.env
#   portal (Q-60)      PORTAL_URL_BASE (o endereço do Nginx) e PORTAL_CHAVE_SESSAO aleatória em deploy/config.env
# Os dois arquivos são ignorados pelo git. Spec: 3.8, T-24.
set -eu
cd "$(dirname "$0")/../deploy"

aleatorio() { head -c 24 /dev/urandom | od -An -tx1 | tr -d ' \n'; }
chave() { head -c 32 /dev/urandom | od -An -tx1 | tr -d ' \n'; }

if [ ! -f .env ]; then
    printf 'OMEGA_DB_SENHA=%s\n' "$(aleatorio)" > .env
    echo "deploy/.env criado (senha do banco aleatória)"
fi

if [ ! -f config.env ]; then
    # DATABASE_URL fica vazia aqui: o compose a define a partir de deploy/.env.
    sed -e "s/^AUTH_TOKEN=.*/AUTH_TOKEN=$(aleatorio)/" \
        -e "s/^ENGAJAMENTO_META_EVENTOS=.*/ENGAJAMENTO_META_EVENTOS=10/" \
        config.env.example > config.env
    echo "deploy/config.env criado (AUTH_TOKEN aleatório; consulte com: grep AUTH_TOKEN deploy/config.env)"
fi

# Painel do operador (Q-28, Q-29): só acrescenta o que falta, sem tocar no que já existe.
if ! grep -q '^PAINEL_SENHA=' .env; then
    printf 'PAINEL_SENHA=%s\n' "$(aleatorio | cut -c1-20)" >> .env
    echo "senha do painel criada (usuário operador; consulte com: grep PAINEL_SENHA deploy/.env)"
fi
if ! grep -q '^PAINEL_SENHA_HASH=.' config.env; then
    sed -i.bak '/^PAINEL_USUARIO=/d; /^PAINEL_SENHA_HASH=/d; /^PAINEL_CHAVE_SESSAO=/d' config.env && rm -f config.env.bak
    docker compose -f docker-compose.yml build app1 >/dev/null
    hash=$(sed -n 's/^PAINEL_SENHA=//p' .env | docker run --rm -i --entrypoint omega-server perfila/omega-server:local hash-senha)
    printf 'PAINEL_USUARIO=operador\nPAINEL_SENHA_HASH=%s\nPAINEL_CHAVE_SESSAO=%s\n' "$hash" "$(chave)" >> config.env
    echo "painel configurado em deploy/config.env (hash bcrypt e chave de sessão aleatória)"
fi

# Portal do titular (Q-60): endereço público do Nginx no link e chave própria, diferente da do painel.
if ! grep -q '^PORTAL_CHAVE_SESSAO=.' config.env; then
    sed -i.bak '/^PORTAL_URL_BASE=/d; /^PORTAL_CHAVE_SESSAO=/d' config.env && rm -f config.env.bak
    printf 'PORTAL_URL_BASE=https://localhost:%s\nPORTAL_CHAVE_SESSAO=%s\n' "${PERFILA_PORTA_HTTPS:-8443}" "$(chave)" >> config.env
    echo "portal do titular configurado em deploy/config.env (endereço do Nginx e chave de sessão aleatória)"
fi
