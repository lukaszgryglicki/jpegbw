#!/bin/bash
for f in $*
do
  LO=4 HI=4 R=1 G=0 B=0 LIB="libjpegbw.so" F="toon(x1, 2)" jpegbw "$f"
  mv "bw_$f" "r_$f"
  LO=4 HI=4 R=0 G=1 B=0 LIB="libjpegbw.so" F="toon(x1, 2)" jpegbw "$f"
  mv "bw_$f" "g_$f"
  LO=4 HI=4 R=0 G=0 B=1 LIB="libjpegbw.so" F="toon(x1, 2)" jpegbw "$f"
  mv "bw_$f" "b_$f"
  convert "r_$f" "g_$f" "b_$f" -combine "out_$f"
  rm "r_$f" "g_$f" "b_$f" 
done
