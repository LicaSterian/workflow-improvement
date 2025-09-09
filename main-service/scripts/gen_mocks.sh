#!/usr/bin/env bash
set -euo pipefail

# Must be run FROM main-service (Makefile cd's here)
SERVICE_ROOT="${SERVICE_ROOT:-$(pwd)}"
OUT_DIR="${OUT_DIR:-$SERVICE_ROOT/mocks}"

if ! command -v mockgen >/dev/null 2>&1; then
  echo "error: mockgen not found. Install with: go install go.uber.org/mock/mockgen@latest" >&2
  exit 1
fi

# Clean re-gen
rm -rf "$OUT_DIR"
mkdir -p "$OUT_DIR"

# Only newline word splitting
OLDIFS=$IFS
IFS=$'\n'

# Determine module path (e.g., github.com/acme/main-service)
MODULE_PATH="$(go list -m -f '{{.Path}}')"

# List all pkgs (omit already-generated mocks)
PKGS="$(go list ./... | grep -v '/mocks$' || true)"
if [ -z "$PKGS" ]; then
  echo "No packages found under $SERVICE_ROOT"
  IFS=$OLDIFS; exit 0
fi

for PKG in $PKGS; do
  # Optional: skip binaries
  PKG_NAME="$(go list -f '{{.Name}}' "$PKG")"
  if [ "$PKG_NAME" = "main" ]; then
    echo "Skipping $PKG (package main)."
    continue
  fi

  DIR="$(go list -f '{{.Dir}}' "$PKG")"
  # Collect non-test sources
  SOURCES="$(find "$DIR" -maxdepth 1 -type f -name '*.go' ! -name '*_test.go' | sort || true)"
  if [ -z "$SOURCES" ]; then
    echo "Skipping $PKG (no non-test .go files)."
    continue
  fi

  # Compute import-subpath relative to module (e.g., "internal/foo/bar")
  # If PKG == MODULE_PATH, REL becomes empty (package at module root)
  REL="${PKG#$MODULE_PATH/}"
  [ "$REL" = "$PKG" ] && REL=""  # handle root package exactly

  DEST_DIR="$OUT_DIR/$REL"
  mkdir -p "$DEST_DIR"
  DEST_FILE="$DEST_DIR/mocks.go"
  OUT_PKG="${PKG_NAME}_mocks"   # always a valid identifier

  echo "Generating mocks for $PKG -> $DEST_FILE"

  # Build args without arrays (macOS Bash 3.2)
  ARGS=""
  while IFS= read -r f; do
    [ -z "$f" ] && continue
    ARGS="$ARGS -source \"${f}\""
  done <<< "$SOURCES"

  # shellcheck disable=SC2086
  eval "mockgen -destination \"${DEST_FILE}\" -package \"${OUT_PKG}\" ${ARGS}"
done

IFS=$OLDIFS
echo "Done. Mocks written under: $OUT_DIR"
