#!/usr/bin/env bash
set -euo pipefail
JPEG_BIN="${JPEG_BIN:-/home/lgryglicki/go/bin/jpeg}"
TR="${TR:-0.2126}"
TG="${TG:-0.7152}"
TB="${TB:-0.0722}"
RLO="${RLO:-1.5}"
RHI="${RHI:-1.5}"
GLO="${GLO:-1.5}"
GHI="${GHI:-1.5}"
BLO="${BLO:-1.5}"
BHI="${BHI:-1.5}"
T0R="${T0R:-0.10}"
T0G="${T0G:-0.06}"
T0B="${T0B:-0.02}"
T1R="${T1R:-1.00}"
T1G="${T1G:-0.95}"
T1B="${T1B:-0.82}"
exec env \
  RR="$TR" RG="$TG" RB="$TB" \
  GR="$TR" GG="$TG" GB="$TB" \
  BR="$TR" BG="$TG" BB="$TB" \
  RLO="$RLO" RHI="$RHI" \
  GLO="$GLO" GHI="$GHI" \
  BLO="$BLO" BHI="$BHI" \
  RF="$T0R+x1*($T1R-$T0R)" \
  GF="$T0G+x1*($T1G-$T0G)" \
  BF="$T0B+x1*($T1B-$T0B)" \
  NA=1 \
  "$JPEG_BIN" "$@"
