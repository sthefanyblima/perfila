#!/bin/sh
# Gera um certificado autoassinado na primeira subida, se nenhum foi montado (Q-12).
# Para usar um certificado real, monte omega.crt e omega.key em /etc/nginx/certs.
set -eu
dir=/etc/nginx/certs
if [ ! -s "$dir/omega.crt" ] || [ ! -s "$dir/omega.key" ]; then
    mkdir -p "$dir"
    openssl req -x509 -nodes -newkey rsa:2048 -days 365 \
        -subj "/CN=localhost/O=Equipe Omega" \
        -addext "subjectAltName=DNS:localhost,IP:127.0.0.1" \
        -keyout "$dir/omega.key" -out "$dir/omega.crt" 2>/dev/null
    chmod 600 "$dir/omega.key"
    echo "10-certificado.sh: certificado autoassinado gerado em $dir"
fi
