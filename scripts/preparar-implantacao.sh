#!/usr/bin/env sh
# Prepara os segredos da implantação (Q-14), sem sobrescrever o que já existe:
#   deploy/.env        OMEGA_DB_SENHA (lida pelo docker compose)
#   deploy/config.env  config.env do UML8, com AUTH_TOKEN aleatório
# Os dois arquivos são ignorados pelo git. Spec: 3.8, T-24.
set -eu
cd "$(dirname "$0")/../deploy"

aleatorio() { head -c 24 /dev/urandom | od -An -tx1 | tr -d ' \n'; }

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
