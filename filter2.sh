#!/bin/bash
#RR=2 RG=7 RB=1 GR=2 GG=7 GB=1 BR=2 BG=7 BB=1 RLO=2 RHI=70 GLO=35 GHI=35 BLO=70 BHI=2 RGA=1.2 BGA=1.2 GGA=1.2 NA=1 jpeg $*
RR=1 RG=1 RB=1 GR=1 GG=1 GB=1 BR=1 BG=1 BB=1 RLO=0 RHI=70 GLO=30 GHI=30 BLO=70 BHI=0 NA=1 LIB=libjpegbw.so RF="saturate(x1, .0001, .9999)" GF="saturate(x1, .0001, .9999)" BF="saturate(x1, .0001, .9999)" jpeg $*
