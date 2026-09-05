#!/usr/bin/env bash
set -euo pipefail

echo "============================================================"
echo "DevSecOps Gate: Verifying GitHub Actions SHA Pinning"
echo "============================================================"

fail=0
for f in .github/workflows/*.yml; do
  if [[ ! -f "$f" ]]; then
    continue
  fi
  while IFS= read -r line; do
    ref=$(echo "$line" | sed -E 's/.*uses:\s*[^@]+@([^ #]+).*/\1/')
    # Skip local relative actions e.g. ./
    if [[ "$line" =~ uses:\s*\./ ]]; then
      continue
    fi
    if [[ ! "$ref" =~ ^[0-9a-f]{40}$ ]]; then
      echo "[UNPINNED] Action detected in $f:"
      echo "   Line: $line"
      fail=1
    fi
  done < <(grep -E '^\s*-?\s*uses:\s*[^./]' "$f" || true)
done

if [[ "$fail" -eq 0 ]]; then
  echo "[PASS] All workflow actions are strictly pinned to 40-char commit SHAs."
  exit 0
else
  echo "[FAIL] Unpinned workflow actions detected."
  exit 1
fi
