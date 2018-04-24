#!/bin/bash
for f in `find . -type f -iname "*.pgo"`
do
  if [ ! -z "$DEBUG" ]
  then
    echo "$1 \"$f\""
  fi
	$1 "$f" || exit 1
done
exit 0
