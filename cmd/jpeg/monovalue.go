package main

import (
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
)

type monoValueConfig struct {
	enabled           bool
	mode              string
	target            float64
	lumaR             float64
	lumaG             float64
	lumaB             float64
	gamutMode         string
	zeroMode          string
	satOverride       float64
	satOverrideSet    bool
	chromaOverride    float64
	chromaOverrideSet bool
}

func monoValueConfigFromEnv() (monoValueConfig, error) {
	cfg := monoValueConfig{
		enabled:   false,
		mode:      "",
		target:    0.5,
		lumaR:     0.2126,
		lumaG:     0.7152,
		lumaB:     0.0722,
		gamutMode: "fit",
		zeroMode:  "gray",
	}
	mode := strings.TrimSpace(strings.ToLower(os.Getenv("MONOVAL")))
	if mode == "" {
		mode = strings.TrimSpace(strings.ToLower(os.Getenv("MVMODE")))
	}
	if mode == "" {
		return cfg, nil
	}
	switch mode {
	case "lum", "luma", "gray", "mono", "monovalue":
		cfg.mode = "luma"
	case "linear", "lumalin", "lin":
		cfg.mode = "linear"
	case "hsv":
		cfg.mode = "hsv"
	case "hsl":
		cfg.mode = "hsl"
	case "oklab", "oklch":
		cfg.mode = "oklch"
	default:
		return cfg, fmt.Errorf("MONOVAL/MVMODE must be one of: luma, linear, hsv, hsl, oklch")
	}
	cfg.enabled = true

	parse := func(env string, dst *float64, lo, hi float64) error {
		s := os.Getenv(env)
		if s == "" {
			return nil
		}
		v, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return err
		}
		if v < lo || v > hi {
			return fmt.Errorf("%s must be from %f-%f range", env, lo, hi)
		}
		*dst = v
		return nil
	}
	if err := parse("MVT", &cfg.target, 0.0, 1.0); err != nil {
		return cfg, err
	}
	if err := parse("MVR", &cfg.lumaR, 0.0, 1.0); err != nil {
		return cfg, err
	}
	if err := parse("MVG", &cfg.lumaG, 0.0, 1.0); err != nil {
		return cfg, err
	}
	if err := parse("MVB", &cfg.lumaB, 0.0, 1.0); err != nil {
		return cfg, err
	}
	if err := parse("MVS", &cfg.satOverride, 0.0, 1.0); err != nil {
		return cfg, err
	}
	if os.Getenv("MVS") != "" {
		cfg.satOverrideSet = true
	}
	if err := parse("MVC", &cfg.chromaOverride, 0.0, 1.0); err != nil {
		return cfg, err
	}
	if os.Getenv("MVC") != "" {
		cfg.chromaOverrideSet = true
	}

	tot := cfg.lumaR + cfg.lumaG + cfg.lumaB
	if tot <= 0.0 {
		return cfg, fmt.Errorf("MVR+MVG+MVB must be positive")
	}
	cfg.lumaR /= tot
	cfg.lumaG /= tot
	cfg.lumaB /= tot

	gamutMode := strings.TrimSpace(strings.ToLower(os.Getenv("MVGAMUT")))
	if gamutMode != "" {
		switch gamutMode {
		case "fit", "clip":
			cfg.gamutMode = gamutMode
		default:
			return cfg, fmt.Errorf("MVGAMUT must be 'fit' or 'clip'")
		}
	}
	zeroMode := strings.TrimSpace(strings.ToLower(os.Getenv("MVZERO")))
	if zeroMode != "" {
		switch zeroMode {
		case "gray", "black":
			cfg.zeroMode = zeroMode
		default:
			return cfg, fmt.Errorf("MVZERO must be 'gray' or 'black'")
		}
	}
	return cfg, nil
}

func srgbToLinear(v float64) float64 {
	v = clamp01(v)
	if v <= 0.04045 {
		return v / 12.92
	}
	return math.Pow((v+0.055)/1.055, 2.4)
}

func linearToSRGB(v float64) float64 {
	v = clamp01(v)
	if v <= 0.0031308 {
		return 12.92 * v
	}
	return 1.055*math.Pow(v, 1.0/2.4) - 0.055
}

func inGamut01(r, g, b float64) bool {
	return r >= 0.0 && r <= 1.0 && g >= 0.0 && g <= 1.0 && b >= 0.0 && b <= 1.0
}

func fitRGBByGrayMix(r, g, b, gray float64) (float64, float64, float64) {
	if inGamut01(r, g, b) {
		return r, g, b
	}
	lo := 0.0
	hi := 1.0
	for i := 0; i < 32; i++ {
		mid := 0.5 * (lo + hi)
		mr := gray + mid*(r-gray)
		mg := gray + mid*(g-gray)
		mb := gray + mid*(b-gray)
		if inGamut01(mr, mg, mb) {
			lo = mid
		} else {
			hi = mid
		}
	}
	mr := gray + lo*(r-gray)
	mg := gray + lo*(g-gray)
	mb := gray + lo*(b-gray)
	return clamp01(mr), clamp01(mg), clamp01(mb)
}

