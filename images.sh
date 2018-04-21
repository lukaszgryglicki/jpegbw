#!/bin/sh
# For example Q=90 LO=5 HI=5 GA=1.11
if [ -z "$1" ]
then
  echo "$0: need to specify folder name"
  exit 1
fi
mkdir "$1" 2>/dev/null
rm images/bw_* 2>/dev/null
./jpegbw images/*
mv images/bw_* $1
