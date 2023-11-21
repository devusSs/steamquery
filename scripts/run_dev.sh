#!/bin/bash

# Detect build OS
BUILD_OS=$(uname -s | tr '[:upper:]' '[:lower:]')

# Detect build architecture
BUILD_ARCH=$(uname -m)

# Create testing directory
mkdir -p ./.testing 

# Copy the built binary to the testing directory
cp ./.release/steamquery_"${BUILD_OS}"_"${BUILD_ARCH}"/steamquery ./.testing/steamquery

# Run the application with specific parameters
./.testing/steamquery -l ./.logs_dev -c ./.config.json --no-update --debug