#!/usr/bin/env bash
# ISOVAL=add|mul|exp - default add
# IVTAUTO=avg|med|min|max|pNN - default p40, "-" to unset
# IVCLIP=val (0-50), default 1.5, "-" to unset
# IVT=val (0-100)
# IVBASE=val (1.001-1000.0)

if [ -z "${IVTAUTO:-}" ]; then
  export IVTAUTO=p40
fi
if [ "${IVTAUTO:-}" = "-" ]; then
  unset IVTAUTO
fi

if [ -z "${IVCLIP:-}" ]; then
  export IVCLIP=1.5
fi
if [ "${IVCLIP:-}" = "-" ]; then
  unset IVCLIP
fi

if [ -z "${ISOVAL:-}" ]; then
  export ISOVAL=add
fi
if [ -z "${IVR:-}" ]; then
  export IVR=0.2126
fi
if [ -z "${IVG:-}" ]; then
  export IVG=0.7152
fi
if [ -z "${IVB:-}" ]; then
  export IVB=0.0722
fi

RR=1 RG=0 RB=0 GR=0 GG=1 GB=0 BR=0 BG=0 BB=1 NA=1 jpeg "$@"
