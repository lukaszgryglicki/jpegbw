#!/bin/bash
LIB="libjpegbw.so" R0=-3.5 R1=3.5 I0=-3.5 I1=3.5 X=500 Y=500 U="
101;
z,  r, 0,    0:  0:  0   :255,  0,     0 :0 :0  :0;
z,  i, 0,    0:  0:  0   :255,  0,     0 :0 :0  :0;
z,  m, 1,    0:  0:  0   :255,  0,     0 :0 :0  :0;
fz, r, 3.5,  255:  0:  0 :255, -0.07,  -1:1 :1  :0;
fz, i, 3.5,  0  :  0:255 :255, -0.07,  1 :1 :-1 :0;
fz, m, 3.5,  0  :255:  0 :255, -0.035, 1 :-1:1  :0
" ./cmap out.gif "csin(x1)"