# Examples

```
JPEG_BIN=./jpeg ISOVAL=add IVT=0.50 ./jpeg-isoval.sh in.png
JPEG_BIN=./jpeg IVT=0.40 ./jpeg-isoval.sh in.png
JPEG_BIN=./jpeg ISOVAL=exp IVBASE=3.0 IVT=0.55 ./jpeg-isoval-mul.sh in.png
JPEG_BIN=./jpeg ISOVAL=add IVTAUTO=avg ./jpeg-isoval.sh in.png
JPEG_BIN=./jpeg IVTAUTO=avg ./jpeg-isoval.sh in.png
JPEG_BIN=./jpeg ISOVAL=add IVTAUTO=med ./jpeg-isoval.sh in.png
JPEG_BIN=./jpeg IVTAUTO=med ./jpeg-isoval.sh in.png
JPEG_BIN=./jpeg ISOVAL=add IVTAUTO=max ./jpeg-isoval.sh in.png
JPEG_BIN=./jpeg IVTAUTO=max ./jpeg-isoval.sh in.png
JPEG_BIN=./jpeg ISOVAL=exp IVBASE=2.0 IVTAUTO=max ./jpeg-isoval-mul.sh in.png
JPEG_BIN=./jpeg ISOVAL=add IVTAUTO=min ./jpeg-isoval.sh in.png
JPEG_BIN=./jpeg IVTAUTO=min ./jpeg-isoval.sh in.png
JPEG_BIN=./jpeg ISOVAL=exp IVBASE=2.0 IVTAUTO=min ./jpeg-isoval-mul.sh in.png
```


# Manual call

```
ISOVAL=add IVTAUTO=avg IVR=0.2126 IVG=0.7152 IVB=0.0722 RR=1 RG=0 RB=0 GR=0 GG=1 GB=0 BR=0 BG=0 BB=1 NA=1 ./jpeg in.png
ISOVAL=mul IVTAUTO=max IVR=0.2126 IVG=0.7152 IVB=0.0722 RR=1 RG=0 RB=0 GR=0 GG=1 GB=0 BR=0 BG=0 BB=1 NA=1 ./jpeg in.png
ISOVAL=exp IVBASE=2.0 IVTAUTO=med IVR=0.2126 IVG=0.7152 IVB=0.0722 RR=1 RG=0 RB=0 GR=0 GG=1 GB=0 BR=0 BG=0 BB=1 NA=1 ./jpeg in.png
```


# Verify

```
OGS=1 GSR=0.2126 GSG=0.7152 GSB=0.0722 RR=1 RG=0 RB=0 GR=0 GG=1 GB=0 BR=0 BG=0 BB=1 NA=1 ./jpeg co_in.png
```


# Typical

```
JPEG_BIN=./jpeg IISOVAL=add VTAUTO=avg ./jpeg-isoval.sh in.png
JPEG_BIN=./jpeg IVTAUTO=max ./jpeg-isoval.sh in.png
JPEG_BIN=./jpeg IISOVAL=add VTAUTO=min ./jpeg-isoval.sh in.png
```

# Note

1) Note that `IVTAUTO` mode `min` with `ISOVAL` mode `mul` or `exp` makes no sense.
2) `IVTAUTO` min actually means use such a IVT (Iso value target) value that fully prevents clipping minimal values (read - image will be rather too bright)
3) `IVTAUTO` max actually means use such a IVT (Iso value target) value that fully prevents clipping maximal values (read - image will be rather too dark, but usually better than in min mode)
