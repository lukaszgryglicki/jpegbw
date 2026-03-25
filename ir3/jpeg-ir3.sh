#!/bin/bash
set -euo pipefail
JPEG_BIN="${JPEG_BIN:-/home/lgryglicki/go/bin/jpeg}"

TAIL="${TAIL:-4.0}"
IRS="${IRS:-0.30}"
IRL="${IRL:-0.55}"
IRSH="${IRSH:-0.12}"
IRGM="${IRGM:-0.90}"

RLO="${RLO:-1}"
RHI="${RHI:-1}"
GLO="${GLO:-1}"
GHI="${GHI:-1}"
BLO="${BLO:-1}"
BHI="${BHI:-1}"

exec env \
  IR3=1 \
  IRT="$TAIL" \
  IRS="$IRS" \
  IRL="$IRL" \
  IRSH="$IRSH" \
  IRGM="$IRGM" \
  RR=1 RG=0 RB=0 \
  GR=0 GG=1 GB=0 \
  BR=0 BG=0 BB=1 \
  RLO="$RLO" RHI="$RHI" \
  GLO="$GLO" GHI="$GHI" \
  BLO="$BLO" BHI="$BHI" \
  NA=1 \
  "$JPEG_BIN" "$@"
