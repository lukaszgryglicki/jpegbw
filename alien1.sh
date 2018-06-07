#!/bin/bash
LIB=./libjpegbw.so RR=2 RG=1 RB=1 GR=1 GG=2 GB=1 BR=1 BG=1 BB=2 RLO=3 RHI=3 GLO=3 GHI=3 BLO=3 BHI=3 NA=1 RF="sin(3.1416*x1)" GF="sin(3.1416*x1)" BF="sin(3.1416*x1)" ./jpeg $*