func monoValueLuma(r, g, b float64, cfg monoValueConfig) (float64, float64, float64) {
	const eps = 1e-12
	y := cfg.lumaR*r + cfg.lumaG*g + cfg.lumaB*b
	if y <= eps {
		if cfg.zeroMode == "black" {
			return 0.0, 0.0, 0.0
		}
		return cfg.target, cfg.target, cfg.target
	}
	s := cfg.target / y
	cr := r * s
	cg := g * s
	cb := b * s
	if cfg.gamutMode == "fit" {
		return fitRGBByGrayMix(cr, cg, cb, cfg.target)
	}
	return clamp01(cr), clamp01(cg), clamp01(cb)
}

func monoValueLinear(r, g, b float64, cfg monoValueConfig) (float64, float64, float64) {
	const eps = 1e-12
	lr := srgbToLinear(r)
	lg := srgbToLinear(g)
	lb := srgbToLinear(b)
	y := cfg.lumaR*lr + cfg.lumaG*lg + cfg.lumaB*lb
	if y <= eps {
		if cfg.zeroMode == "black" {
			return 0.0, 0.0, 0.0
		}
		gray := linearToSRGB(cfg.target)
		return gray, gray, gray
	}
	s := cfg.target / y
	clr := lr * s
	clg := lg * s
	clb := lb * s
	if cfg.gamutMode == "fit" {
		clr, clg, clb = fitRGBByGrayMix(clr, clg, clb, cfg.target)
	} else {
		clr = clamp01(clr)
		clg = clamp01(clg)
		clb = clamp01(clb)
	}
	return linearToSRGB(clr), linearToSRGB(clg), linearToSRGB(clb)
}

func rgbToHSV(r, g, b float64) (float64, float64, float64) {
	maxc := math.Max(r, math.Max(g, b))
	minc := math.Min(r, math.Min(g, b))
	delta := maxc - minc
	h := 0.0
	s := 0.0
	v := maxc
	if maxc > 0.0 {
		s = delta / maxc
	}
	if delta > 0.0 {
		switch maxc {
		case r:
			h = math.Mod((g-b)/delta, 6.0)
		case g:
			h = ((b-r)/delta + 2.0)
		default:
			h = ((r-g)/delta + 4.0)
		}
		h /= 6.0
		if h < 0.0 {
			h += 1.0
		}
	}
	return h, s, v
}

func hsvToRGB(h, s, v float64) (float64, float64, float64) {
	h = h - math.Floor(h)
	s = clamp01(s)
	v = clamp01(v)
	if s <= 0.0 {
		return v, v, v
	}
	h6 := h * 6.0
	i := int(math.Floor(h6))
	f := h6 - float64(i)
	p := v * (1.0 - s)
	q := v * (1.0 - s*f)
	t := v * (1.0 - s*(1.0-f))
	switch i % 6 {
	case 0:
		return v, t, p
	case 1:
		return q, v, p
	case 2:
		return p, v, t
	case 3:
		return p, q, v
	case 4:
		return t, p, v
	default:
		return v, p, q
	}
}

func monoValueHSV(r, g, b float64, cfg monoValueConfig) (float64, float64, float64) {
	h, s, _ := rgbToHSV(r, g, b)
	if cfg.satOverrideSet {
		s = cfg.satOverride
	}
	return hsvToRGB(h, s, cfg.target)
}

func rgbToHSL(r, g, b float64) (float64, float64, float64) {
	maxc := math.Max(r, math.Max(g, b))
	minc := math.Min(r, math.Min(g, b))
	delta := maxc - minc
	l := 0.5 * (maxc + minc)
	h := 0.0
	s := 0.0
	if delta > 0.0 {
		if l < 0.5 {
			s = delta / (maxc + minc)
		} else {
			s = delta / (2.0 - maxc - minc)
		}
		switch maxc {
		case r:
			h = math.Mod((g-b)/delta, 6.0)
		case g:
			h = ((b-r)/delta + 2.0)
		default:
			h = ((r-g)/delta + 4.0)
		}
		h /= 6.0
		if h < 0.0 {
			h += 1.0
		}
	}
	return h, s, l
}

func hueToRGB(p, q, t float64) float64 {
	if t < 0.0 {
		t += 1.0
	}
	if t > 1.0 {
		t -= 1.0
	}
	if t < 1.0/6.0 {
		return p + (q-p)*6.0*t
	}
	if t < 1.0/2.0 {
		return q
	}
	if t < 2.0/3.0 {
		return p + (q-p)*(2.0/3.0-t)*6.0
	}
	return p
}

func hslToRGB(h, s, l float64) (float64, float64, float64) {
	h = h - math.Floor(h)
	s = clamp01(s)
	l = clamp01(l)
	if s <= 0.0 {
		return l, l, l
	}
	q := 0.0
	if l < 0.5 {
		q = l * (1.0 + s)
	} else {
		q = l + s - l*s
	}
	p := 2.0*l - q
	return hueToRGB(p, q, h+1.0/3.0), hueToRGB(p, q, h), hueToRGB(p, q, h-1.0/3.0)
}

