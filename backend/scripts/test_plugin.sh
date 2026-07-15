#!/usr/bin/env bash
#
# test_plugin.sh — live E2E smoke test for the POS loyalty plugin API.
#
# Simulates the till (.dll) with plain HTTP: logs in, mints an API key,
# then drives identify -> redeem -> accrue plus the negative cases, printing
# PASS/FAIL per check.
#
# Requirements: bash, curl, jq. Running backend + Redis + Postgres, migration
# 00046 applied, an active loyalty program, and a guest who took a word-code in
# the Telegram bot (Баланс -> "🎫 Код для кассы").
#
# Usage:
#   EMAIL=... PASSWORD=... INTEGRATION_ID=1 [CODE=трасса] [BASE_URL=...] \
#     ./scripts/test_plugin.sh
#
# If CODE is unset the script prompts for it (the word shown in the bot).

set -uo pipefail

BASE_URL="${BASE_URL:-http://localhost:9721}"
EMAIL="${EMAIL:-}"
PASSWORD="${PASSWORD:-}"
INTEGRATION_ID="${INTEGRATION_ID:-}"
CODE="${CODE:-}"

# --- tiny output helpers ---
green() { printf '\033[32m%s\033[0m\n' "$1"; }
red()   { printf '\033[31m%s\033[0m\n' "$1"; }
dim()   { printf '\033[2m%s\033[0m\n' "$1"; }

PASS=0
FAIL=0

# expect NAME EXPECTED_STATUS ACTUAL_STATUS BODY
expect() {
  local name="$1" want="$2" got="$3" body="$4"
  if [ "$want" = "$got" ]; then
    green "PASS  $name  (HTTP $got)"
    PASS=$((PASS + 1))
  else
    red   "FAIL  $name  (want HTTP $want, got $got)"
    dim   "      body: $body"
    FAIL=$((FAIL + 1))
  fi
}

# http METHOD URL AUTH_HEADER [JSON_BODY]
# echoes "<body>\n<status>"; caller splits the trailing status line.
http() {
  local method="$1" url="$2" auth="$3" data="${4:-}"
  if [ -n "$data" ]; then
    curl -sS -X "$method" "$url" -H "$auth" \
      -H 'Content-Type: application/json' -d "$data" -w $'\n%{http_code}'
  else
    curl -sS -X "$method" "$url" -H "$auth" -w $'\n%{http_code}'
  fi
}

# split last invocation's output into $BODY / $STATUS
run() {
  local out; out="$(http "$@")"
  STATUS="${out##*$'\n'}"
  BODY="${out%$'\n'*}"
}

# --- preflight ---
command -v jq >/dev/null 2>&1 || { red "jq is required"; exit 1; }
[ -n "$EMAIL" ]  || { red "set EMAIL";  exit 1; }
[ -n "$PASSWORD" ] || { red "set PASSWORD"; exit 1; }
[ -n "$INTEGRATION_ID" ] || { red "set INTEGRATION_ID"; exit 1; }

echo "Target: $BASE_URL  integration=$INTEGRATION_ID"
echo

# --- 1. login -> JWT ---
run POST "$BASE_URL/api/v1/auth/login" "Accept: application/json" \
  "{\"email\":\"$EMAIL\",\"password\":\"$PASSWORD\"}"
if [ "$STATUS" != "200" ]; then
  red "login failed (HTTP $STATUS): $BODY"; exit 1
fi
JWT="$(echo "$BODY" | jq -r '.tokens.access_token')"
[ -n "$JWT" ] && [ "$JWT" != "null" ] || { red "no access_token in login response"; exit 1; }
green "login OK"

# --- 2. mint an API key (shown once) ---
run POST "$BASE_URL/api/v1/pos-plugin/admin/keys" "Authorization: Bearer $JWT" \
  "{\"integration_id\":$INTEGRATION_ID,\"label\":\"test_plugin.sh\"}"
if [ "$STATUS" != "201" ]; then
  red "create key failed (HTTP $STATUS): $BODY"; exit 1
fi
APIKEY="$(echo "$BODY" | jq -r '.key')"
KEYID="$(echo "$BODY" | jq -r '.id')"
[ -n "$APIKEY" ] && [ "$APIKEY" != "null" ] || { red "no key in response"; exit 1; }
green "api key minted (id=$KEYID)"
echo

# --- 3. negative cases that need no valid code ---
run GET "$BASE_URL/api/v1/pos-plugin/config" "X-API-Key: rvk_totally_wrong"
expect "bad api key -> 401" 401 "$STATUS" "$BODY"

run GET "$BASE_URL/api/v1/pos-plugin/config" "X-API-Key: $APIKEY"
expect "config with valid key -> 200" 200 "$STATUS" "$BODY"
dim "      config: $BODY"

run POST "$BASE_URL/api/v1/pos-plugin/identify" "X-API-Key: $APIKEY" \
  '{"code":"этогословатнет","order_total":1000}'
