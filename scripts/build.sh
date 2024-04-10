#!/bin/bash

# Set the target folder
TARGET_FOLDER=".target"
rm -rf "$TARGET_FOLDER"    
mkdir -p "$TARGET_FOLDER"
# Build the Go application
go build -o "$TARGET_FOLDER/leapmailr" main.go
cp -r templates/ ./.target/templates
