#!/bin/sh
set -e

# Default values
STANDARD="${CPP_STANDARD:-c++20}"
SOURCE_FILE="${SOURCE_FILE:-/workspace/source.cpp}"
OUTPUT_FILE="${OUTPUT_FILE:-/tmp/output}"
TIMEOUT="${COMPILE_TIMEOUT:-25}"

# Compilation flags
# -Wall -Wextra: Enable most warnings
# -Wpedantic: Warn about non-standard C++
# -Wconversion: Warn about implicit type conversions
# -Wsign-conversion: Warn about sign conversions
# -Wshadow: Warn when variables shadow other variables
FLAGS="-std=${STANDARD} -Wall -Wextra -Wpedantic -Wconversion -Wsign-conversion -Wshadow"

# Check if source file exists
if [ ! -f "${SOURCE_FILE}" ]; then
    echo "Error: Source file ${SOURCE_FILE} not found" >&2
    exit 1
fi

# Compile with timeout
timeout ${TIMEOUT}s g++ ${FLAGS} -o "${OUTPUT_FILE}" "${SOURCE_FILE}"

# If we get here, compilation succeeded
echo "Compilation successful"
exit 0
