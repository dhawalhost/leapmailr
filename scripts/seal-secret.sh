#!/usr/bin/env bash
#
# seal-secret.sh
#
# Seals a plaintext Kubernetes Secret manifest (YAML) using kubeseal,
# producing a SealedSecret that can be safely committed to git.
#
# Usage:
#   ./scripts/seal-secret.sh <plaintext-secret.yaml> <output-sealedsecret.yaml>
#
# Prerequisites:
#   - kubeseal CLI installed (https://github.com/bitnami-labs/sealed-secrets)
#   - kubectl context set to the target cluster
#   - Sealed Secrets controller deployed in the cluster

set -euo pipefail

PLAINTEXT_SECRET="${1:-}"
OUTPUT_SEALED="${2:-}"

if [ -z "$PLAINTEXT_SECRET" ] || [ -z "$OUTPUT_SEALED" ]; then
  echo "Usage: $0 <plaintext-secret.yaml> <output-sealedsecret.yaml>"
  echo ""
  echo "Example:"
  echo "  $0 /tmp/leapmailr-backend-secret.yaml k8s/leapmailr-backend-sealedsecret.yaml"
  exit 1
fi

if [ ! -f "$PLAINTEXT_SECRET" ]; then
  echo "Error: plaintext secret file not found: $PLAINTEXT_SECRET"
  exit 1
fi

# Optional: override controller namespace/name if needed
CONTROLLER_NS="${SEALED_SECRETS_CONTROLLER_NAMESPACE:-kube-system}"
CONTROLLER_NAME="${SEALED_SECRETS_CONTROLLER_NAME:-sealed-secrets-controller}"

echo "Sealing secret using controller:"
echo "  Namespace: $CONTROLLER_NS"
echo "  Name:      $CONTROLLER_NAME"
echo ""

kubeseal \
  --controller-namespace "$CONTROLLER_NS" \
  --controller-name "$CONTROLLER_NAME" \
  --format yaml \
  < "$PLAINTEXT_SECRET" \
  > "$OUTPUT_SEALED"

echo "✓ Sealed secret written to: $OUTPUT_SEALED"
echo ""
echo "You can now safely commit this file to git."
echo "Apply it with: kubectl apply -f $OUTPUT_SEALED"
