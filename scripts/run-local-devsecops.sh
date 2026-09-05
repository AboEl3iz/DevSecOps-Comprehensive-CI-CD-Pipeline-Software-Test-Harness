#!/usr/bin/env bash
set -euo pipefail

echo "============================================================"
echo "Running Complete DevSecOps Pipeline Suite (Local Execution)"
echo "============================================================"

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$PROJECT_ROOT"

echo ""
echo "--- Stage 1: Workflow Action Pinning Check ---"
bash scripts/test-action-pinning.sh

echo ""
echo "--- Stage 2: Dev Part - Go Unit Tests & Code Coverage ---"
(cd app && go test -v -cover ./...)

echo ""
echo "--- Stage 3: Dev Part - Integration Tests ---"
(cd app && go test -v -run TestIntegration_FullFlow ./...)

echo ""
echo "--- Stage 4: Static Code Analysis & Linting ---"
if command -v golangci-lint &>/dev/null; then
  (cd app && golangci-lint run)
else
  echo "[INFO] golangci-lint not installed locally, skipping local golangci-lint run."
fi

echo ""
echo "--- Stage 5: Secrets Scanning (Gitleaks) ---"
if command -v gitleaks &>/dev/null; then
  gitleaks detect --config config/.gitleaks.toml --verbose
else
  echo "[INFO] gitleaks not installed locally, skipping local secret scan."
fi

echo ""
echo "--- Stage 6: SAST Scan (Gosec) ---"
if command -v gosec &>/dev/null; then
  gosec ./app/...
else
  echo "[INFO] gosec not installed locally, skipping local SAST scan."
fi

echo ""
echo "--- Stage 7: Docker Container Hermetic Build ---"
if command -v docker &>/dev/null; then
  docker build -t security-ci-app:local ./app
  echo "[PASS] Local Docker container build completed successfully."
else
  echo "[INFO] docker not installed locally, skipping local container build."
fi

echo ""
echo "--- Stage 8: Immutable Image Reference Policy ---"
bash scripts/test-image-immutable-ref.sh

echo ""
echo "============================================================"
echo "ALL DEVSECOPS LOCAL GATES PASSED SUCCESSFULLY!"
echo "============================================================"
