#!/usr/bin/env bash
set -euo pipefail

if [ "$#" -ne 3 ]; then
  echo "Usage: $0 input1 input2 output"
  exit 1
fi

in1="$1"
in2="$2"
out="$3"

magick "$in1" "$in2" \
  -compose Multiply -composite \
  -evaluate Pow 0.5 \
  "$out"
