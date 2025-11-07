#!/bin/bash

# Simple Website Pressure Test
# Continuously curls websites to generate network traffic

CONCURRENT=${1:-20}           # Concurrent requests
STATS_INTERVAL=${2:-5}        # Print stats every N seconds

WEBSITES=(
    "https://www.google.com"
    "https://www.github.com"
    "https://www.stackoverflow.com"
    "https://www.wikipedia.org"
    "https://www.reddit.com"
    "https://www.youtube.com"
    "https://www.twitter.com"
    "https://www.linkedin.com"
    "https://www.amazon.com"
    "https://www.microsoft.com"
    "https://www.apple.com"
    "https://www.netflix.com"
    "https://www.bbc.com"
    "https://www.cnn.com"
    "https://www.nytimes.com"
)

STATS_FILE=$(mktemp)
START_TIME=$(date +%s)

cleanup() {
    echo ""
    echo "Stopping..."
    jobs -p | xargs kill 2>/dev/null
    rm -f "$STATS_FILE"
    exit
}
trap cleanup EXIT INT TERM

# Initialize counters
echo "0|0|0" > "$STATS_FILE"

# Worker function - just curl and update stats
make_request() {
    local url=$1
    local result=$(curl -s -o /dev/null -w "%{http_code}" --max-time 10 --connect-timeout 5 "$url" 2>/dev/null)
    
    # Update stats
    (
        IFS='|' read -r total success fail < "$STATS_FILE"
        total=$((total + 1))
        if [ "$result" -ge 200 ] && [ "$result" -lt 400 ] 2>/dev/null; then
            success=$((success + 1))
        else
            fail=$((fail + 1))
        fi
        echo "$total|$success|$fail" > "$STATS_FILE"
    )
}

# Print stats
print_stats() {
    IFS='|' read -r total success fail < "$STATS_FILE"
    local elapsed=$(($(date +%s) - START_TIME))
    local rps=0
    if [ $elapsed -gt 0 ]; then
        rps=$((total / elapsed))
    fi
    clear
    echo "=== Pressure Test Running ==="
    echo "Time: ${elapsed}s"
    echo "Total Requests: $total"
    echo "Success: $success"
    echo "Failed: $fail"
    echo "Requests/sec: $rps"
    echo "=============================="
}

# Stats printer
stats_loop() {
    while true; do
        sleep "$STATS_INTERVAL"
        print_stats
    done
}

# Main loop - just keep curling
main() {
    echo "Starting pressure test..."
    echo "Concurrent: $CONCURRENT"
    echo "Stats every: ${STATS_INTERVAL}s"
    echo "Press Ctrl+C to stop"
    echo ""
    
    stats_loop &
    
    while true; do
        # Launch concurrent requests
        for ((i=0; i<CONCURRENT; i++)); do
            url="${WEBSITES[$((RANDOM % ${#WEBSITES[@]}))]}"
            make_request "$url" &
        done
        
        # Wait a tiny bit before next batch
        sleep 0.1
    done
}

main