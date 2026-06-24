#!/usr/bin/env bash
# Seed realistic demo delivery orders into the iiko test stand for Revisitr sync.
# Pulls real product ids + prices from Resto, creates closed delivery orders via
# iikoCloud for a small set of repeat customers (matched by phone in Revisitr).
#
# Requires: bash, curl, jq, uuidgen, openssl.
# Env: IIKO_API_LOGIN (required). Optional overrides below.
set -euo pipefail

: "${IIKO_API_LOGIN:?Set IIKO_API_LOGIN to the raw iikoCloud apiLogin}"
BASE="${IIKO_BASE_URL:-https://api-ru.iiko.services}"
ORG="${IIKO_ORG_ID:-22fc8cb3-0e70-4c9a-b195-8d2301ee0c43}"
TG="${IIKO_TERMINAL_GROUP_ID:-7a35c826-7314-9d1c-019e-5eb21b0a0066}"
RESTO="${IIKO_RESTO_URL:-https://260-347-461.iiko.it/resto}"
RESTO_LOGIN="${IIKO_RESTO_LOGIN:-user}"
RESTO_PASS="${IIKO_RESTO_PASS:-user#test}"
# Cash payment type (from Resto; Cloud /payment_types is empty on this stand).
# Required so /deliveries/close succeeds (else PaymentSumNotEnough).
# NOTE: close also needs an OPEN cash shift in iikoFront, else CafeSessionNotFound.
CASH_PAYMENT_ID="${IIKO_CASH_PAYMENT_ID:-09322f46-578a-d210-add7-eec222a08871}"

api_post() {
  curl -sS -X POST "${BASE}$1" -H "Content-Type: application/json" \
    -H "Authorization: Bearer ${TOKEN}" --data-binary "$2"
}

wait_command() {
  local cid="$1" state=""
  for _ in $(seq 1 30); do
    state="$(api_post "/api/1/commands/status" \
      "$(jq -n --arg o "$ORG" --arg c "$cid" '{organizationId:$o, correlationId:$c}')" \
      | jq -r '.state // "Unknown"')"
    case "$state" in
      Success) return 0 ;;
      Error)
        api_post "/api/1/commands/status" \
          "$(jq -n --arg o "$ORG" --arg c "$cid" '{organizationId:$o, correlationId:$c}')" \
          | jq -c '.exception' >&2
        return 1 ;;
    esac
    sleep 1
  done
  echo "command ${cid} timed out (${state})" >&2; return 1
}

# --- Cloud token ---
TOKEN="$(curl -sS -X POST "${BASE}/api/1/access_token" -H "Content-Type: application/json" \
  --data-binary "$(jq -n --arg l "$IIKO_API_LOGIN" '{apiLogin:$l}')" | jq -r '.token')"
[ -n "$TOKEN" ] && [ "$TOKEN" != "null" ] || { echo "no cloud token" >&2; exit 1; }

# --- Resto products (id + real price) ---
RESTO_SHA1="$(printf '%s' "$RESTO_PASS" | openssl sha1 | awk '{print $NF}')"
RESTO_KEY="$(curl -sS "${RESTO}/api/auth?login=${RESTO_LOGIN}&pass=${RESTO_SHA1}")"
PRODUCTS="$(curl -sS "${RESTO}/api/v2/entities/products/list" \
  --data-urlencode "key=${RESTO_KEY}" --data-urlencode "includeDeleted=false" -G \
  | jq -c '[.[] | select(.name|test("Revisitr";"i"))
            | select(.defaultIncludedInMenu==true and .placeType!=null and .defaultSalePrice>0)
            | {id, name, price:.defaultSalePrice}]')"
PCOUNT="$(echo "$PRODUCTS" | jq 'length')"
[ "$PCOUNT" -gt 0 ] || { echo "no active priced products in Resto" >&2; exit 1; }
echo "Using terminalGroup=${TG}, ${PCOUNT} products"

