#!/usr/bin/env bash
# auto-test.sh — Flood the fraud detection API with random transactions.
#
# Usage:
#   ./auto-test.sh [options]
#
# Options:
#   --duration=<seconds>         How long to run (default: 60)
#   --rate=<requests_per_second> Target RPS (default: 2)
#   --anomaly-chance=<0-100>     % chance each burst is anomalous (default: 20)
#   --users=<count>              Number of synthetic users (default: 10)
#   --api=<url>                  API base URL (default: http://localhost:8080/api/v1/transactions)
#
# Anomaly types generated:
#   velocity   — same user sends 8 rapid-fire transactions (triggers >5/min rule)
#   amount     — same user sends a normal baseline then a 10× spike
#   travel     — same user sends from Istanbul, then immediately from New York

set -euo pipefail

# ── Defaults ──────────────────────────────────────────────────────────────

DURATION=60
RATE=2
ANOMALY_CHANCE=20
USER_COUNT=10
API_URL="${API_URL:-http://localhost:8080/api/v1/transactions}"

# ── Parse args ────────────────────────────────────────────────────────────

for arg in "$@"; do
  case "$arg" in
    --duration=*)    DURATION="${arg#*=}" ;;
    --rate=*)        RATE="${arg#*=}" ;;
    --anomaly-chance=*) ANOMALY_CHANCE="${arg#*=}" ;;
    --users=*)       USER_COUNT="${arg#*=}" ;;
    --api=*)         API_URL="${arg#*=}" ;;
    -h|--help)
      sed -n '2,25p' "$0" | sed 's/^# \{0,1\}//'
      exit 0
      ;;
    *)
      echo "Unknown option: $arg" >&2
      exit 1
      ;;
  esac
done

# ── Locations ─────────────────────────────────────────────────────────────

LATS=( 41.0082  40.7128  35.6762 -33.8688  25.2048  34.0522  1.3521  51.5074 -23.5505  48.8566 )
LONS=( 28.9784 -74.0060 139.6503 151.2093  55.2708 -118.2437 103.8198 -0.1278 -46.6333   2.3522 )
LOC_NAMES=( istanbul newyork tokyo sydney dubai losangeles singapore london saopaulo paris )

# ── Counters ──────────────────────────────────────────────────────────────

SENT=0
FAILED=0
NORMAL_COUNT=0
ANOMALY_COUNT=0

# ── Helpers ───────────────────────────────────────────────────────────────

rand_int() {
  # random int in [0, $1)
  echo $(( RANDOM % $1 ))
}

rand_amount() {
  # $10 – $400, two decimal places
  local cents=$(( RANDOM % 39001 + 1000 ))
  printf "%.2f" "$(echo "$cents / 100" | bc -l)"
}

post() {
  local payload="$1"
  local code
  code=$(curl -s -o /dev/null -w "%{http_code}" \
    -X POST "$API_URL" \
    -H "Content-Type: application/json" \
    -d "$payload")
  echo "$code"
}

send_tx() {
  local user="$1" amount="$2" lat="$3" lon="$4"
  local payload
  payload=$(printf '{"user_id":"%s","amount":%s,"lat":%s,"lon":%s}' \
    "$user" "$amount" "$lat" "$lon")
  local code
  code=$(post "$payload")
  if [[ "$code" == "201" ]]; then
    (( SENT++ )) || true
    return 0
  else
    (( FAILED++ )) || true
    return 1
  fi
}

# ── Anomaly generators ────────────────────────────────────────────────────

anomaly_velocity() {
  local user="$1"
  local idx=$(( RANDOM % ${#LATS[@]} ))
  local lat="${LATS[$idx]}" lon="${LONS[$idx]}"
  echo "    [anomaly:velocity] $user — 6 rapid txs"
  for _ in 1 2 3 4 5 6; do
    local amt
    amt=$(rand_amount)
    send_tx "$user" "$amt" "$lat" "$lon" || true
    sleep 0.05
  done
  (( ANOMALY_COUNT++ )) || true
}

anomaly_amount() {
  local user="$1"
  local idx=$(( RANDOM % ${#LATS[@]} ))
  local lat="${LATS[$idx]}" lon="${LONS[$idx]}"
  # baseline: 3 small txs to set average
  echo "    [anomaly:amount] $user — baseline then 10× spike"
  for _ in 1 2 3; do
    local base_amt
    base_amt=$(rand_amount)
    send_tx "$user" "$base_amt" "$lat" "$lon" || true
    sleep 0.1
  done
  # spike: ~10× a typical $200 baseline
  local spike
  spike=$(printf "%.2f" "$(echo "$(( RANDOM % 1500 + 2000 )) / 1" | bc -l)")
  send_tx "$user" "$spike" "$lat" "$lon" || true
  (( ANOMALY_COUNT++ )) || true
}

anomaly_travel() {
  local user="$1"
  # Istanbul → New York, 5 minutes apart in real time is impossible
  echo "    [anomaly:travel] $user — Istanbul then New York"
  send_tx "$user" "$(rand_amount)" "41.0082" "28.9784" || true
  sleep 1  # give worker time to set location cache
  send_tx "$user" "$(rand_amount)" "40.7128" "-74.0060" || true
  (( ANOMALY_COUNT++ )) || true
}

trigger_anomaly() {
  local user="$1"
  local roll=$(( RANDOM % 3 ))
  case "$roll" in
    0) anomaly_velocity "$user" ;;
    1) anomaly_amount "$user" ;;
    2) anomaly_travel "$user" ;;
  esac
}

# ── Main loop ─────────────────────────────────────────────────────────────

SLEEP_MS=$(echo "scale=4; 1 / $RATE" | bc -l)

START=$(date +%s)
END=$(( START + DURATION ))

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  auto-test.sh"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  duration      : ${DURATION}s"
echo "  rate          : ${RATE} req/s"
echo "  anomaly chance: ${ANOMALY_CHANCE}%"
echo "  users         : ${USER_COUNT}"
echo "  endpoint      : $API_URL"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

TICK=0
while [[ $(date +%s) -lt $END ]]; do
  (( TICK++ )) || true

  # Pick a random synthetic user
  USER_IDX=$(rand_int "$USER_COUNT")
  USER=$(printf "autotest-%03d" "$USER_IDX")

  # Roll for anomaly
  ROLL=$(rand_int 100)
  if [[ $ROLL -lt $ANOMALY_CHANCE ]]; then
    trigger_anomaly "$USER"
  else
    # Normal transaction
    LOC_IDX=$(rand_int "${#LATS[@]}")
    AMT=$(rand_amount)
    LAT="${LATS[$LOC_IDX]}"
    LON="${LONS[$LOC_IDX]}"
    LOC_NAME="${LOC_NAMES[$LOC_IDX]}"
    if send_tx "$USER" "$AMT" "$LAT" "$LON"; then
      (( NORMAL_COUNT++ )) || true
      printf "  [%4d] OK   %-16s \$%7s  %s\n" "$TICK" "$USER" "$AMT" "$LOC_NAME"
    else
      printf "  [%4d] FAIL %-16s \$%7s  %s\n" "$TICK" "$USER" "$AMT" "$LOC_NAME" >&2
    fi
  fi

  sleep "$SLEEP_MS"
done

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  Done."
echo "  sent    : $SENT"
echo "  failed  : $FAILED"
echo "  normal  : $NORMAL_COUNT"
echo "  anomaly : $ANOMALY_COUNT bursts"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
