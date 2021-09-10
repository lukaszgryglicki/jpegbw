#!/bin/bash
LIB=./libjpegbw.so INF=200 ACM='' RC=1 GC=1 BC=1 RR=1 RG=0 RB=0 GR=0 GG=1 GB=0 BR=0 BG=0 BB=1 RLO=1 RHI=1 GLO=1 GHI=1 BLO=1 BHI=1 AR=0.2125 AG=0.7154 AB=0.0721 ALO=1 AHI=1 AF="1-.4*cabs(2*x2-1_1)" RF="sin(3.1416*x1)" GF="sin(3.1416*x1)" BF="sin(3.1416*x1)" jpeg $*
