#!/bin/bash
# Test profile commands
set -e

TMPHOME=$(mktemp -d)
export HOME="$TMPHOME"
export TEST_KEY="sk-test-123"

echo "=== Testing profile commands ==="

echo "1. sage init"
./bin/sage init

echo ""
echo "2. sage profile list (empty)"
./bin/sage profile list

echo ""
echo "3. sage provider add openai --api-key-env=TEST_KEY"
./bin/sage provider add openai --api-key-env=TEST_KEY

echo ""
echo "4. sage profile add default --provider=openai --model=gpt-4o"
./bin/sage profile add default --provider=openai --model=gpt-4o

echo ""
echo "5. sage profile list (should show default)"
./bin/sage profile list

echo ""
echo "6. sage profile add fast --provider=openai --model=gpt-4o-mini"
./bin/sage profile add fast --provider=openai --model=gpt-4o-mini

echo ""
echo "7. sage profile list (should show two profiles)"
./bin/sage profile list

echo ""
echo "8. sage profile set-default fast"
./bin/sage profile set-default fast

echo ""
echo "9. sage profile list (fast should be default now)"
./bin/sage profile list

echo ""
echo "10. sage profile remove default"
./bin/sage profile remove default

echo ""
echo "11. sage profile list (only fast remaining)"
./bin/sage profile list

echo ""
echo "=== All profile tests passed ==="
rm -rf "$TMPHOME"
