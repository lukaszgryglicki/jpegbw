#!/bin/bash
LIB="./libtet.so" X=400 Y=400 R0=-2 R1=2 I0=-2 I1=2 U="
101
|fz;r;0;255:0:0:255;exp((x2-.5)*4)-exp((x2-.5)*-4);0:0:0:0;1
|fz;i;0;0:0:255:255;exp((x2-.5)*4)-exp((x2-.5)*-4);0:0:0:0;1
|fz;m;0;0:255:0:255;exp((x2-.5)*8);0:0:0:0;1
" ./cmap out.gif "tet(5*x1)+tet(-5*x1)+tet(_5*x1)+tet(-_5*x1)"
