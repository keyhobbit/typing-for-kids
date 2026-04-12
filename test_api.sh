#!/usr/bin/env bash
# test_api.sh – KidTyping VN API test suite
# Usage: ./test_api.sh [base_url]
#   e.g. ./test_api.sh http://localhost:11100

set -euo pipefail

BASE="${1:-http://localhost:11100}"
DEVICE="test-device-$(date +%s)-$$"
TOKEN=""
USER_ID=""
REGISTER_NAME="testuser_$$"
FAILURES=0

# ── Colors ────────────────────────────────────────────────────────────────
GREEN='\033[0;32m'; RED='\033[0;31m'; CYAN='\033[0;36m'
BOLD='\033[1m'; RESET='\033[0m'

pass()    { echo -e "  ${GREEN}✓ PASS${RESET}  $1"; }
fail()    { echo -e "  ${RED}✗ FAIL${RESET}  $1"; FAILURES=$((FAILURES+1)); }
section() { echo -e "\n${BOLD}${CYAN}── $1 ──${RESET}"; }

# ── Helpers ───────────────────────────────────────────────────────────────
# call METHOD PATH [BODY] → prints <body>\n<status_code>
call_api() {
  local method=$1 path=$2 data=${3:-}
  local args=(-s -w '\n%{http_code}' -X "$method" "$BASE$path"
              -H "Content-Type: application/json")
  [[ -n "${TOKEN:-}" ]] && args+=(-H "Authorization: Bearer $TOKEN")
  [[ -n "$data" ]]      && args+=(-d "$data")
  curl "${args[@]}"
}

body()   { echo "$1" | head -n1; }
status() { echo "$1" | tail -n1; }

# Extract a key from a JSON body using python3
jget() { echo "$1" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d$2)" 2>/dev/null || echo ""; }

check() {
  local actual=$1 expected=$2 name=$3
  if [[ "$actual" == "$expected" ]]; then
    pass "$name (HTTP $actual)"
  else
    fail "$name – expected HTTP $expected, got HTTP $actual"
  fi
}

# ── Preflight ─────────────────────────────────────────────────────────────
echo -e "${BOLD}KidTyping VN – API Test Suite${RESET}"
echo -e "Base URL : $BASE"
echo -e "Device ID: ${DEVICE:0:32}…"

if ! curl -s --max-time 3 "$BASE/" > /dev/null 2>&1; then
  echo -e "${RED}Server not reachable at $BASE – is it running?${RESET}"
  exit 1
fi
echo -e "${GREEN}Server is up ✓${RESET}"

# ─────────────────────────────────────────────────────────────────────────
section "1. Lesson API"

RESP=$(call_api GET "/api/next?level=1")
check "$(status "$RESP")" "200" "GET /api/next?level=1"
CONTENT=$(jget "$(body "$RESP")" "['content']")
[[ -n "$CONTENT" ]] && pass "Response has 'content' field: '$CONTENT'" \
                    || fail "Response missing 'content' field"

for lvl in 2 3; do
  RESP=$(call_api GET "/api/next?level=$lvl")
  check "$(status "$RESP")" "200" "GET /api/next?level=$lvl"
done

# ─────────────────────────────────────────────────────────────────────────
section "2. Guest Auth"

RESP=$(call_api POST "/api/auth/guest" "{\"device_id\":\"$DEVICE\"}")
check "$(status "$RESP")" "200" "POST /api/auth/guest"

TOKEN=$(jget "$(body "$RESP")" "['token']")
USER_ID=$(jget "$(body "$RESP")" "['user']['id']")
IS_GUEST=$(jget "$(body "$RESP")" "['user']['is_guest']")

[[ -n "$TOKEN" ]]                   && pass "Token received (${TOKEN:0:12}…)" \
                                    || fail "No token in response"
[[ "$IS_GUEST" == "True" ]]         && pass "User is guest"  \
                                    || fail "Expected is_guest=true"

# Idempotency: same device_id must return the same user
RESP2=$(call_api POST "/api/auth/guest" "{\"device_id\":\"$DEVICE\"}")
USER_ID2=$(jget "$(body "$RESP2")" "['user']['id']")
[[ "$USER_ID" == "$USER_ID2" ]]     && pass "Same device_id → same user (idempotent)" \
                                    || fail "Same device_id returned different user ($USER_ID vs $USER_ID2)"

# Bad request (no device_id)
RESP=$(call_api POST "/api/auth/guest" '{}')
check "$(status "$RESP")" "400" "POST /api/auth/guest with empty device_id → 400"

# ─────────────────────────────────────────────────────────────────────────
section "3. Rename"

RESP=$(call_api PUT "/api/user/name" '{"name":"TestPlayer"}')
check "$(status "$RESP")" "200" "PUT /api/user/name"
NEW_NAME=$(jget "$(body "$RESP")" "['user']['name']")
[[ "$NEW_NAME" == "TestPlayer" ]]   && pass "Name updated to 'TestPlayer'" \
                                    || fail "Expected name 'TestPlayer', got '$NEW_NAME'"

# Empty name → 400
RESP=$(call_api PUT "/api/user/name" '{"name":""}')
check "$(status "$RESP")" "400" "PUT /api/user/name with empty name → 400"

# No auth → 401
OLD_TOKEN=$TOKEN; TOKEN=""
RESP=$(call_api PUT "/api/user/name" '{"name":"Hacker"}')
check "$(status "$RESP")" "401" "PUT /api/user/name without token → 401"
TOKEN=$OLD_TOKEN

