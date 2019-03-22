#!/bin/bash
LIB=./libjpegbw.so RC=1 GC=1 BC=1 RR=4 RG=-1 RB=-1 GR=-1 GG=4 GB=-1 BR=-1 BG=-1 BB=4 RLO=5 RHI=5 GLO=5 GHI=5 BLO=5 BHI=5 NA=1 RF="sin(3.1416*x1)" GF="sin(3.1416*x1)" BF="sin(3.1416*x1)" jpeg $*
