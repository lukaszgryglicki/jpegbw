#!/bin/bash
LIB="libjpegbw.so" RLO=4 RHI=4 RR=0.25 RG=0.6 RB=0.15 GLO=4 GHI=4 GR=0.25 GG=0.6 GB=0.15 BLO=4 BHI=4 BR=0.25 BG=0.6 BB=0.15 ALO=4 AHI=4 AR=0.25 AG=0.6 AB=0.15 RF="alpha(x1, 6.28, -.1, .9)" GF="alpha(x1, 6.48, 0., 1.)" BF="alpha(x1, 6.68, .1, 1.1)" AF="cbrt(vingette(1., x2, x3))" jpeg $*
