# IR3 wrapper reference

This document describes the **full IR3 wrapper** shown below, without the EXIF-copy step.

```bash
#!/usr/bin/env bash
set -euo pipefail
JPEG_BIN="${JPEG_BIN:-/home/lgryglicki/go/bin/jpeg}"
IRRATIO="${IRRATIO:-${IRT:-2.5}}"
IRS="${IRS:-1.0}"
IRL="${IRL:-1.0}"
IRSH="${IRSH:-0.08}"
IRGM="${IRGM:-1.0}"
IRSPLIT="${IRSPLIT:-${IRSSPLIT:-0.55}}"
IRSENDR="${IRSENDR:-0.0}"
IRSENDG="${IRSENDG:-1.0}"
IRSENDB="${IRSENDB:-0.0}"
IRLSPLIT="${IRLSPLIT:-0.685}"
IRLVR="${IRLVR:-0.445}"
IRLVG="${IRLVG:-0.0}"
IRLVB="${IRLVB:-1.0}"
IRLENDR="${IRLENDR:-0.75}"
IRLENDG="${IRLENDG:-0.18}"
IRLENDB="${IRLENDB:-0.87}"
RLO="${RLO:-1.5}"
RHI="${RHI:-1.5}"
GLO="${GLO:-1.5}"
GHI="${GHI:-1.5}"
BLO="${BLO:-1.5}"
BHI="${BHI:-1.5}"
exec env \
  IR3=1 \
  IRRATIO="$IRRATIO" IRT="$IRRATIO" \
  IRS="$IRS" IRL="$IRL" IRSH="$IRSH" IRGM="$IRGM" \
  IRSPLIT="$IRSPLIT" IRSENDR="$IRSENDR" IRSENDG="$IRSENDG" IRSENDB="$IRSENDB" \
  IRLSPLIT="$IRLSPLIT" IRLVR="$IRLVR" IRLVG="$IRLVG" IRLVB="$IRLVB" \
  IRLENDR="$IRLENDR" IRLENDG="$IRLENDG" IRLENDB="$IRLENDB" \
  RR=0 RG=0 RB=1 \
  GR=0 GG=1 GB=0 \
  BR=1 BG=0 BB=0 \
  RLO="$RLO" RHI="$RHI" \
  GLO="$GLO" GHI="$GHI" \
  BLO="$BLO" BHI="$BHI" \
  NA=1 \
  "$JPEG_BIN" "$@"
```

## Processing order

The wrapper performs processing in this order:

1. **Channel routing** using the `RR/RG/RB ... BR/BG/BB` matrix.
2. **Per-channel percentile clipping and stretch** using `RLO/RHI`, `GLO/GHI`, `BLO/BHI`.
3. **IR3 full mapping**.
4. Output is written with the normal `jpeg` output naming.

With the matrix in this wrapper:

```bash
RR=0 RG=0 RB=1
GR=0 GG=1 GB=0
BR=1 BG=0 BB=0
```

the working channels are:

- `R_work = B_source`
- `G_work = G_source`
- `B_work = R_source`

So this wrapper does **RGB → BGR before any clipping or IR mapping**.

## Shell / executable parameters

### `set -euo pipefail`

Shell safety flags:

- `-e`: stop if a command fails
- `-u`: treat unset variables as errors
- `-o pipefail`: fail a pipeline if any stage fails

### `JPEG_BIN`

Path to the `jpeg` binary.

Default:

```bash
/home/lgryglicki/go/bin/jpeg
```

## IR3 mode switch

### `IR3=1`

Enables **full IR3 mapping** inside `jpeg`.

This wrapper is the full-map version. The green-only wrapper uses `IR3GONLY=1` instead.

## Tail detection and overall strength

### `IRRATIO`

Main tail threshold. `IRT` is just a compatibility alias for the same value.

Default:

```bash
2.5
```

IR3 computes:

```text
pos = B / (R + B)
```

using the **working** `R` and `B` after channel routing and after percentile stretch.

Then:

