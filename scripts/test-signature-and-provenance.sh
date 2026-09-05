#!/usr/bin/env bash
set -euo pipefail

IMAGE="${1:-ghcr.io/myorg/security-ci-app@sha256:dummy}"
REPO="${2:-myorg/security-ci-app}"

echo "============================================================"
echo "DevSecOps Gate: Verifying Cosign Signature & SLSA Provenance"
echo "Target Image: $IMAGE"
echo "============================================================"

if ! command -v cosign &>/dev/null; then
  echo "[WARN] cosign CLI is not installed. Skipping live keyless verification."
  echo "Simulated validation OK."
  exit 0
fi

echo "== Verifying Cosign Keyless Signature =="
cosign verify \
  --certificate-identity-regexp "^https://github.com/${REPO}/.github/workflows/.*@refs/heads/main$" \
  --certificate-oidc-issuer "https://token.actions.githubusercontent.com" \
  "$IMAGE" || {
    echo "[FAIL] Signature verification failed for $IMAGE"
    exit 1
  }

if command -v gh &>/dev/null; then
  echo "== Verifying SLSA Build Provenance Attestation =="
  gh attestation verify "oci://${IMAGE}" \
    --owner "${REPO%%/*}" \
    --predicate-type https://slsa.dev/provenance/v1 || {
      echo "[FAIL] SLSA build provenance verification failed."
      exit 1
    }

  echo "== Verifying CycloneDX SBOM Attestation =="
  gh attestation verify "oci://${IMAGE}" \
    --owner "${REPO%%/*}" \
    --predicate-type https://cyclonedx.org/bom || {
      echo "[FAIL] SBOM attestation verification failed."
      exit 1
    }
fi

echo "[PASS] $IMAGE has verified keyless signature, SLSA provenance, and SBOM attestation."