func monoValueHSL(r, g, b float64, cfg monoValueConfig) (float64, float64, float64) {
	h, s, _ := rgbToHSL(r, g, b)
	if cfg.satOverrideSet {
		s = cfg.satOverride
	}
	return hslToRGB(h, s, cfg.target)
}

func linearRGBToOklab(r, g, b float64) (float64, float64, float64) {
	l := 0.4122214708*r + 0.5363325363*g + 0.0514459929*b
	m := 0.2119034982*r + 0.6806995451*g + 0.1073969566*b
	s := 0.0883024619*r + 0.2817188376*g + 0.6299787005*b
	l3 := math.Cbrt(math.Max(l, 0.0))
	m3 := math.Cbrt(math.Max(m, 0.0))
	s3 := math.Cbrt(math.Max(s, 0.0))
	L := 0.2104542553*l3 + 0.7936177850*m3 - 0.0040720468*s3
	a := 1.9779984951*l3 - 2.4285922050*m3 + 0.4505937099*s3
	bb := 0.0259040371*l3 + 0.7827717662*m3 - 0.8086757660*s3
	return L, a, bb
}

func oklabToLinearRGB(L, a, b float64) (float64, float64, float64) {
	l3 := L + 0.3963377774*a + 0.2158037573*b
	m3 := L - 0.1055613458*a - 0.0638541728*b
	s3 := L - 0.0894841775*a - 1.2914855480*b
	l := l3 * l3 * l3
	m := m3 * m3 * m3
	s := s3 * s3 * s3
	r := 4.0767416621*l - 3.3077115913*m + 0.2309699292*s
	g := -1.2684380046*l + 2.6097574011*m - 0.3413193965*s
	bb := -0.0041960863*l - 0.7034186147*m + 1.7076147010*s
	return r, g, bb
}

func monoValueOKLCh(r, g, b float64, cfg monoValueConfig) (float64, float64, float64) {
	lr := srgbToLinear(r)
	lg := srgbToLinear(g)
	lb := srgbToLinear(b)
	_, a, bb := linearRGBToOklab(lr, lg, lb)
	C := math.Hypot(a, bb)
	H := 0.0
	if C > 1e-12 {
		H = math.Atan2(bb, a)
	}
	if cfg.chromaOverrideSet {
		C = cfg.chromaOverride
	}
	newA := C * math.Cos(H)
	newB := C * math.Sin(H)
	or, og, ob := oklabToLinearRGB(cfg.target, newA, newB)
	if cfg.gamutMode == "fit" && !inGamut01(or, og, ob) {
		lo := 0.0
		hi := C
		for i := 0; i < 32; i++ {
			mid := 0.5 * (lo + hi)
			ma := mid * math.Cos(H)
			mb := mid * math.Sin(H)
			tr, tg, tb := oklabToLinearRGB(cfg.target, ma, mb)
			if inGamut01(tr, tg, tb) {
				lo = mid
			} else {
				hi = mid
			}
		}
		newA = lo * math.Cos(H)
		newB = lo * math.Sin(H)
		or, og, ob = oklabToLinearRGB(cfg.target, newA, newB)
	}
	or = clamp01(or)
	og = clamp01(og)
	ob = clamp01(ob)
	return linearToSRGB(or), linearToSRGB(og), linearToSRGB(ob)
}

func applyMonoValue(pxdata [][][4]uint16, x, y, thrN int, cfg monoValueConfig) error {
	if !cfg.enabled {
		return nil
	}
	che := make(chan error)
	nThreads := 0
	for ii := 0; ii < x; ii++ {
		go func(c chan error, i int) {
			for j := 0; j < y; j++ {
				px := pxdata[i][j]
				r := float64(px[0]) / 65535.0
				g := float64(px[1]) / 65535.0
				b := float64(px[2]) / 65535.0
				var rr, gg, bb float64
				switch cfg.mode {
				case "luma":
					rr, gg, bb = monoValueLuma(r, g, b, cfg)
				case "linear":
					rr, gg, bb = monoValueLinear(r, g, b, cfg)
				case "hsv":
					rr, gg, bb = monoValueHSV(r, g, b, cfg)
				case "hsl":
					rr, gg, bb = monoValueHSL(r, g, b, cfg)
				case "oklch":
					rr, gg, bb = monoValueOKLCh(r, g, b, cfg)
				default:
					rr, gg, bb = r, g, b
				}
				pxdata[i][j][0] = uint16(clamp01(rr)*65535.0 + 0.5)
				pxdata[i][j][1] = uint16(clamp01(gg)*65535.0 + 0.5)
				pxdata[i][j][2] = uint16(clamp01(bb)*65535.0 + 0.5)
			}
			c <- nil
		}(che, ii)
		nThreads++
		if nThreads == thrN {
			e := <-che
			if e != nil {
				return e
			}
			nThreads--
		}
	}
	for nThreads > 0 {
		e := <-che
		if e != nil {
			return e
		}
		nThreads--
	}
	return nil
}