expect "unknown code -> 404" 404 "$STATUS" "$BODY"

run POST "$BASE_URL/api/v1/pos-plugin/redeem" "X-API-Key: $APIKEY" \
  '{"session":"bogus-session","order_id":"neg-1","amount":50}'
expect "invalid session -> 401" 401 "$STATUS" "$BODY"
echo

# --- 4. real flow: needs a fresh code from the bot ---
if [ -z "$CODE" ]; then
  echo "Open the client bot -> Баланс -> «🎫 Код для кассы», then paste the word:"
  read -r CODE
fi
[ -n "$CODE" ] || { red "no code provided; skipping the live flow"; exit 1; }

ORDER="e2e-$(date +%s)"   # unique per run so idempotency starts clean

run POST "$BASE_URL/api/v1/pos-plugin/identify" "X-API-Key: $APIKEY" \
  "{\"code\":\"$CODE\",\"order_total\":1000}"
expect "identify with real code -> 200" 200 "$STATUS" "$BODY"
if [ "$STATUS" != "200" ]; then
  red "cannot continue without a session"; exit 1
fi
SESSION="$(echo "$BODY" | jq -r '.session')"
AVAILABLE="$(echo "$BODY" | jq -r '.client.available_to_redeem')"
dim "      client: $(echo "$BODY" | jq -c '.client')"

# redeem a small, safe amount (min(10, available))
REDEEM=10
awk "BEGIN{exit !($AVAILABLE < 10)}" && REDEEM="$AVAILABLE"

run POST "$BASE_URL/api/v1/pos-plugin/redeem" "X-API-Key: $APIKEY" \
  "{\"session\":\"$SESSION\",\"order_id\":\"$ORDER\",\"amount\":$REDEEM}"
expect "redeem $REDEEM -> 200" 200 "$STATUS" "$BODY"
BAL1="$(echo "$BODY" | jq -r '.balance_after')"

# idempotency: same order_id + same op must not double-spend
run POST "$BASE_URL/api/v1/pos-plugin/redeem" "X-API-Key: $APIKEY" \
  "{\"session\":\"$SESSION\",\"order_id\":\"$ORDER\",\"amount\":$REDEEM}"
expect "redeem repeat -> 200" 200 "$STATUS" "$BODY"
BAL2="$(echo "$BODY" | jq -r '.balance_after')"
if [ "$BAL1" = "$BAL2" ]; then
  green "PASS  redeem idempotent (balance stable at $BAL1)"; PASS=$((PASS + 1))
else
  red "FAIL  redeem NOT idempotent ($BAL1 -> $BAL2)"; FAIL=$((FAIL + 1))
fi

# over-cap redeem must be rejected
run POST "$BASE_URL/api/v1/pos-plugin/redeem" "X-API-Key: $APIKEY" \
  "{\"session\":\"$SESSION\",\"order_id\":\"$ORDER-big\",\"amount\":99999999}"
expect "redeem over available -> 400" 400 "$STATUS" "$BODY"

# accrue on the same order
run POST "$BASE_URL/api/v1/pos-plugin/accrue" "X-API-Key: $APIKEY" \
  "{\"session\":\"$SESSION\",\"order_id\":\"$ORDER\",\"amount\":1000}"
expect "accrue 1000 -> 200" 200 "$STATUS" "$BODY"
ACC1="$(echo "$BODY" | jq -r '.accrued')"
dim "      accrued: $ACC1"

run POST "$BASE_URL/api/v1/pos-plugin/accrue" "X-API-Key: $APIKEY" \
  "{\"session\":\"$SESSION\",\"order_id\":\"$ORDER\",\"amount\":1000}"
expect "accrue repeat -> 200" 200 "$STATUS" "$BODY"
ACC2="$(echo "$BODY" | jq -r '.accrued')"
if [ "$ACC1" = "$ACC2" ]; then
  green "PASS  accrue idempotent (accrued stable at $ACC1)"; PASS=$((PASS + 1))
else
  red "FAIL  accrue NOT idempotent ($ACC1 -> $ACC2)"; FAIL=$((FAIL + 1))
fi

# code is one-time: a second identify with the same word must fail
run POST "$BASE_URL/api/v1/pos-plugin/identify" "X-API-Key: $APIKEY" \
  "{\"code\":\"$CODE\",\"order_total\":1000}"
expect "reuse consumed code -> 404" 404 "$STATUS" "$BODY"
echo

# --- 5. cleanup: revoke the test key ---
run DELETE "$BASE_URL/api/v1/pos-plugin/admin/keys/$KEYID" "Authorization: Bearer $JWT"
expect "revoke test key -> 200" 200 "$STATUS" "$BODY"

echo
echo "----------------------------------------"
if [ "$FAIL" -eq 0 ]; then
  green "ALL PASSED ($PASS checks)"
else
  red "$FAIL FAILED, $PASS passed"
  exit 1
fi
