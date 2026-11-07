#!/usr/bin/env bash
# Demonstração da API (fase Support) contra um omega-server rodando: make serve
# Uso: scripts/demo-api.sh [URL] [TOKEN]
set -euo pipefail
URL="${1:-http://localhost:8080}"
TOKEN="${2:-dev-token}"
H=(-sS -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json")
passo() { printf '\n\033[1m[%s] %s\033[0m\n' "$1" "$2"; }
json() { if command -v jq >/dev/null; then jq .; else cat; echo; fi; }

passo "GET /saude" "sem token"
curl -sS "$URL/saude" | json

passo "UC-09" "POST /v1/clientes"
ID=$(curl "${H[@]}" -X POST "$URL/v1/clientes" -d '{"nome":"Cliente Sintético","email":"c-001@exemplo.test"}' \
  | sed -E 's/.*"id":"([^"]+)".*/\1/')
echo "clienteId = $ID"
C="$URL/v1/clientes/$ID"

passo "RN-01" "processar SEM consentimento → 403"
curl "${H[@]}" -o /dev/stderr -w '%{http_code}\n' -X POST "$C/processamento" 2>/dev/null || true

passo "UC-08" "PUT /consentimento"
curl "${H[@]}" -X PUT "$C/consentimento" -d '{"finalidade":"segmentacao","versaoTermo":"v1"}' | json

passo "UC-01" "POST /eventos (3 compras em livros + 1 navegação)"
QUANDO=$(date -u -d '1 hour ago' +%Y-%m-%dT%H:%M:%SZ 2>/dev/null || date -u -v-1H +%Y-%m-%dT%H:%M:%SZ)
for V in 520 610 700; do
  curl "${H[@]}" -o /dev/null -w "compra $V → %{http_code}\n" -X POST "$C/eventos" \
    -d "{\"tipo\":\"compra\",\"origem\":\"ecommerce\",\"valor\":$V,\"categoria\":\"livros\",\"timestamp\":\"$QUANDO\"}"
done
curl "${H[@]}" -o /dev/null -w "navegacao → %{http_code}\n" -X POST "$C/eventos" \
  -d "{\"tipo\":\"navegacao\",\"origem\":\"web\",\"valor\":0,\"timestamp\":\"$QUANDO\"}"

passo "RN-09" "POST /processamento (UC-02 → UC-03 → UC-04)"
curl "${H[@]}" -X POST "$C/processamento" | json

passo "UC-08" "revogar e tentar de novo → 403"
curl "${H[@]}" -o /dev/null -X POST "$C/consentimento/revogacao"
curl "${H[@]}" -X POST "$C/processamento" | json
