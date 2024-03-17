#!/bin/bash
if ( [ -z "$RR" ] || [ -z "$RG" ] || [ -z "$RB" ] )
then
  export RR=1
  export RG=0
  export RB=0
fi
if ( [ -z "$GR" ] || [ -z "$GG" ] || [ -z "$GB" ] )
then
  export GR=0
  export GG=1
  export GB=0
fi
if ( [ -z "$BR" ] || [ -z "$BG" ] || [ -z "$BB" ] )
then
  export BR=0
  export BG=0
  export BB=1
fi
if ( [ -z "$RLO" ] || [ -z "$RHI" ] )
then
  export RLO=1
  export RHI=1
fi
if ( [ -z "$GLO" ] || [ -z "$GHI" ] )
then
  export GLO=1
  export GHI=1
fi
if ( [ -z "$BLO" ] || [ -z "$BHI" ] )
then
  export BLO=1
  export BHI=1
fi
if [ -z "$RX"]
then
  export RX="1-x1"
fi
if [ -z "$RPE" ]
then
  export RPE=3.1415926
fi
if [ -z "$ROF" ]
then
  export ROF=0
fi
if [ -z "$RPO" ]
then
  export RPO=1
fi
if [ -z "$GX"]
then
  export GX="1-x1"
fi
if [ -z "$GPE" ]
then
  export GPE=3.1415926
fi
if [ -z "$GOF" ]
then
  export GOF=0
fi
if [ -z "$GPO" ]
then
  export GPO=1
fi
if [ -z "$BX"]
then
  export BX="1-x1"
fi
if [ -z "$BPE" ]
then
  export BPE=3.1415926
fi
if [ -z "$BOF" ]
then
  export BOF=0
fi
if [ -z "$BPO" ]
then
  export BPO=1
fi
# double complex alpha(double complex arg, double complex period, double complex offset, double complex power) {
echo "LIB=\"libjpegbw.so\" RC=1 GC=1 BC=1 NA=1 RF=\"alpha($RX, $RPE, $ROF, $RPO)\" GF=\"alpha($GX, $GPE, $GOF, $GPO)\" BF=\"alpha($BX, $BPE, $BOF, $BPO)\" jpeg [args]"
LIB="libjpegbw.so" RC=1 GC=1 BC=1 NA=1 RF="alpha($RX, $RPE, $ROF, $RPO)" GF="alpha($GX, $GPE, $GOF, $GPO)" BF="alpha($BX, $BPE, $BOF, $BPO)" jpeg $*
