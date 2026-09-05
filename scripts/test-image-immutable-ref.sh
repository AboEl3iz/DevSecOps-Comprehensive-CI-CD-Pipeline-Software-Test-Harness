#!/usr/bin/env bash
set -euo pipefail

echo "============================================================"
echo "DevSecOps Gate: Verifying Immutable Container Digests in K8s"
echo "============================================================"

fail=0
if [[ ! -d "k8s" ]]; then
  echo "[INFO] No k8s/ directory found. Checking policies/ for manifest examples..."
  target_dir="policies"
else
  target_dir="k8s"
fi

while IFS= read -r -d '' f; do
  while IFS= read -r img; do
    if [[ -n "$img" && "$img" != *"@sha256:"* ]]; then
      echo "[FAIL] Mutable image reference detected in $f: $img"
      fail=1
    fi
  done < <(grep -E '^\s*image:\s*' "$f" | awk '{print $2}' || true)
done < <(find "$target_dir" -name '*.yaml' -o -name '*.yml' -print0 2>/dev/null || true)

if [[ "$fail" -eq 0 ]]; then
  echo "[PASS] All deployment image references use immutable @sha256 digests."
  exit 0
else
  echo "[FAIL] Mutable image tags found in Kubernetes manifests."
  exit 1
fi
