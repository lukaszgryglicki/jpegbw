# Examples

```
JPEG_BIN=./jpeg ./jpeg-isoval-add.sh in.png
JPEG_BIN=./jpeg IVT=0.40 ./jpeg-isoval-mul.sh in.png
```


# Manual call

```
ISOVAL=add IVT=0.50 IVR=0.2126 IVG=0.7152 IVB=0.0722 \
RR=1 RG=0 RB=0 GR=0 GG=1 GB=0 BR=0 BG=0 BB=1 NA=1 \
./jpeg in.png
```


# Verify

```
OGS=1 GSR=0.2126 GSG=0.7152 GSB=0.0722 \
RR=1 RG=0 RB=0 GR=0 GG=1 GB=0 BR=0 BG=0 BB=1 NA=1 \
./jpeg co_in.png
```
