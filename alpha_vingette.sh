#!/bin/bash
LIB=./libjpegbw.so RR=1 RG=0 RB=0 GR=0 GG=1 GB=0 BR=0 BG=0 BB=1 AR=0.2125 AG=0.7154 AB=0.0721 RLO=2 RHI=2 GLO=2 GHI=2 BLO=2 BHI=2 ALO=2 AHI=2 AF="1-.35*cabs(2*x2-1_1)" ./jpeg in.png
