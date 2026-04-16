# ISOVAL

`ISOVAL` equalizes weighted RGB value/lightness while preserving color relationships as much as the selected mode allows.

Weighted value is:

```text
IVR*R + IVG*G + IVB*B
```

with `IVR/IVG/IVB` normalized internally.

## Modes

### `ISOVAL=add`

Add the same delta to all channels:

```text
delta = target - value
R' = clamp(R + delta)
G' = clamp(G + delta)
B' = clamp(B + delta)
```

### `ISOVAL=mul`

Multiply all channels by the same factor:

```text
scale = target / value
R' = clamp(R * scale)
G' = clamp(G * scale)
B' = clamp(B * scale)
```

Pixels with weighted value `0` are left unchanged.

### `ISOVAL=exp`

Exponential-space equalization:

```text
outC = IVBASE^(inC + shift)
```

with the same `shift` for `R`, `G`, `B` in a pixel, chosen so:

```text
IVR*R' + IVG*G' + IVB*B' = target
```

before clipping.

Equivalent closed form:

```text
br = IVBASE^R
bg = IVBASE^G
bb = IVBASE^B
denom = IVR*br + IVG*bg + IVB*bb
R' = clamp(br * target / denom)
G' = clamp(bg * target / denom)
B' = clamp(bb * target / denom)
```

## Parameters

### `IVR`, `IVG`, `IVB`

Weights for the equalized value.

Defaults:

```text
IVR=0.2126
IVG=0.7152
IVB=0.0722
```

### `IVT`

Manual target in `0..1`.

Default:

```text
IVT=0.33
```

Ignored when `IVTAUTO` is set.

### `IVTAUTO`

Automatic target selection.

Supported values:

- `avg`
- `med`
- `min`
- `max`
- `pNN`

Examples:

- `p25`
- `p50`
- `p75`
- `p90`
- `p99.5`

`med` is just an alias for `p50`.

### `IVCLIP`

Trim / outlier-ignore percentage used by `IVTAUTO`.

Default:

```text
IVCLIP=0
```

Range:

```text
0 <= IVCLIP < 50
```

Semantics:

- `avg`, `med`, `pNN`:
  discard `IVCLIP%` low tail and `IVCLIP%` high tail from the value histogram first
- `max`:
  ignore the most restrictive `IVCLIP%` of safe upper-bound pixels
- `min`:
  ignore the most restrictive `IVCLIP%` of safe lower-bound pixels in `add` mode

This makes `max` far more useful on real photos, because one rare saturated pixel no longer dictates the target.

### `IVBASE`

Only used by `ISOVAL=exp`.

Default:

```text
IVBASE=2.78
```

Must be greater than `1`.

## Practical examples

### Manual target

```bash
JPEG_BIN=./jpeg ISOVAL=add IVT=0.50 ./jpeg-isoval.sh in.png
JPEG_BIN=./jpeg ISOVAL=mul IVT=0.40 ./jpeg-isoval.sh in.png
JPEG_BIN=./jpeg ISOVAL=exp IVBASE=3.0 IVT=0.55 ./jpeg-isoval.sh in.png
```

### Average / median / percentile targets

```bash
JPEG_BIN=./jpeg IVTAUTO=avg ./jpeg-isoval.sh in.png
JPEG_BIN=./jpeg IVTAUTO=med ./jpeg-isoval.sh in.png
JPEG_BIN=./jpeg IVTAUTO=p75 ./jpeg-isoval.sh in.png
JPEG_BIN=./jpeg IVTAUTO=p90 IVCLIP=1 ./jpeg-isoval.sh in.png
```

### Safe edge targets

```bash
JPEG_BIN=./jpeg ISOVAL=add IVTAUTO=max ./jpeg-isoval.sh in.png
JPEG_BIN=./jpeg ISOVAL=add IVTAUTO=min ./jpeg-isoval.sh in.png
JPEG_BIN=./jpeg ISOVAL=mul IVTAUTO=max IVCLIP=0.5 ./jpeg-isoval.sh in.png
JPEG_BIN=./jpeg ISOVAL=exp IVBASE=2.0 IVTAUTO=max IVCLIP=1 ./jpeg-isoval.sh in.png
```

## Manual calls

```bash
ISOVAL=add IVTAUTO=p75 IVCLIP=1 IVR=0.2126 IVG=0.7152 IVB=0.0722 \
RR=1 RG=0 RB=0 GR=0 GG=1 GB=0 BR=0 BG=0 BB=1 NA=1 \
./jpeg in.png
```

```bash
ISOVAL=mul IVTAUTO=max IVCLIP=0.5 IVR=0.2126 IVG=0.7152 IVB=0.0722 \
RR=1 RG=0 RB=0 GR=0 GG=1 GB=0 BR=0 BG=0 BB=1 NA=1 \
./jpeg in.png
```

```bash
ISOVAL=exp IVBASE=2.0 IVTAUTO=p90 IVCLIP=1 IVR=0.2126 IVG=0.7152 IVB=0.0722 \
RR=1 RG=0 RB=0 GR=0 GG=1 GB=0 BR=0 BG=0 BB=1 NA=1 \
./jpeg in.png
```

## Verify with grayscale

To see how flat the result becomes under the same weights:

```bash
OGS=1 GSR=0.2126 GSG=0.7152 GSB=0.0722 \
RR=1 RG=0 RB=0 GR=0 GG=1 GB=0 BR=0 BG=0 BB=1 NA=1 \
./jpeg co_in.png
```

## Notes

1. `IVTAUTO=min` with `ISOVAL=mul` or `ISOVAL=exp` resolves to `0`.
2. `IVTAUTO=max` is usually the most useful auto for `mul` and `exp`.
3. `IVCLIP=0.1`, `0.5`, `1`, `2` are the most practical starting values.
4. `med` is equivalent to `p50`.

## Build

```bash
go build -o jpeg ./cmd/jpeg
```
