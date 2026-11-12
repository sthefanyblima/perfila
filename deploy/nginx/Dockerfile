# Balanceador do UML8 (T-24). 1.28 traz o parâmetro "resolve" no upstream (open source),
# que acompanha réplicas paradas ou recriadas pelo DNS do Docker.
FROM nginx:1.28-alpine
RUN apk add --no-cache openssl
COPY nginx.conf /etc/nginx/nginx.conf
COPY 10-certificado.sh /docker-entrypoint.d/10-certificado.sh
RUN chmod +x /docker-entrypoint.d/10-certificado.sh