- **short tail** is active when `R / B > IRRATIO`
- **long tail** is active when `B / R > IRRATIO`

Equivalent `pos` thresholds are:

- short tail starts when `pos < 1 / (IRRATIO + 1)`
- long tail starts when `pos > IRRATIO / (IRRATIO + 1)`

With `IRRATIO=2.5`:

- short tail starts below `0.2857`
- long tail starts above `0.7143`

Lower `IRRATIO` makes tails start earlier and affect more pixels.  
Higher `IRRATIO` narrows the tails and makes them more selective.

Because this wrapper swaps `R` and `B` before IR3, in terms of the **original source image**:

- original **blue-heavy** areas become the **short tail**
- original **red-heavy** areas become the **long tail**

### `IRS`

Short-tail strength.

Default:

```bash
1.0
```

Range: `0..1`.

This multiplies how strongly the short-tail ramp is mixed into the output.

- `0`: disable short-tail recoloring
- `1`: full configured short-tail effect

### `IRL`

Long-tail strength.

Default:

```bash
1.0
```

Range: `0..1`.

Same as `IRS`, but for the long tail.

### `IRSH`

Shadow floor.

Default:

```bash
0.08
```

Range: `0..0.99`.

IR3 suppresses tail coloring in dark pixels using:

```text
lum = max(R, B)
amp = smoothstep((lum - IRSH) / (1 - IRSH))
```

Both tail strengths are multiplied by `amp`.

Effect:

- below roughly `IRSH`, tail recoloring is mostly suppressed
- above it, tail recoloring ramps in smoothly

Raise `IRSH` if shadows get false-color noise.  
Lower it if you want more tail coloring in dark regions.

### `IRGM`

Base green multiplier in the middle region.

Default:

```bash
1.0
```

Range: `0..2`.

Before tail blending, full IR3 starts from:

```text
outR = R
outG = G * IRGM
outB = B
```

So `IRGM` scales the working green channel before any tail ramp is applied.

- `< 1.0`: less green in the neutral/middle region
- `1.0`: unchanged green
- `> 1.0`: more green in the neutral/middle region

## Short-tail shape

The short tail is the branch that moves:

```text
red -> yellow -> final short-tail endpoint
```

With the current defaults, the endpoint is pure green, so the short tail becomes:

```text
red -> yellow -> green
```

### `IRSPLIT`

Compatibility alias: `IRSSPLIT`.

Default:

```bash
0.55
```

Range: `0.01..0.99`.

This is the split point inside the short-tail ramp:

- from `u=0` to `u=IRSPLIT`: move from red toward yellow
- from `u=IRSPLIT` to `u=1`: move from yellow toward the configured endpoint

Lower `IRSPLIT`:
- reach yellow earlier
- spend more of the tail moving toward the endpoint

Higher `IRSPLIT`:
- spend longer in the red→yellow stage
- delay movement toward the endpoint

### `IRSENDR`, `IRSENDG`, `IRSENDB`

Final short-tail endpoint color.

Defaults:

```bash
IRSENDR=0.0
IRSENDG=1.0
IRSENDB=0.0
```

So the endpoint is:

```text
(0.0, 1.0, 0.0) = pure green
```

Examples:

- `(0, 1, 0)` = green
- `(0.2, 1, 0)` = yellow-green
- `(0, 0.8, 0.2)` = green-cyan

## Long-tail shape

The long tail is the branch that moves:

```text
blue -> violet anchor -> final long-tail endpoint
```

With the current defaults it ends at a **pinkish-violet**, not pure pink.

### `IRLSPLIT`

Default:

```bash
0.685
```

Range: `0.01..0.99`.

This is the split point inside the long-tail ramp:

- from `u=0` to `u=IRLSPLIT`: move from blue toward the violet anchor
- from `u=IRLSPLIT` to `u=1`: move from the violet anchor toward the final endpoint

Lower `IRLSPLIT`:
- enter violet earlier
- spend more of the tail moving toward the endpoint

