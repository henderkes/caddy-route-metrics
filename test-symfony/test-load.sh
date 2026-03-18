#!/usr/bin/env bash

HOST="${HOST:-localhost:8080}"
TOTAL="${1:-500}"

uuid() {
    printf '%04x%04x-%04x-4%03x-%04x-%04x%04x%04x' \
        $RANDOM $RANDOM $RANDOM $((RANDOM % 4096)) \
        $(( (RANDOM % 16384) + 32768 )) $RANDOM $RANDOM $RANDOM
}

ROUTES=(
    "GET  /                                 20"
    "GET  /about                            10"
    "GET  /contact                           8"
    "POST /contact                           5"
    "GET  /login                             8"
    "POST /login                             5"
    "GET  /users                            15"
    "POST /users                             8"
    "GET  /users/:id                        20"
    "PUT  /users/:id                        10"
    "DELETE /users/:id                       5"
    "GET  /users/:id/posts                  12"
    "GET  /users/:id/posts/:id               8"
    "GET  /orders/:uuid                     10"
    "PUT  /orders/:uuid/cancel               4"
    "GET  /api/health                       15"
    "GET  /api/search                       10"
    "POST /api/upload                        5"
    "GET  /api/reports/:id                   6"
    "GET  /error/bad-request                 3"
    "GET  /error/forbidden                   2"
    "GET  /error/not-found                   3"
    "POST /error/validation                  3"
    "GET  /error/server                      2"
    "GET  /error/unavailable                 1"
)

POOL=()
for entry in "${ROUTES[@]}"; do
    weight="${entry##* }"
    weight=$(echo "$weight" | tr -d ' ')
    for ((w=0; w<weight; w++)); do
        POOL+=("$entry")
    done
done
POOL_SIZE=${#POOL[@]}

declare -A HITS
sent=0
errors=0
start=$(date +%s)

echo "Sending $TOTAL requests to $HOST..."
echo ""

while [ "$sent" -lt "$TOTAL" ]; do
    entry="${POOL[$((RANDOM % POOL_SIZE))]}"
    method=$(echo "$entry" | awk '{print $1}')
    path=$(echo "$entry" | awk '{print $2}')

    path="${path//:id/$((RANDOM % 500 + 1))}"
    path="${path//:id/$((RANDOM % 200 + 1))}"
    path="${path//:uuid/$(uuid)}"

    status=$(curl -s -o /dev/null -w "%{http_code}" -X "$method" "http://$HOST$path" 2>/dev/null || echo "000")

    norm=$(echo "$path" | sed -E 's|/[0-9a-f]+-[0-9a-f]+-[0-9a-f]+-[0-9a-f]+-[0-9a-f]+(/\|$)|/:uuid\1|g; s|/[0-9]+|/:id|g')
    counter_key="$method $norm"
    HITS["$counter_key"]=$(( ${HITS["$counter_key"]:-0} + 1 ))

    if [ "$status" = "000" ]; then
        errors=$((errors + 1))
    fi

    sent=$((sent + 1))

    if [ $((sent % 50)) -eq 0 ]; then
        elapsed=$(( $(date +%s) - start ))
        echo "  $sent / $TOTAL (${elapsed}s)"
    fi
done

elapsed=$(( $(date +%s) - start ))

echo ""
echo "Done: $sent requests in ${elapsed}s ($errors connection errors)"
echo ""
printf "%-8s %-40s %s\n" "METHOD" "PATH" "COUNT"
printf "%-8s %-40s %s\n" "------" "----" "-----"
(
    for key in "${!HITS[@]}"; do
        echo "${HITS[$key]} $key"
    done
) | sort -rn | while IFS= read -r line; do
    count="${line%% *}"
    rest="${line#* }"
    method="${rest%% *}"
    path="${rest#* }"
    printf "%-8s %-40s %s\n" "$method" "$path" "$count"
done
