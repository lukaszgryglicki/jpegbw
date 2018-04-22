#!/bin/bash
for f in $*
do
  LIB="jpegbw.so" F="alpha(x1, 6.28, -.1, .9)" jpegbw "$f"
  mv "bw_$f" "r_$f"
  LIB="jpegbw.so" F="alpha(x1, 6.48, 0., 1.)" jpegbw "$f"
  mv "bw_$f" "g_$f"
  LIB="jpegbw.so" F="alpha(x1, 6.68, .1, 1.1)" jpegbw "$f"
  mv "bw_$f" "b_$f"
  LIB="jpegbw.so" F="cbrt(vingette(1., x2, x3))" jpegbw "$f"
  :q
  mv "bw_$f" "a_$f"
  convert "r_$f" "g_$f" "b_$f" "a_$f" -combine "out_$f"
  rm "r_$f" "g_$f" "b_$f" "a_$f"
done
