#!/usr/bin/env bash
# manual-input.sh — Send a single transaction to the fraud detection API.
#
# Usage:
#   ./manual-input.sh <user_id> <amount> <location>
#
# Location can be:
#   - A city name shorthand: istanbul, newyork, tokyo, sydney, dubai,
#                            losangeles, singapore, london, saopaulo
#   - Explicit coordinates:  "lat,lon"  (e.g. "41.0082,28.9784")
#
# Examples:
#   ./manual-input.sh user-001 250.00 istanbul
#   ./manual-input.sh user-007 9500.00 newyork
#   ./manual-input.sh user-003 120.50 "41.0082,28.9784"

set -euo pipefail

API_URL="${API_URL:-http://localhost:8080/api/v1/transactions}"

# ── Args ──────────────────────────────────────────────────────────────────

if [[ $# -lt 3 ]]; then
  echo "Usage: $0 <user_id> <amount> <location>" >&2
  echo "       location: istanbul | newyork | tokyo | sydney | dubai |" >&2
  echo "                 losangeles | singapore | london | saopaulo | \"lat,lon\"" >&2
  exit 1
fi

USER_ID="$1"
AMOUNT="$2"
LOCATION="$3"

# ── Resolve location ──────────────────────────────────────────────────────

resolve_location() {
  case "$(echo "$1" | tr '[:upper:]' '[:lower:]')" in
    istanbul)    echo "41.0082 28.9784" ;;
    newyork)     echo "40.7128 -74.0060" ;;
    tokyo)       echo "35.6762 139.6503" ;;
    sydney)      echo "-33.8688 151.2093" ;;
    dubai)       echo "25.2048 55.2708" ;;
    losangeles)  echo "34.0522 -118.2437" ;;
    singapore)   echo "1.3521 103.8198" ;;
    london)      echo "51.5074 -0.1278" ;;
    saopaulo)    echo "-23.5505 -46.6333" ;;
    *,*)
      LAT="${1%%,*}"
      LON="${1##*,}"
      echo "$LAT $LON"
      ;;
    *)
      echo "Unknown location: $1" >&2
      echo "Use a city name or 'lat,lon' format." >&2
      exit 1
      ;;
  esac
}

read -r LAT LON <<< "$(resolve_location "$LOCATION")"

# ── Send request ──────────────────────────────────────────────────────────

PAYLOAD=$(printf '{"user_id":"%s","amount":%s,"lat":%s,"lon":%s}' \
  "$USER_ID" "$AMOUNT" "$LAT" "$LON")

echo "Sending transaction..."
echo "  user_id : $USER_ID"
echo "  amount  : \$$AMOUNT"
echo "  location: $LOCATION ($LAT, $LON)"
echo "  endpoint: $API_URL"
echo ""

RESPONSE=$(curl -s -w "\n%{http_code}" \
  -X POST "$API_URL" \
  -H "Content-Type: application/json" \
  -d "$PAYLOAD")

HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | head -n-1)

if [[ "$HTTP_CODE" == "201" ]]; then
  TX_ID=$(echo "$BODY" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
  echo "OK  (HTTP $HTTP_CODE)"
  echo "  transaction id: $TX_ID"
else
  echo "FAIL  (HTTP $HTTP_CODE)" >&2
  echo "$BODY" >&2
  exit 1
fi
