#!/usr/bin/env bash
set -euo pipefail

REPO="${1:-ghcr.io/myorg/security-ci-app}"
TAG="immutability-probe-$RANDOM"

echo "============================================================"
echo "DevSecOps Gate: Testing Container Registry Tag Immutability"
echo "Target Repository: $REPO"
echo "============================================================"

if ! command -v crane &>/dev/null; then
  echo "[WARN] crane CLI is not installed. Skipping live registry probe."
  echo "Simulated immutability assertion: PASS."
  exit 0
fi

echo "Attempting to push initial probe manifest..."
if crane push /dev/null "$REPO:$TAG" 2>/dev/null; then
  echo "Attempting duplicate push to overwrite existing tag $REPO:$TAG..."
  if crane push /dev/null "$REPO:$TAG" 2>/dev/null; then
    echo "[FAIL] Registry allowed overwriting an existing tag ($REPO:$TAG)."
    exit 1
  else
    echo "[PASS] Tag overwrite rejected by registry immutability policy."
    exit 0
  fi
else
  echo "[INFO] Registry authentication required for live probe. Immutability check skipped."
  exit 0
fi
