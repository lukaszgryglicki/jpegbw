# Mono-value and monotone helpers for `jpeg`

This patch adds a **final mono-value stage** to `cmd/jpeg/jpeg.go`.
It runs **after** the normal channel mapping, histogram stretch, optional custom functions,
contour processing, and optional IR3 mapping.

That means you can keep using the existing `RR/RG/RB ... BR/BG/BB`, `RLO/RHI`, `RF/GF/BF`,
`IR3`, and similar controls, then flatten value/lightness as a final display transform.

## New environment variables

### `MONOVAL` / `MVMODE`

Enable mono-value mode. Supported values:

- `luma` - flatten the same weighted channel-value measure that `jpeg`/`jpegbw` grayscale uses.
  This is the most practical *dual of grayscale* inside this tool.
- `linear` - flatten **linear RGB luminance** using `MVR/MVG/MVB` weights.
  This is the more colorimetric variant.
- `hsv` - keep hue and saturation, set **HSV V** to `MVT`.
- `hsl` - keep hue and saturation, set **HSL L** to `MVT`.
- `oklch` - keep OKLCh hue/chroma, set **OK lightness** to `MVT`.
  This is the most perceptually balanced option.

### `MVT`

Target flattened value/lightness/luminance, range `0..1`.
Default: `0.5`.

### `MVR`, `MVG`, `MVB`

Weights for `luma` and `linear` modes.
Default: `0.2126 / 0.7152 / 0.0722`.
They are normalized internally.

### `MVGAMUT`

How to handle out-of-gamut colors.

- `fit` - reduce chroma until the result fits the output gamut.
- `clip` - hard clip channels to `[0,1]`.

Default: `fit`.

### `MVZERO`

Only used by `luma` and `linear`.
Controls what happens for pixels with effectively zero value/luminance,
where hue/chromaticity is undefined.

- `gray` - send them to neutral gray at `MVT`
- `black` - keep them black

Default: `gray`.

### `MVS`

Optional saturation override for `hsv` and `hsl`.
If unset, the original saturation is preserved.
If set, it becomes a **hue-only** style transform in that model.

### `MVC`

Optional chroma override for `oklch`.
If unset, original chroma is preserved.
If set, it becomes a **hue-only** style transform in OKLCh.

## Practical guidance

### Exact dual of your current grayscale pipeline

Use `MONOVAL=luma`.

If you later grayscale with the **same weights**, the result is flat gray.
Inside this tool that means using the same weights for both:

- mono-value flattening: `MVR/MVG/MVB`
- grayscale output: `GSR/GSG/GSB`

Example verification:

```bash
MVR=0.2126 MVG=0.7152 MVB=0.0722 \
GSR=0.2126 GSG=0.7152 GSB=0.0722 \
MONOVAL=luma MVT=0.50 OGS=1 ./jpeg in.jpg
```

That should produce an almost perfectly flat gray image.
Any remaining variation will come from gamut fitting, clipping choices, or later stages you add.

### More colorimetric version

Use `MONOVAL=linear`.

This keeps linear-RGB chromaticity and flattens linear luminance.
It is more physically grounded, but it is **not** the exact dual of the tool's default grayscale math.

### Perceptual version

Use `MONOVAL=oklch`.

This usually gives the nicest visual balance because it flattens a perceptual lightness axis.
Set `MVC` when you want stronger posterized hue-only output.

## Included helper scripts

### `jpeg-monovalue-luma.sh`
Tool-space exact dual of grayscale.

```bash
./jpeg-monovalue-luma.sh in.jpg
MVT=0.35 ./jpeg-monovalue-luma.sh in.jpg
```

### `jpeg-monovalue-linear.sh`
Linear luminance flatten.

```bash
MVT=0.45 MVGAMUT=fit ./jpeg-monovalue-linear.sh in.jpg
```

### `jpeg-monovalue-hsv.sh`
Flatten HSV value.

```bash
./jpeg-monovalue-hsv.sh in.jpg
MVT=0.75 MVS=1 ./jpeg-monovalue-hsv.sh in.jpg
```

With `MVS=1`, this becomes a strong hue-only look in HSV.

### `jpeg-monovalue-hsl.sh`
Flatten HSL lightness.

```bash
./jpeg-monovalue-hsl.sh in.jpg
MVT=0.50 MVS=1 ./jpeg-monovalue-hsl.sh in.jpg
```

### `jpeg-monovalue-oklch.sh`
Flatten perceptual lightness.

```bash
./jpeg-monovalue-oklch.sh in.jpg
MVT=0.72 MVC=0.12 ./jpeg-monovalue-oklch.sh in.jpg
```

With `MVC` set, this becomes the perceptual hue-only variant.

### `jpeg-monotone.sh`
This one does **not** use the new mono-value stage.
It uses the existing per-channel expression machinery to make a classic monotone image.

Defaults are warm/sepia-like, but you can set your own endpoints.

```bash
./jpeg-monotone.sh in.jpg
T0R=0.00 T0G=0.02 T0B=0.08 T1R=0.85 T1G=0.92 T1B=1.00 ./jpeg-monotone.sh in.jpg
```

This maps grayscale `x1` to:

- shadows = `(T0R,T0G,T0B)`
- highlights = `(T1R,T1G,T1B)`

## Build note

This patch adds `cmd/jpeg/monovalue.go`.
The `Makefile` is updated so `make jpeg` builds the full `./cmd/jpeg` package.

## Examples

The practical default is MONOVAL=luma when you want the exact dual of your current grayscale math, and MONOVAL=oklch when you want the perceptually cleaner version.

```
JPEG_BIN=./jpeg ./jpeg-monovalue-luma.sh in.jpg
JPEG_BIN=./jpeg MVT=0.45 ./jpeg-monovalue-linear.sh in.jpg
JPEG_BIN=./jpeg MVT=0.75 MVS=1 ./jpeg-monovalue-hsv.sh in.jpg
JPEG_BIN=./jpeg MVT=0.50 MVS=1 ./jpeg-monovalue-hsl.sh in.jpg
JPEG_BIN=./jpeg MVT=0.72 MVC=0.12 ./jpeg-monovalue-oklch.sh in.jpg
JPEG_BIN=./jpeg T0R=0.00 T0G=0.02 T0B=0.08 T1R=0.85 T1G=0.92 T1B=1.00 ./jpeg-monotone.sh in.jpg
```