Higher `IRLSPLIT`:
- keep the blue→violet region longer
- delay the endpoint phase

### `IRLVR`, `IRLVG`, `IRLVB`

The **violet anchor** color.

Defaults:

```bash
IRLVR=0.445
IRLVG=0.0
IRLVB=1.0
```

So the anchor is:

```text
(0.445, 0.0, 1.0)
```

Increase `IRLVR` to make the violet anchor more magenta/pink.  
Decrease `IRLVR` to keep it bluer.  
`IRLVG` is normally kept near `0` so the violet stays clean.

### `IRLENDR`, `IRLENDG`, `IRLENDB`

Final long-tail endpoint color.

Defaults:

```bash
IRLENDR=0.75
IRLENDG=0.18
IRLENDB=0.87
```

So the endpoint is:

```text
(0.75, 0.18, 0.87)
```

This is a violet-biased pink, not a pure pink.

Effect:

- more `IRLENDR` = pinker / more magenta
- less `IRLENDR` = bluer / more violet
- more `IRLENDG` = warmer / dirtier endpoint
- more `IRLENDB` = cooler / more violet-blue

## Channel routing matrix

These nine variables define how source channels are mixed into the working channels **before** clipping and before IR3:

- `RR`, `RG`, `RB`: build working `R` from source `R/G/B`
- `GR`, `GG`, `GB`: build working `G` from source `R/G/B`
- `BR`, `BG`, `BB`: build working `B` from source `R/G/B`

In this wrapper the matrix is:

```bash
RR=0 RG=0 RB=1
GR=0 GG=1 GB=0
BR=1 BG=0 BB=0
```

So the routing is exactly:

```text
R_work = B_source
G_work = G_source
B_work = R_source
```

That is a pre-map **RGB → BGR swap**.

If you change this matrix, all tail logic and clipping act on the **routed** channels, not on the original source channels.

## Histogram clipping and stretch

### `RLO`, `RHI`

Discard the darkest `RLO%` and brightest `RHI%` of the **working red** channel, then stretch the remainder to full scale.

### `GLO`, `GHI`

Same for the working green channel.

### `BLO`, `BHI`

Same for the working blue channel.

Current defaults in the wrapper example:

```bash
1.5 / 1.5
```

So each working channel discards:

- bottom `1.5%`
- top `1.5%`

before stretching.

Because this wrapper pre-swaps `R` and `B`, that means:

- `RLO/RHI` apply to the **original blue** channel
- `GLO/GHI` apply to the original green channel
- `BLO/BHI` apply to the **original red** channel

Lower clipping percentages preserve more highlight and shadow detail but reduce contrast.  
Higher clipping percentages increase contrast but clip more.

## Alpha handling

### `NA=1`

Skip alpha processing and force alpha to full opacity.

For normal JPEG-style workflows this is usually what you want.

## Current default behavior of this wrapper

With the wrapper exactly as shown:

1. Source channels are first swapped `RGB -> BGR`.
2. Each working channel is clipped at `1.5% / 1.5%`.
3. IR3 runs with threshold `2.5`.
4. Short tail ends at pure green.
5. Long tail passes through violet and ends at pinkish-violet.

So, in terms of the **original unswapped source image**:

- original **blue-heavy** areas are the ones that go toward **red -> yellow -> green**
- original **red-heavy** areas are the ones that go toward **blue -> violet -> pink**

That inversion happens entirely because of the `RR/RG/RB ... BR/BG/BB` swap matrix.

## Green-only wrapper differences

The green-only wrapper enables `IR3GONLY=1` instead of `IR3=1`.

In green-only mode:

- output `R` is left unchanged
- output `B` is left unchanged
- only output `G` is modified

It uses these extra parameters instead of the full long-tail endpoint colors:

- `IRGSHORTEND`: gain for the green bump at the short-tail extreme
- `IRGLONGMID`: green level reached by the long-tail mid stage
- `IRGLONGEND`: final green level at the long-tail extreme
- `IRGLONGSPLIT`: split point inside the long-tail green ramp
