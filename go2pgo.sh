for f in `find . -iname "*.go"`
do
  echo ${f//\.go/\.pgo}
done
