#!/usr/bin/env bash
set -euo pipefail

BASE_URL="http://localhost:8080"

# === Config utilisateur par défaut ===
email="test@test.com"
password="secret"

# 1) Reset
echo "==> Reset DB"
curl -sS -X POST "$BASE_URL/admin/reset" -i
echo

# 2) Créer un user
echo "==> Create user"
curl -sS -X POST "$BASE_URL/api/users" \
  -H 'Content-Type: application/json' \
  -d "{\"email\":\"$email\",\"password\":\"$password\"}" -i
echo

# 3) Login initial (JSON pur)
echo "==> Login (before upgrade)"
resp="$(curl -sS -X POST "$BASE_URL/api/login" \
  -H 'Content-Type: application/json' \
  -d "{\"email\":\"$email\",\"password\":\"$password\"}")"

echo "$resp" | jq .
token="$(echo "$resp" | jq -r '.token')"
user_id="$(echo "$resp" | jq -r '.id')"

echo "   Got user_id=$user_id"
echo "   Got token=$token"
echo

# 4) Envoyer le webhook
echo "==> Send webhook"
curl -sS -X POST "$BASE_URL/api/polka/webhooks" \
  -H 'Content-Type: application/json' \
  -H "Authorization: Bearer $token" \
  -d "{
    \"event\": \"user.upgraded\",
    \"data\": { \"user_id\": \"$user_id\" }
  }" -i
echo

# 5) Login après upgrade
echo "==> Login (after upgrade)"
resp2="$(curl -sS -X POST "$BASE_URL/api/login" \
  -H 'Content-Type: application/json' \
  -d "{\"email\":\"$email\",\"password\":\"$password\"}")"

echo "$resp2" | jq .
echo "   is_chirpy_red (after): $(echo "$resp2" | jq -r '.is_chirpy_red')"
echo

# 6) Vérifier en DB (optionnel)
echo "==> Vérification DB"
docker exec -it chirp_pg_database psql -U postgres -d chirpy \
  -c "SELECT id, email, is_chirpy_red FROM users WHERE id = '$user_id';"
