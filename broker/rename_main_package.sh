#!/bin/bash

# Find all Go files with 'package main' at the top
# and replace it with 'package broker'
find . -type f -name "*.go" | while read -r file; do
    if grep -q '^package main' "$file"; then
        echo "Updating $file"
        sed -i 's/^package main/package broker/' "$file"
    fi
done

echo "All matching files have been updated to package broker."
