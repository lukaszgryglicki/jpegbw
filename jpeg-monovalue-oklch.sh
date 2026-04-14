#!/usr/bin/env bash
set -euo pipefail
JPEG_BIN="${JPEG_BIN:-/home/lgryglicki/go/bin/jpeg}"
MVT="${MVT:-0.72}"
MVC="${MVC:-}"
MVGAMUT="${MVGAMUT:-fit}"
RLO="${RLO:-1.5}"
RHI="${RHI:-1.5}"
GLO="${GLO:-1.5}"
GHI="${GHI:-1.5}"
BLO="${BLO:-1.5}"
BHI="${BHI:-1.5}"
exec env \
  MONOVAL=oklch MVT="$MVT" MVC="$MVC" MVGAMUT="$MVGAMUT" \
  RR=1 RG=0 RB=0 \
  GR=0 GG=1 GB=0 \
  BR=0 BG=0 BB=1 \
  RLO="$RLO" RHI="$RHI" \
  GLO="$GLO" GHI="$GHI" \
  BLO="$BLO" BHI="$BHI" \
  NA=1 \
  "$JPEG_BIN" "$@"
