for f in `find . -iname "*.go"`
do
  cp "$f" "${f//\.go/.pgo}"
done
