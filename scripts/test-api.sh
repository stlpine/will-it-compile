#!/bin/bash
# Test script for the will-it-compile API

set -e

API_URL="${API_URL:-http://localhost:8080}"

echo "Testing will-it-compile API at ${API_URL}"
echo "========================================="

# Test 1: Health check
echo ""
echo "Test 1: Health check"
curl -s "${API_URL}/health" | jq .

# Test 2: Get supported environments
echo ""
echo "Test 2: Get supported environments"
curl -s "${API_URL}/api/v1/environments" | jq .

# Test 3: Compile valid C++ code
echo ""
echo "Test 3: Compile valid C++ code (Hello World)"
SOURCE_CODE=$(echo '#include <iostream>
int main() {
    std::cout << "Hello, World!" << std::endl;
    return 0;
}' | base64)

RESPONSE=$(curl -s -X POST "${API_URL}/api/v1/compile" \
  -H "Content-Type: application/json" \
  -d "{
    \"code\": \"${SOURCE_CODE}\",
    \"language\": \"cpp\",
    \"compiler\": \"gcc-13\",
    \"standard\": \"c++20\"
  }")

echo "$RESPONSE" | jq .
JOB_ID=$(echo "$RESPONSE" | jq -r .job_id)

# Wait for compilation
echo ""
echo "Waiting for compilation to complete..."
sleep 3

# Test 4: Get compilation result
echo ""
echo "Test 4: Get compilation result"
curl -s "${API_URL}/api/v1/compile/${JOB_ID}" | jq .

# Test 5: Compile invalid C++ code
echo ""
echo "Test 5: Compile invalid C++ code (syntax error)"
SOURCE_CODE=$(echo '#include <iostream>
int main() {
    std::cout << "Missing semicolon"
    return 0;
}' | base64)

RESPONSE=$(curl -s -X POST "${API_URL}/api/v1/compile" \
  -H "Content-Type: application/json" \
  -d "{
    \"code\": \"${SOURCE_CODE}\",
    \"language\": \"cpp\",
    \"compiler\": \"gcc-13\",
    \"standard\": \"c++20\"
  }")

echo "$RESPONSE" | jq .
JOB_ID=$(echo "$RESPONSE" | jq -r .job_id)

# Wait for compilation
echo ""
echo "Waiting for compilation to complete..."
sleep 3

# Get result
echo ""
echo "Getting result (should show compilation error)..."
curl -s "${API_URL}/api/v1/compile/${JOB_ID}" | jq .

echo ""
echo "========================================="
echo "Tests completed!"
