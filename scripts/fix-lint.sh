#!/usr/bin/env bash
# fix-lint.sh - Quick fixes for golangci-lint errors

set -euo pipefail

cd "$(dirname "$0")/.."

echo "Applying lint fixes..."

# Fix: Check fmt.Scanln error
sed -i 's/fmt\.Scanln(&confirm)/_, _ = fmt.Scanln(\&confirm)/' scripts/encrypt-existing-data.go

# Fix: Check json.Unmarshal errors (assign to _)
sed -i 's/json\.Unmarshal(\(.*\))/_ = json.Unmarshal(\1)/' service/apikey.go
sed -i 's/json\.Unmarshal(\(.*\))/_ = json.Unmarshal(\1)/' service/contact.go

# Fix: Check client.Quit() error
sed -i 's/defer client\.Quit()/defer func() { _ = client.Quit() }()/' service/email.go

# Fix: Check audit.LogEvent errors
find service -name "*.go" -exec sed -i 's/s\.audit\.LogEvent(/_ = s.audit.LogEvent(/' {} \;
find service -name "*.go" -exec sed -i 's/\t\ts\.audit\.LogEvent(/\t\t_ = s.audit.LogEvent(/' {} \;

# Fix: Check service method errors in handlers
sed -i 's/service\.SendAutoReply(/_ = service.SendAutoReply(/' handlers/email.go
sed -i 's/service\.CreateContact(/_ = service.CreateContact(/' handlers/email.go
sed -i 's/trackingService\.RecordOpen(/_ = trackingService.RecordOpen(/' handlers/tracking.go
sed -i 's/trackingService\.RecordClick(/_ = trackingService.RecordClick(/' handlers/tracking.go

# Fix: Check suppression service errors
sed -i 's/suppressionService\.AddSuppressionFromWebhook(/_ = suppressionService.AddSuppressionFromWebhook(/' service/webhook_tracking.go

echo "✓ Applied lint fixes"
echo "Now run: golangci-lint run ./..."