# --- Customers (repeat visitors; Revisitr matches by phone) ---
NAMES=(Анна Дмитрий Мария Сергей Елена Алексей)
SURNAMES=(Смирнова Иванов Кузнецова Попов Соколова Лебедев)
PHONES=(+79161234501 +79161234502 +79161234503 +79161234504 +79161234505 +79161234506)

# Order plan: "<customerIdx>:<itemCount>" — uneven distribution = repeat visits
ORDERS=(0:2 0:1 0:3 0:1 1:2 1:1 1:2 2:1 2:3 2:2 3:2 3:1 4:3 4:1 5:2 5:2)

cursor=0
n=0
for spec in "${ORDERS[@]}"; do
  ci="${spec%%:*}"; cnt="${spec##*:}"
  n=$((n+1))

  # build item list by walking the product cursor
  items="[]"; total=0
  for _ in $(seq 1 "$cnt"); do
    p="$(echo "$PRODUCTS" | jq -c ".[$(( cursor % PCOUNT ))]")"
    cursor=$((cursor+1))
    pid="$(echo "$p" | jq -r '.id')"; price="$(echo "$p" | jq -r '.price')"
    items="$(echo "$items" | jq -c --arg id "$pid" --argjson pr "$price" \
      '. + [{type:"Product", productId:$id, amount:1, price:$pr}]')"
    total=$((total + price))
  done

  oid="$(uuidgen | tr '[:upper:]' '[:lower:]')"
  body="$(jq -n \
    --arg org "$ORG" --arg tg "$TG" --arg oid "$oid" \
    --arg ext "revisitr-demo-$(printf '%02d' "$n")" \
    --arg phone "${PHONES[$ci]}" --arg name "${NAMES[$ci]}" --arg surname "${SURNAMES[$ci]}" \
    --argjson items "$items" \
    --arg cash "$CASH_PAYMENT_ID" --argjson total "$total" \
    '{organizationId:$org, terminalGroupId:$tg,
      createOrderSettings:{transportToFrontTimeout:30},
      order:{id:$oid, externalNumber:$ext, phone:$phone,
        orderServiceType:"DeliveryByClient",
        comment:"Revisitr demo seed",
        customer:{name:$name, surname:$surname, type:"regular",
                  shouldReceiveOrderStatusNotifications:false},
        items:$items,
        payments:[{paymentTypeKind:"Cash", sum:$total, paymentTypeId:$cash,
                   isProcessedExternally:false}],
        sourceKey:"revisitr-demo"}}')"

  resp="$(api_post "/api/1/deliveries/create" "$body")"
  st="$(echo "$resp" | jq -r '.orderInfo.creationStatus // empty')"
  if [ "$st" = "Error" ]; then
    echo "[$n] create error: $(echo "$resp" | jq -c '.orderInfo.errorInfo.code')" >&2
    continue
  fi
  cid="$(echo "$resp" | jq -r '.correlationId // empty')"
  [ -n "$cid" ] && { wait_command "$cid" || { echo "[$n] create failed" >&2; continue; }; }

  api_post "/api/1/deliveries/confirm" \
    "$(jq -n --arg o "$ORG" --arg id "$oid" '{organizationId:$o, orderId:$id}')" >/dev/null

  close_resp="$(api_post "/api/1/deliveries/close" \
    "$(jq -n --arg o "$ORG" --arg id "$oid" '{organizationId:$o, orderId:$id}')")"
  close_cid="$(echo "$close_resp" | jq -r '.correlationId // empty')"
  close_state="closed?"
  if [ -n "$close_cid" ]; then
    if wait_command "$close_cid"; then close_state="CLOSED"; else close_state="CLOSE-FAILED"; fi
  fi

  echo "[$n] ${NAMES[$ci]} ${PHONES[$ci]} — ${cnt} item(s), ${total}₽ — ${close_state} — ${oid}"
done

echo "Done. Verify with /api/1/deliveries/by_delivery_date_and_status (today)."
