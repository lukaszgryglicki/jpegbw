#!/bin/bash
for f in $*
do
  LO=4 HI=4 R=0.25 G=0.6 B=0.15 LIB="libjpegbw.so" F="alpha(x1, 6.28, -.1, .9)" jpegbw "$f"
  mv "bw_$f" "r_$f"
  LO=4 HI=4 R=0.25 G=0.6 B=0.15 LIB="libjpegbw.so" F="alpha(x1, 6.48, 0., 1.)" jpegbw "$f"
  mv "bw_$f" "g_$f"
  LO=4 HI=4 R=0.25 G=0.6 B=0.15 LIB="libjpegbw.so" F="alpha(x1, 6.68, .1, 1.1)" jpegbw "$f"
  mv "bw_$f" "b_$f"
  LO=4 HI=4 R=0.25 G=0.6 B=0.15 LIB="libjpegbw.so" F="cbrt(vingette(1., x2, x3))" jpegbw "$f"
  mv "bw_$f" "a_$f"
  convert "r_$f" "g_$f" "b_$f" "a_$f" -combine "out_$f"
  rm "r_$f" "g_$f" "b_$f" "a_$f"
done
