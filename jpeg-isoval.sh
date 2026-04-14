#!/usr/bin/env bash
# ISOVAL=add|mul|exp
# IVTAUTO=avg|med|max|min
set -euo pipefail
JPEG_BIN="${JPEG_BIN:-./jpeg}"
ISOVAL="${ISOVAL:-mul}"
IVT="${IVT:-0.33}"
IVTAUTO="${IVTAUTO:-}"
IVBASE="${IVBASE:-2.78}"
IVR="${IVR:-0.2126}"
IVG="${IVG:-0.7152}"
IVB="${IVB:-0.0722}"
exec env \
  ISOVAL="$ISOVAL" IVT="$IVT" IVTAUTO="$IVTAUTO" IVBASE="$IVBASE" IVR="$IVR" IVG="$IVG" IVB="$IVB" \
  RR=1 RG=0 RB=0 \
  GR=0 GG=1 GB=0 \
  BR=0 BG=0 BB=1 \
  NA=1 \
  "$JPEG_BIN" "$@"
