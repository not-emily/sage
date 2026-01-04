#!/bin/bash
# Test provider commands
set -e

TMPHOME=$(mktemp -d)
export HOME="$TMPHOME"
export TEST_KEY="sk-test-123"

echo "=== Testing provider commands ==="

echo "1. sage init"
./bin/sage init

echo ""
echo "2. sage provider list (empty)"
./bin/sage provider list

echo ""
echo "3. sage provider add openai --api-key-env=TEST_KEY"
./bin/sage provider add openai --api-key-env=TEST_KEY

echo ""
echo "4. sage provider list (should show openai)"
./bin/sage provider list

echo ""
echo "5. sage provider add openai --account=work --api-key-env=TEST_KEY"
./bin/sage provider add openai --account=work --api-key-env=TEST_KEY

echo ""
echo "6. sage provider list (should show two accounts)"
./bin/sage provider list

echo ""
echo "7. sage provider remove openai --account=work"
./bin/sage provider remove openai --account=work

echo ""
echo "8. sage provider list (should show one account)"
./bin/sage provider list

echo ""
echo "=== All provider tests passed ==="
rm -rf "$TMPHOME"
