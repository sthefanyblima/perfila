# omega-server (UML8): binário Go estático em contêiner. Spec: 3.8, T-24.
# Contexto de build: raiz do repositório (docker compose usa context: ..).

# ---- build ----
FROM golang:1.22-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY cmd ./cmd
COPY internal ./internal
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/omega-server ./cmd/omega-server

# ---- execução ----
# Alpine só pelo wget do healthcheck; o binário não depende de nada da imagem.
FROM alpine:3.20
RUN adduser -D -H -u 10001 omega
COPY --from=build /out/omega-server /usr/local/bin/omega-server
USER omega
EXPOSE 8080
HEALTHCHECK --interval=5s --timeout=3s --start-period=10s --retries=10 \
    CMD wget -qO- http://127.0.0.1:8080/saude >/dev/null || exit 1
ENTRYPOINT ["omega-server", "-config", "/etc/omega/config.env"]
CMD ["serve"]
