#!/bin/bash
if [ -z "$NF" ]
then
  echo "$0: lease specify number of frames with NF=n"
  exit 1
fi
if ( [ -z "$1" ] || [ -z "$2" ] )
then
  echo "$0: needs 2 arguments: filename.gif and function definition"
  exit 1
fi
if [ -z "$IR" ]
then
  IR=`echo "scale=8;1/${NF}" | bc`
  IR="x1+$IR*x2"
fi
if [ -z "$II" ]
then
  II=`echo "scale=8;1/${NF}" | bc`
  II="x1+$II*x2"
fi
if [ -z "$IM" ]
then
  IM=`echo "scale=8;1/${NF}" | bc`
  IM="x1+$IM*x2"
fi
if [ -z "$R" ]
then
  R=0
fi
if [ -z "$I" ]
then
  I=0
fi
if [ -z "$M" ]
then
  M=0
fi
# fz;rim;v;rC:rG:rb:cA;vinc;ciR:ciB:ciG:ciA;lh
LIB="./libtet.so" U="
${NF}
|fz;r;${R};255:0:0:255;${IR};0:0:0:0;1
|fz;i;${I};0:0:255:255;${II};0:0:0:0;1
|fz;m;${M};0:255:0:255;${IM};0:0:0:0;1
|z;r;0;0:0:0:255;x1;0:0:0:0;0
|z;i;0;0:0:0:255;x1;0:0:0:0;0
|z;m;1;0:0:0:255;x1;0:0:0:0;0
" ./cmap "$1" "$2"
