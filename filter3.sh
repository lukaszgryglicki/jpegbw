#!/bin/bash
#RR=2 RG=7 RB=1 GR=2 GG=7 GB=1 BR=2 BG=7 BB=1 RLO=2 RHI=70 GLO=35 GHI=35 BLO=70 BHI=2 RGA=1.2 BGA=1.2 GGA=1.2 NA=1 jpeg $*
RR=0.3 RG=0.5 RB=0.2 GR=0.3 GG=0.5 GB=0.2 BR=0.3 BG=0.5 BB=0.2 RLO=1 RHI=29 GLO=15 GHI=15 BLO=29 BHI=1 NA=1 LIB=libjpegbw.so RF="saturate(x1, .0001, .9999)" GF="saturate(x1, .0001, .9999)" BF="saturate(x1, .0001, .9999)" jpeg $*