# ─────────────────────────────────────────────────────────────────────────
section "4. Score Submission"

RESP=$(call_api POST "/api/score" '{"correct":5,"total":5,"level":1}')
check "$(status "$RESP")" "200" "POST /api/score (5 correct, level 1)"

RESP=$(call_api POST "/api/score" '{"correct":3,"total":5,"level":2}')
check "$(status "$RESP")" "200" "POST /api/score (3 correct, level 2)"

RESP=$(call_api POST "/api/score" '{"correct":7,"total":10,"level":3}')
check "$(status "$RESP")" "200" "POST /api/score (7 correct, level 3)"

# Negative correct → 400
RESP=$(call_api POST "/api/score" '{"correct":-1,"total":5,"level":1}')
check "$(status "$RESP")" "400" "POST /api/score with negative correct → 400"

# No auth → 401
OLD_TOKEN=$TOKEN; TOKEN=""
RESP=$(call_api POST "/api/score" '{"correct":1,"total":1,"level":1}')
check "$(status "$RESP")" "401" "POST /api/score without token → 401"
TOKEN=$OLD_TOKEN

# ─────────────────────────────────────────────────────────────────────────
section "5. Ranking"

for period in day week month year; do
  RESP=$(call_api GET "/api/ranking?period=$period")
  check "$(status "$RESP")" "200" "GET /api/ranking?period=$period"
done

# Default period (no param)
RESP=$(call_api GET "/api/ranking")
check "$(status "$RESP")" "200" "GET /api/ranking (no period param)"

# Current user should appear in day ranking
RESP=$(call_api GET "/api/ranking?period=day")
APPEARS=$(echo "$(body "$RESP")" | python3 -c \
  "import sys,json; entries=json.load(sys.stdin); print(any(e['user_id']=='$USER_ID' for e in entries))" 2>/dev/null)
[[ "$APPEARS" == "True" ]]          && pass "Current user appears in day ranking" \
                                    || fail "Current user NOT found in day ranking"

# ─────────────────────────────────────────────────────────────────────────
section "6. Registration"

RESP=$(call_api POST "/api/auth/register" \
  "{\"device_id\":\"$DEVICE\",\"username\":\"$REGISTER_NAME\",\"password\":\"pass1234\"}")
check "$(status "$RESP")" "200" "POST /api/auth/register"
if [[ "$(status "$RESP")" == "200" ]]; then
  TOKEN=$(jget "$(body "$RESP")" "['token']")
  IS_GUEST2=$(jget "$(body "$RESP")" "['user']['is_guest']")
  [[ "$IS_GUEST2" == "False" ]] && pass "Registered user is_guest=False" \
                                 || fail "Registered user should not be guest"
fi

# Duplicate username → 400
RESP=$(call_api POST "/api/auth/register" \
  "{\"device_id\":\"other-dev-1\",\"username\":\"$REGISTER_NAME\",\"password\":\"pass5678\"}")
check "$(status "$RESP")" "400" "Duplicate username → 400"

# Short password → 400
RESP=$(call_api POST "/api/auth/register" \
  "{\"device_id\":\"other-dev-2\",\"username\":\"${REGISTER_NAME}b\",\"password\":\"123\"}")
check "$(status "$RESP")" "400" "Short password (3 chars) → 400"

# Username with spaces → 400
RESP=$(call_api POST "/api/auth/register" \
  "{\"device_id\":\"other-dev-3\",\"username\":\"bad user\",\"password\":\"pass1234\"}")
check "$(status "$RESP")" "400" "Username with spaces → 400"

# ─────────────────────────────────────────────────────────────────────────
section "7. Login"

RESP=$(call_api POST "/api/auth/login" \
  "{\"username\":\"$REGISTER_NAME\",\"password\":\"pass1234\"}")
check "$(status "$RESP")" "200" "POST /api/auth/login (correct credentials)"
if [[ "$(status "$RESP")" == "200" ]]; then
  TOKEN=$(jget "$(body "$RESP")" "['token']")
  pass "New token received after login (${TOKEN:0:12}…)"
fi

# Wrong password → 401
RESP=$(call_api POST "/api/auth/login" \
  "{\"username\":\"$REGISTER_NAME\",\"password\":\"wrongpass\"}")
check "$(status "$RESP")" "401" "POST /api/auth/login (wrong password) → 401"

# Non-existent user → 401
RESP=$(call_api POST "/api/auth/login" \
  '{"username":"nobody_xyz","password":"pass1234"}')
check "$(status "$RESP")" "401" "POST /api/auth/login (non-existent user) → 401"

# ─────────────────────────────────────────────────────────────────────────
section "8. Logout"

RESP=$(call_api POST "/api/auth/logout")
check "$(status "$RESP")" "200" "POST /api/auth/logout"

# Revoked token must be rejected for protected endpoints
RESP=$(call_api POST "/api/score" '{"correct":1,"total":1,"level":1}')
check "$(status "$RESP")" "401" "Revoked token rejected by POST /api/score → 401"

RESP=$(call_api PUT "/api/user/name" '{"name":"PostLogout"}')
check "$(status "$RESP")" "401" "Revoked token rejected by PUT /api/user/name → 401"

# ─────────────────────────────────────────────────────────────────────────
echo ""
if [[ $FAILURES -eq 0 ]]; then
  echo -e "${BOLD}${GREEN}All tests passed! 🎉${RESET}"
else
  echo -e "${BOLD}${RED}$FAILURES test(s) failed.${RESET}"
  exit 1
fi
