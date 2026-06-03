#!/usr/bin/env bash
set -euo pipefail

: "${IIKO_API_LOGIN:?Set IIKO_API_LOGIN to the raw iikoCloud apiLogin}"

BASE_URL="${IIKO_BASE_URL:-https://api-ru.iiko.services}"
ORG_ID="${IIKO_ORG_ID:-22fc8cb3-0e70-4c9a-b195-8d2301ee0c43}"
EXTERNAL_MENU_ID="${IIKO_EXTERNAL_MENU_ID:-82279}"

api_post() {
  local path="$1"
  local body="$2"
  curl -sS -X POST "${BASE_URL}${path}" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer ${TOKEN}" \
    --data-binary "${body}"
}

command_status() {
  local correlation_id="$1"
  api_post "/api/1/commands/status" \
    "$(jq -n --arg organizationId "$ORG_ID" --arg correlationId "$correlation_id" \
      '{organizationId:$organizationId, correlationId:$correlationId}')"
}

wait_command() {
  local correlation_id="$1"
  local state=""
  for _ in $(seq 1 30); do
    state="$(command_status "$correlation_id" | jq -r '.state // "Unknown"')"
    case "$state" in
      Success) return 0 ;;
      Error)
        command_status "$correlation_id" | jq .
        return 1
        ;;
    esac
    sleep 1
  done
  echo "Command ${correlation_id} timed out in state ${state}" >&2
  return 1
}

create_delivery() {
  local suffix="$1"
  local phone="$2"
  local customer_json="$3"
  local order_id
  order_id="$(uuidgen | tr '[:upper:]' '[:lower:]')"

  local body
  body="$(jq -n \
    --arg organizationId "$ORG_ID" \
    --arg terminalGroupId "$TERMINAL_GROUP_ID" \
    --arg orderId "$order_id" \
    --arg externalNumber "revisitr-demo-${suffix}" \
    --arg phone "$phone" \
    --arg productId "$PRODUCT_ID" \
    --arg paymentTypeId "$PAYMENT_TYPE_ID" \
    --argjson price "$PRODUCT_PRICE" \
    --argjson customer "$customer_json" \
    '{
      organizationId: $organizationId,
      terminalGroupId: $terminalGroupId,
      createOrderSettings: {transportToFrontTimeout: 30},
      order: {
        id: $orderId,
        externalNumber: $externalNumber,
        phone: $phone,
        orderServiceType: "DeliveryByClient",
        comment: "Revisitr demo sync fixture",
        customer: $customer,
        items: [{
          type: "Product",
          productId: $productId,
          amount: 1,
          price: $price
        }],
        sourceKey: "revisitr-demo"
      }
    }
    | if $paymentTypeId != "" and $paymentTypeId != "null" then
        .order.payments = [{
          paymentTypeKind: "Cash",
          sum: $price,
          paymentTypeId: $paymentTypeId
        }]
      else . end')"

  local create_response
  create_response="$(api_post "/api/1/deliveries/create" "$body")"
  echo "$create_response" | jq .

  local creation_status
  creation_status="$(echo "$create_response" | jq -r '.orderInfo.creationStatus // empty')"
  if [[ "$creation_status" == "Error" ]]; then
    echo "Delivery creation failed" >&2
    return 1
  fi

  local correlation_id
  correlation_id="$(echo "$create_response" | jq -r '.correlationId // empty')"
  if [[ -n "$correlation_id" ]]; then
    wait_command "$correlation_id"
  fi

  api_post "/api/1/deliveries/confirm" \
    "$(jq -n --arg organizationId "$ORG_ID" --arg orderId "$order_id" \
      '{organizationId:$organizationId, orderId:$orderId}')" | jq .

  api_post "/api/1/deliveries/close" \
    "$(jq -n --arg organizationId "$ORG_ID" --arg orderId "$order_id" \
      '{organizationId:$organizationId, orderId:$orderId}')" | jq .

  echo "$order_id"
}

TOKEN="$(curl -sS -X POST "${BASE_URL}/api/1/access_token" \
  -H "Content-Type: application/json" \
  --data-binary "$(jq -n --arg apiLogin "$IIKO_API_LOGIN" '{apiLogin:$apiLogin}')" |
  jq -r '.token')"

if [[ -z "$TOKEN" || "$TOKEN" == "null" ]]; then
  echo "Failed to obtain iikoCloud token" >&2
  exit 1
fi

TERMINAL_GROUP_ID="$(api_post "/api/1/terminal_groups" \
  "$(jq -n --arg organizationId "$ORG_ID" '{organizationIds:[$organizationId], includeDisabled:true}')" |
  jq -r '.terminalGroups[0].items[0].id')"

PAYMENT_TYPE_ID="$(api_post "/api/1/payment_types" \
  "$(jq -n --arg organizationId "$ORG_ID" '{organizationIds:[$organizationId]}')" |
  jq -r '[.paymentTypes[] | select(.paymentTypeKind=="Cash")][0].id')"

PRODUCT_JSON="$(api_post "/api/2/menu/by_id" \
  "$(jq -n --arg externalMenuId "$EXTERNAL_MENU_ID" --arg organizationId "$ORG_ID" \
    '{externalMenuId:$externalMenuId, organizationIds:[$organizationId]}')" |
  jq -c '[.itemCategories[].items[] | select((.iikoItemId // .itemId // .id) != null)][0]')"

PRODUCT_ID="$(echo "$PRODUCT_JSON" | jq -r '.iikoItemId // .itemId // .id')"
PRODUCT_PRICE="$(echo "$PRODUCT_JSON" | jq -r '(.itemSizes[0].price // .itemSizes[0].prices[0].price // 250)')"

if [[ -z "$TERMINAL_GROUP_ID" || "$TERMINAL_GROUP_ID" == "null" ]]; then
  echo "No terminal group found" >&2
  exit 1
fi
if [[ -z "$PRODUCT_ID" || "$PRODUCT_ID" == "null" ]]; then
  echo "No product found in external menu ${EXTERNAL_MENU_ID}" >&2
  exit 1
fi

echo "Using terminalGroupId=${TERMINAL_GROUP_ID}"
if [[ -n "$PAYMENT_TYPE_ID" && "$PAYMENT_TYPE_ID" != "null" ]]; then
  echo "Using paymentTypeId=${PAYMENT_TYPE_ID}"
else
  echo "No Cash payment type found; creating unpaid delivery orders"
fi
echo "Using productId=${PRODUCT_ID} price=${PRODUCT_PRICE}"

create_delivery "with-customer" "+79990000101" \
  '{"type":"regular","name":"Revisitr","surname":"Customer","shouldReceiveOrderStatusNotifications":false}'

create_delivery "no-customer" "+79990000102" "null"

echo "Created demo deliveries. Verify with /api/1/deliveries/by_delivery_date_and_status."
