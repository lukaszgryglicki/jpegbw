#!/bin/bash
fr=-4
cr=8
nr=11
ir=0
fi=-4
ci=8
ni=11
ii=0
echo 'r,i,m,fr,fi,fm' > out.csv
while true
do
  zr=`echo "scale=10; $fr+($ir*$cr)/($nr-1)" | bc`
  while true
  do
    zi=`echo "scale=10; $fi+($ii*$ci)/($ni-1)" | bc`
    if [ ${zi:0:1} == "-" ]
    then
      i="-_${zi:1}"
    else
      i="_$zi"
    fi
    echo "$zr$i"
    LIB=libtet.so ./f "$1" "${zr}_${zi}" 2>>out.csv 1>/dev/null
    ii=$(( ii + 1 ))
    if [ "$ii" = "$ni" ]
    then
      ii=0
      break
    fi
  done
  ir=$(( ir + 1 ))
  if [ "$ir" = "$nr" ]
  then
    break
  fi
done
