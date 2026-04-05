#!/usr/bin/env bash
# Check which markdown source files are out of sync with their HTML counterparts.
# Run from the repo root.

set -euo pipefail

MANIFEST="SYNC-MANIFEST.sha256"
STALE=0

echo "Checking markdown → HTML sync..."
echo ""

while IFS= read -r line; do
  # Skip comments and blank lines
  [[ "$line" =~ ^#.*$ || -z "$line" ]] && continue

  hash=$(echo "$line" | awk '{print $1}')
  md_path=$(echo "$line" | awk '{print $2}')
  site_path=$(echo "$line" | sed 's/.*→ //' | sed 's/ \[.*$//')

  if [ ! -f "$md_path" ]; then
    echo "  MISSING  $md_path"
    STALE=$((STALE + 1))
    continue
  fi

  current_hash=$(shasum -a 256 "$md_path" | awk '{print $1}')

  if [ "$current_hash" != "$hash" ]; then
    echo "  CHANGED  $md_path"
    STALE=$((STALE + 1))
  elif [ ! -f "$site_path" ]; then
    echo "  NO HTML  $md_path → $site_path"
    STALE=$((STALE + 1))
  else
    echo "  OK       $md_path"
  fi
done < "$MANIFEST"

echo ""
if [ $STALE -eq 0 ]; then
  echo "All files in sync."
else
  echo "$STALE file(s) need attention."
fi
