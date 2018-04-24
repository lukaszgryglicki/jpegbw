#!/bin/bash
# USE DEFINE env variable to define constants
# For example DEFINE="debug info"
# It will uncomment lines that start with '// debug:' or '// info:'
# It will save abc.pgo -> abc.go
# path/to/def.pgo -> path/to/def.go
for f in `find . -iname "*.pgo"`
do
  ./gengo $f
done
