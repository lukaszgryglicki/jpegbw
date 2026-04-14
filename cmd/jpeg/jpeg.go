package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io/ioutil"
	"math"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/lukaszgryglicki/jpegbw"
)

type ir3Config struct {
	enabled        bool
	greenOnly      bool
	threshold      float64
	shortStrength  float64
	longStrength   float64
	shadowFloor    float64
	greenMid       float64
	shortSplit     float64
	shortEndR      float64
	shortEndG      float64
	shortEndB      float64
	longSplit      float64
	longVioletR    float64
	longVioletG    float64
	longVioletB    float64
	longEndR       float64
	longEndG       float64
	longEndB       float64
	gOnlyShortEnd  float64
	gOnlyLongMid   float64
	gOnlyLongEnd   float64
	gOnlyLongSplit float64
}

func clamp01(v float64) float64 {
	if v <= 0.0 {
		return 0.0
	}
	if v >= 1.0 {
		return 1.0
	}
	return v
}

func smooth01(v float64) float64 {
	v = clamp01(v)
	return v * v * (3.0 - 2.0*v)
}

func mixf(a, b, t float64) float64 {
	return a + (b-a)*t
}

func ir3ConfigFromEnv() (ir3Config, error) {
	cfg := ir3Config{
		enabled:       false,
		greenOnly:     false,
		threshold:     2.5,
		shortStrength: 1.0,
		longStrength:  1.0,
		shadowFloor:   0.08,
		greenMid:      1.0,

		/* shortwave full-map defaults: unchanged from v3 */
		shortSplit: 0.55,
		shortEndR:  0.0,
		shortEndG:  1.0,
		shortEndB:  0.0,

		/*
			Longwave full-map defaults:
			- violet should start earlier than in v3
			- endpoint should be halfway between v2 and v3
			v2: split=0.55 violet=(0.55,0.00,1.00) end=(1.00,0.30,0.82)
			v3: split=0.82 violet=(0.34,0.00,1.00) end=(0.50,0.06,0.92)
			v4 midpoint default:
			    split=0.685 violet=(0.445,0.00,1.00) end=(0.75,0.18,0.87)
		*/
		longSplit:   0.685,
		longVioletR: 0.445,
		longVioletG: 0.0,
		longVioletB: 1.0,
		longEndR:    0.75,
		longEndG:    0.18,
		longEndB:    0.87,

		/* green-only defaults: unchanged from v3 */
		gOnlyShortEnd:  1.0,
		gOnlyLongMid:   0.60,
		gOnlyLongEnd:   1.0,
		gOnlyLongSplit: 0.80,
	}
	if os.Getenv("IR3") != "" {
		cfg.enabled = true
	}
	if os.Getenv("IR3GONLY") != "" {
		cfg.enabled = true
		cfg.greenOnly = true
	}
	parseAny := func(envs []string, dst *float64, lo, hi float64) error {
		for _, env := range envs {
			s := os.Getenv(env)
			if s == "" {
				continue
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
		return nil
	}
	if err := parseAny([]string{"IRRATIO", "IRT"}, &cfg.threshold, 1.01, 65535.0); err != nil {
		return cfg, err
	}
	if err := parseAny([]string{"IRS"}, &cfg.shortStrength, 0.0, 1.0); err != nil {
		return cfg, err
	}
	if err := parseAny([]string{"IRL"}, &cfg.longStrength, 0.0, 1.0); err != nil {
		return cfg, err
	}
	if err := parseAny([]string{"IRSH"}, &cfg.shadowFloor, 0.0, 0.99); err != nil {
		return cfg, err
	}
	if err := parseAny([]string{"IRGM"}, &cfg.greenMid, 0.0, 2.0); err != nil {
		return cfg, err
	}

	if err := parseAny([]string{"IRSPLIT", "IRSSPLIT"}, &cfg.shortSplit, 0.01, 0.99); err != nil {
		return cfg, err
	}
	if err := parseAny([]string{"IRSENDR", "IRSHORTENDR"}, &cfg.shortEndR, 0.0, 1.0); err != nil {
		return cfg, err
	}
	if err := parseAny([]string{"IRSENDG", "IRSHORTENDG"}, &cfg.shortEndG, 0.0, 1.0); err != nil {
		return cfg, err
	}
	if err := parseAny([]string{"IRSENDB", "IRSHORTENDB"}, &cfg.shortEndB, 0.0, 1.0); err != nil {
		return cfg, err
	}

	if err := parseAny([]string{"IRLSPLIT"}, &cfg.longSplit, 0.01, 0.99); err != nil {
		return cfg, err
	}
	if err := parseAny([]string{"IRLVR", "IRLONGVIOLETR"}, &cfg.longVioletR, 0.0, 1.0); err != nil {
		return cfg, err
	}
	if err := parseAny([]string{"IRLVG", "IRLONGVIOLETG"}, &cfg.longVioletG, 0.0, 1.0); err != nil {
		return cfg, err
	}
	if err := parseAny([]string{"IRLVB", "IRLONGVIOLETB"}, &cfg.longVioletB, 0.0, 1.0); err != nil {
		return cfg, err
	}
	if err := parseAny([]string{"IRLENDR", "IRLONGENDR"}, &cfg.longEndR, 0.0, 1.0); err != nil {
		return cfg, err
	}
	if err := parseAny([]string{"IRLENDG", "IRLONGENDG"}, &cfg.longEndG, 0.0, 1.0); err != nil {
		return cfg, err
	}
	if err := parseAny([]string{"IRLENDB", "IRLONGENDB"}, &cfg.longEndB, 0.0, 1.0); err != nil {
		return cfg, err
	}

	if err := parseAny([]string{"IRGSHORTEND"}, &cfg.gOnlyShortEnd, 0.0, 2.0); err != nil {
		return cfg, err
	}
	if err := parseAny([]string{"IRGLONGMID"}, &cfg.gOnlyLongMid, 0.0, 2.0); err != nil {
		return cfg, err
	}
	if err := parseAny([]string{"IRGLONGEND"}, &cfg.gOnlyLongEnd, 0.0, 2.0); err != nil {
		return cfg, err
	}
	if err := parseAny([]string{"IRGLONGSPLIT"}, &cfg.gOnlyLongSplit, 0.01, 0.99); err != nil {
		return cfg, err
	}
	return cfg, nil
}

func ir3ShortRamp(u float64, cfg ir3Config) (float64, float64, float64) {
	u = clamp01(u)
	t1 := smooth01(u / cfg.shortSplit)
	t2 := smooth01((u - cfg.shortSplit) / (1.0 - cfg.shortSplit))
	r := 1.0
	g := t1
	b := 0.0
	r = mixf(r, cfg.shortEndR, t2)
	g = mixf(g, cfg.shortEndG, t2)
	b = mixf(b, cfg.shortEndB, t2)
	return clamp01(r), clamp01(g), clamp01(b)
}

func ir3LongRamp(u float64, cfg ir3Config) (float64, float64, float64) {
	u = clamp01(u)
	t1 := smooth01(u / cfg.longSplit)
	t2 := smooth01((u - cfg.longSplit) / (1.0 - cfg.longSplit))
	r := mixf(0.0, cfg.longVioletR, t1)
	g := mixf(0.0, cfg.longVioletG, t1)
	b := mixf(1.0, cfg.longVioletB, t1)
	r = mixf(r, cfg.longEndR, t2)
	g = mixf(g, cfg.longEndG, t2)
	b = mixf(b, cfg.longEndB, t2)
	return clamp01(r), clamp01(g), clamp01(b)
}

func ir3LongRampGOnly(u float64, cfg ir3Config) float64 {
	u = clamp01(u)
	t1 := smooth01(u / cfg.gOnlyLongSplit)
	t2 := smooth01((u - cfg.gOnlyLongSplit) / (1.0 - cfg.gOnlyLongSplit))
	g := mixf(0.0, cfg.gOnlyLongMid, t1)
	g = mixf(g, cfg.gOnlyLongEnd, t2)
	return clamp01(g)
}

func ir3Map(r, g, b float64, cfg ir3Config) (float64, float64, float64) {
	const eps = 1e-12

	r = clamp01(r)
	g = clamp01(g)
	b = clamp01(b)

	pos := b / (r + b + eps)
	shortThr := 1.0 / (cfg.threshold + 1.0)
	longThr := cfg.threshold / (cfg.threshold + 1.0)

	shortTail := 0.0
	longTail := 0.0
	if pos < shortThr {
		shortTail = smooth01((shortThr - pos) / (shortThr + eps))
	}
	if pos > longThr {
		longTail = smooth01((pos - longThr) / (1.0 - longThr + eps))
	}

	lum := math.Max(r, b)
	amp := smooth01((lum - cfg.shadowFloor) / (1.0 - cfg.shadowFloor))
	shortTail *= amp
	longTail *= amp

	outR := r
	outG := clamp01(g * cfg.greenMid)
	outB := b

	if shortTail > 0.0 {
		sr, sg, sb := ir3ShortRamp(shortTail, cfg)
		t := clamp01(cfg.shortStrength * shortTail)
		if cfg.greenOnly {
			outG = mixf(outG, lum*clamp01(sg*cfg.gOnlyShortEnd), t)
		} else {
			outR = mixf(outR, lum*sr, t)
			outG = mixf(outG, lum*sg, t)
			outB = mixf(outB, lum*sb, t)
		}
	}
	if longTail > 0.0 {
		t := clamp01(cfg.longStrength * longTail)
		if cfg.greenOnly {
			lg := ir3LongRampGOnly(longTail, cfg)
			outG = mixf(outG, lum*lg, t)
		} else {
			lr, lg, lb := ir3LongRamp(longTail, cfg)
			outR = mixf(outR, lum*lr, t)
			outG = mixf(outG, lum*lg, t)
			outB = mixf(outB, lum*lb, t)
		}
	}

	if cfg.greenOnly {
		return r, clamp01(outG), b
	}
	return clamp01(outR), clamp01(outG), clamp01(outB)
}

func applyIR3(pxdata [][][4]uint16, x, y, thrN int, cfg ir3Config) error {
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
				rr, gg, bb := ir3Map(r, g, b, cfg)
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

type isoValConfig struct {
	enabled bool
	mode    string
	target  float64
	wR      float64
	wG      float64
	wB      float64
}

func isoValConfigFromEnv() (isoValConfig, error) {
	cfg := isoValConfig{
		enabled: false,
		mode:    "",
		target:  0.5,
		wR:      0.2126,
		wG:      0.7152,
		wB:      0.0722,
	}
	mode := strings.TrimSpace(strings.ToLower(os.Getenv("ISOVAL")))
	if mode == "" {
		return cfg, nil
	}
	switch mode {
	case "add", "offset", "shift":
		cfg.mode = "add"
	case "mul", "mult", "multiply", "scale":
		cfg.mode = "mul"
	default:
		return cfg, fmt.Errorf("ISOVAL must be one of: add, mul")
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
	if err := parse("IVT", &cfg.target, 0.0, 1.0); err != nil {
		return cfg, err
	}
	if err := parse("IVR", &cfg.wR, 0.0, 1.0); err != nil {
		return cfg, err
	}
	if err := parse("IVG", &cfg.wG, 0.0, 1.0); err != nil {
		return cfg, err
	}
	if err := parse("IVB", &cfg.wB, 0.0, 1.0); err != nil {
		return cfg, err
	}
	fact := cfg.wR + cfg.wG + cfg.wB
	if fact <= 0.0 {
		return cfg, fmt.Errorf("IVR+IVG+IVB must be positive")
	}
	cfg.wR /= fact
	cfg.wG /= fact
	cfg.wB /= fact
	return cfg, nil
}

func isoValValue(r, g, b float64, cfg isoValConfig) float64 {
	return cfg.wR*r + cfg.wG*g + cfg.wB*b
}

func isoValAddMode(r, g, b float64, cfg isoValConfig) (float64, float64, float64) {
	delta := cfg.target - isoValValue(r, g, b, cfg)
	return clamp01(r + delta), clamp01(g + delta), clamp01(b + delta)
}

func isoValMulMode(r, g, b float64, cfg isoValConfig) (float64, float64, float64) {
	const eps = 1e-12
	v := isoValValue(r, g, b, cfg)
	if v <= eps {
		return r, g, b
	}
	s := cfg.target / v
	return clamp01(r * s), clamp01(g * s), clamp01(b * s)
}

func applyIsoVal(pxdata [][][4]uint16, x, y, thrN int, cfg isoValConfig) error {
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
				case "add":
					rr, gg, bb = isoValAddMode(r, g, b, cfg)
				case "mul":
					rr, gg, bb = isoValMulMode(r, g, b, cfg)
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

// images2RGBA: convert given images to bw: iname.ext -> co_iname.ext, dir/iname.ext -> dir/co_iname.ext
// Other parameters are set via env variables (see main() function it describes all env params):
func images2RGBA(args []string) error {
	// JPEG Quality
	jpegqStr := os.Getenv("Q")
	jpegq := -1
	if jpegqStr != "" {
		v, err := strconv.Atoi(jpegqStr)
		if err != nil {
			return err
		}
		if v < 1 || v > 100 {
			return fmt.Errorf("Q must be from 1-100 range")
		}
		jpegq = v
	}

	// PNG Quality
	pngqStr := os.Getenv("PQ")
	pngq := png.DefaultCompression
	if pngqStr != "" {
		v, err := strconv.Atoi(pngqStr)
		if err != nil {
			return err
		}
		if v < 0 || v > 3 {
			return fmt.Errorf("PQ must be from 0-3 range")
		}
		pngq = png.CompressionLevel(-v)
	}

	// Threads
	thrsS := os.Getenv("N")
	thrs := -1
	if thrsS != "" {
		t, err := strconv.Atoi(thrsS)
		if err != nil {
			return err
		}
		thrs = t
	}
	thrN := thrs
	if thrs < 0 {
		thrN = runtime.NumCPU()
	}
	runtime.GOMAXPROCS(thrN)

	// Override file name config
	overS := os.Getenv("O")
	overB := false
	overFrom := ""
	overTo := ""
	if overS != "" {
		ary := strings.Split(overS, ":")
		if len(ary) != 2 {
			return fmt.Errorf("bad override filename config: %s", overS)
		}
		overFrom = ary[0]
		overTo = ary[1]
		overB = true
	}

	ir3cfg, err := ir3ConfigFromEnv()
	if err != nil {
		return err
	}
	if ir3cfg.enabled {
		fmt.Printf("IR3 enabled: threshold=%f short=%f long=%f shadow=%f gmid=%f green_only=%v\n", ir3cfg.threshold, ir3cfg.shortStrength, ir3cfg.longStrength, ir3cfg.shadowFloor, ir3cfg.greenMid, ir3cfg.greenOnly)
	}

	isovalcfg, err := isoValConfigFromEnv()
	if err != nil {
		return err
	}
	if isovalcfg.enabled {
		fmt.Printf("ISOVAL enabled: mode=%s target=%f weights=(%f,%f,%f)\n", isovalcfg.mode, isovalcfg.target, isovalcfg.wR, isovalcfg.wG, isovalcfg.wB)
	}

	monocfg, err := monoValueConfigFromEnv()
	if err != nil {
		return err
	}
	if monocfg.enabled {
		fmt.Printf("MONOVAL enabled: mode=%s target=%f gamut=%s zero=%s weights=(%f,%f,%f)", monocfg.mode, monocfg.target, monocfg.gamutMode, monocfg.zeroMode, monocfg.lumaR, monocfg.lumaG, monocfg.lumaB)
		if monocfg.satOverrideSet {
			fmt.Printf(" sat=%f", monocfg.satOverride)
		}
		if monocfg.chromaOverrideSet {
			fmt.Printf(" chroma=%f", monocfg.chromaOverride)
		}
		fmt.Printf("\n")
	}

	// RGBA arrays
	rgba := [4]string{"R", "G", "B", "A"}
	var (
		fctx    [4]jpegbw.FparCtx
		bFun    [4]bool
		useImag [4]bool
		ar      [4]float64
		ag      [4]float64
		ab      [4]float64
		alo     [4]float64
		aloi    [4]uint16
		ahi     [4]float64
		ahii    [4]uint16
		aga     [4]float64
		agaB    [4]bool
		acl     [4]int
		acont   [4]uint16
		aedge   [4]uint16
		asurf   [4]uint16
		agcont  [4]bool
	)

	// Global contour setting
	bGlobCont := false
	conti := uint16(0)
	contS := os.Getenv("CONT")
	if contS != "" {
		v, err := strconv.Atoi(contS)
		if err != nil {
			return err
		}
		if v < 1 || v > 0x3fff {
			return fmt.Errorf("CONT must be from 0001-3FFF range")
		}
		conti = uint16(v)
	}
	if conti > 0 {
		bGlobCont = true
		acont = [4]uint16{conti, conti, conti, conti}
	}

	// Edge and surface modes
	// 0 - 0
	// 1 - 1
	// 2 - original value
	// 3 - inverted
	// Defaults
	// EDGE=1 SURF=0
	bGlobEdge := false
	edge := uint16(1)
	edgeS := os.Getenv("EDGE")
	if edgeS != "" {
		v, err := strconv.Atoi(edgeS)
		if err != nil {
			return err
		}
		if v < 0 || v > 3 {
			return fmt.Errorf("EDGE must be from [0, 1, 2, 3]")
		}
		edge = uint16(v)
		bGlobEdge = true
		aedge = [4]uint16{edge, edge, edge, edge}
	}
	bGlobSurf := false
	surf := uint16(0)
	surfS := os.Getenv("SURF")
	if surfS != "" {
		v, err := strconv.Atoi(surfS)
		if err != nil {
			return err
		}
		if v < 0 || v > 3 {
			return fmt.Errorf("SURF must be from [0, 1, 2, 3]")
		}
		surf = uint16(v)
		bGlobSurf = true
		asurf = [4]uint16{surf, surf, surf, surf}
	}
	bGlobGCont := false
	gcont := false
	gcontS := os.Getenv("GCONT")
	if gcontS != "" {
		v, err := strconv.Atoi(gcontS)
		if err != nil {
			return err
		}
		if v < 0 || v > 1 {
			return fmt.Errorf("GCONT must be from [0, 1]")
		}
		if v == 1 {
			gcont = true
		}
		bGlobGCont = true
		agcont = [4]bool{gcont, gcont, gcont, gcont}
	}

	// Main library context
	var mfctx jpegbw.FparCtx
	lib := os.Getenv("LIB")
	if lib != "" {
		nf := 128
		nfs := os.Getenv("NF")
		if nfs != "" {
			v, err := strconv.Atoi(nfs)
			if err != nil {
				return err
			}
			if v < 1 || v > 0xffff {
				return fmt.Errorf("NF must be from 1-65535 range")
			}
			nf = v
		}
		ok := mfctx.Init(lib, uint(nf))
		if !ok {
			return fmt.Errorf("LIB init failed for: %s", lib)
		}
		defer func() { mfctx.Tidy() }()
	}

	// Additional in-image info: R,G,B,G scale to the right and RGB histogram on the bottom
	einf := os.Getenv("EINF") != ""
	infS := os.Getenv("INF")
	inf := 0
	shpow := 1.0
	bhpow := false
	if infS != "" {
		in, err := strconv.Atoi(infS)
		if err != nil {
			return err
		}
		inf = in
		shpowS := os.Getenv("HPOW")
		if shpowS != "" {
			v, err := strconv.ParseFloat(shpowS, 64)
			if err != nil {
				return err
			}
			if v < 0.05 || v > 20.0 {
				return fmt.Errorf("HPOW must be from 0.05-20 range")
			}
			shpow = v
			bhpow = true
		}
	}

	// Hint mode
	useHints := os.Getenv("HINT") != ""
	hintRequired := os.Getenv("HINTREQ") != ""

	// No alpha processing
	noA := os.Getenv("NA") != ""

	// Reverse the calculaTION
	rev := os.Getenv("REV") != ""

	// All colors mult (will keep smallest low index, biggest high index and calculate mult accoring to that)
	acm := os.Getenv("ACM") != ""
	acmFact := 0.0
	if acm {
		acmS := os.Getenv("ACM")
		v, err := strconv.ParseFloat(acmS, 64)
		if err != nil {
			return err
		}
		if v < 0.0 || v > 1.0 {
			return fmt.Errorf("ACM must be from 0-1 range")
		}
		acmFact = v
	}

	// Grayscale output
	ogs := os.Getenv("OGS") != ""
	gsr := 1.0
	gsg := 1.0
	gsb := 1.0
	if ogs {
		// R red
		rS := os.Getenv("GSR")
		r := 1.0
		if rS != "" {
			v, err := strconv.ParseFloat(rS, 64)
			if err != nil {
				return err
			}
			r = v
		}

		// G green
		gS := os.Getenv("GSG")
		g := 1.0
		if gS != "" {
			v, err := strconv.ParseFloat(gS, 64)
			if err != nil {
				return err
			}
			g = v
		}

		// B blue
		bS := os.Getenv("GSB")
		b := 1.0
		if bS != "" {
			v, err := strconv.ParseFloat(bS, 64)
			if err != nil {
				return err
			}
			b = v
		}
		fact := r + g + b
		if fact <= 0 {
			r = 0.0
			g = 0.0
			b = 0.0
		} else {
			r /= fact
			g /= fact
			b /= fact
		}
		gsr = r
		gsg = g
		gsb = b
		fmt.Printf("Enabling GS output: %f,%f,%f\n", gsr, gsg, gsb)
	}

	// Process colors
	for colidx, colrgba := range rgba {
		if noA && colidx == 3 {
			continue
		}

		// Per color cache level 0-4
		clS := os.Getenv(colrgba + "C")
		cl := 0
		if clS != "" {
			v, err := strconv.Atoi(clS)
			if err != nil {
				return err
			}
			if v < 0 || v > 4 {
				return fmt.Errorf("C (cache level) must be from 0-4 range")
			}
			cl = v
		}

		fun := os.Getenv(colrgba + "F")
		bFun[colidx] = false
		if fun != "" {
			fctx[colidx] = mfctx.Cpy()
			err := fctx[colidx].FparFunction(fun)
			if err != nil {
				return err
			}
			err = fctx[colidx].FparOK(5)
			if err != nil {
				return err
			}
			fctx[colidx].SetCache(cl, colidx)
			bFun[colidx] = true
		}
		// I (use imaginary part of function result instead of real)
		useImag[colidx] = os.Getenv(colrgba+"I") != ""

		// ENV
		// R red
		rS := os.Getenv(colrgba + "R")
		r := 1.0
		if rS != "" {
			v, err := strconv.ParseFloat(rS, 64)
			if err != nil {
				return err
			}
			r = v
		}

		// G green
		gS := os.Getenv(colrgba + "G")
		g := 1.0
		if gS != "" {
			v, err := strconv.ParseFloat(gS, 64)
			if err != nil {
				return err
			}
			g = v
		}

		// B blue
		bS := os.Getenv(colrgba + "B")
		b := 1.0
		if bS != "" {
			v, err := strconv.ParseFloat(bS, 64)
			if err != nil {
				return err
			}
			b = v
		}
		fact := r + g + b
		if fact <= 0 {
			r = 0.0
			g = 0.0
			b = 0.0
		} else {
			r /= fact
			g /= fact
			b /= fact
		}

		// LO
		loS := os.Getenv(colrgba + "LO")
		lo := 0.0
		if loS != "" {
			v, err := strconv.ParseFloat(loS, 64)
			if err != nil {
				return err
			}
			if v < 0.0 || v > 100.0 {
				return fmt.Errorf("LO must be from 0-100 range")
			}
			lo = v
		}

		// LOI
		loS = os.Getenv(colrgba + "LOI")
		loi := uint16(0)
		if loS != "" {
			v, err := strconv.Atoi(loS)
			if err != nil {
				return err
			}
			if v < 1 || v > 0xffff {
				return fmt.Errorf("LOI must be from 0001-FFFF range")
			}
			loi = uint16(v)
		}

		// HI
		hiS := os.Getenv(colrgba + "HI")
		hi := 0.0
		if hiS != "" {
			v, err := strconv.ParseFloat(hiS, 64)
			if err != nil {
				return err
			}
			if v < 0.0 || v > 100.0 {
				return fmt.Errorf("HI must be from 0-100 range")
			}
			hi = v
		}
		hi = 100 - hi
		if lo >= hi {
			return fmt.Errorf("invalid lo-hi range: %f%% - %f%%", lo, hi)
		}

		// HII
		hiS = os.Getenv(colrgba + "HII")
		hii := uint16(0xffff)
		if hiS != "" {
			v, err := strconv.Atoi(hiS)
			if err != nil {
				return err
			}
			if v < 0 || v > 0xfffe {
				return fmt.Errorf("HII must be from 0000-FFFE range")
			}
			hii = uint16(v)
		}

		// GA gamma
		gaS := os.Getenv(colrgba + "GA")
		ga := 1.0
		gaB := false
		if gaS != "" {
			v, err := strconv.ParseFloat(gaS, 64)
			if err != nil {
				return err
			}
			ga = v
			gaB = true
		}

		// CONT
		conti := uint16(0)
		if !bGlobCont {
			contS := os.Getenv(colrgba + "CONT")
			if contS != "" {
				v, err := strconv.Atoi(contS)
				if err != nil {
					return err
				}
				if v < 1 || v > 0x3fff {
					return fmt.Errorf("CONT must be from 0001-3FFF range")
				}
				conti = uint16(v)
			}
		}

		// SURF and EDGE
		edge := uint16(1)
		if !bGlobEdge {
			edgeS := os.Getenv(colrgba + "EDGE")
			if edgeS != "" {
				v, err := strconv.Atoi(edgeS)
				if err != nil {
					return err
				}
				if v < 0 || v > 3 {
					return fmt.Errorf("EDGE must be from [0, 1, 2, 3]")
				}
				edge = uint16(v)
			}
		}
		surf := uint16(0)
		if !bGlobSurf {
			surfS := os.Getenv(colrgba + "SURF")
			if surfS != "" {
				v, err := strconv.Atoi(surfS)
				if err != nil {
					return err
				}
				if v < 0 || v > 3 {
					return fmt.Errorf("SURF must be from [0, 1, 2, 3]")
				}
				surf = uint16(v)
			}
		}
		gcont := false
		if !bGlobGCont {
			gcontS := os.Getenv(colrgba + "GCONT")
			if gcontS != "" {
				v, err := strconv.Atoi(gcontS)
				if err != nil {
					return err
				}
				if v < 0 || v > 1 {
					return fmt.Errorf("GCONT must be from [0, 1]")
				}
				if v == 1 {
					gcont = true
				}
			}
		}
		ar[colidx] = r
		ag[colidx] = g
		ab[colidx] = b
		alo[colidx] = lo
		aloi[colidx] = loi
		acl[colidx] = cl
		ahi[colidx] = hi
		ahii[colidx] = hii
		agaB[colidx] = gaB
		aga[colidx] = ga
		if !bGlobCont {
			acont[colidx] = conti
		}
		if !bGlobEdge {
			aedge[colidx] = edge
		}
		if !bGlobSurf {
			asurf[colidx] = surf
		}
		if !bGlobGCont {
			agcont[colidx] = gcont
		}

		fmt.Printf(
			"Final %s RGB multiplier: %f(%f, %f, %f), range %f%% - %f%%, idx range: %04x-%04x, cont: %d/%v, surf/edge: %d/%d, quality: %d, gamma: (%v, %f), cache: %d, threads: %d, override: %v,%s,%s\n",
			colrgba, fact, ar[colidx], ag[colidx], ab[colidx], alo[colidx], ahi[colidx], aloi[colidx], ahii[colidx], acont[colidx], agcont[colidx], asurf[colidx], aedge[colidx],
			jpegq, agaB[colidx], aga[colidx], acl[colidx], thrN, overB, overFrom, overTo,
		)
	}

	// Flushing before endline
	flush := bufio.NewWriter(os.Stdout)

	// Function extracting image data
	var (
		getPixelFunc    func(img *image.Image, i, j int) (uint32, uint32, uint32, uint32)
		getPixelFuncAry [4]func(img *image.Image, i, j int) (uint32, uint32, uint32, uint32)
	)
	if inf <= 0 {
		getPixelFunc = func(img *image.Image, i, j int) (uint32, uint32, uint32, uint32) {
			return (*img).At(i, j).RGBA()
		}
	}

	// Iterate given files
	n := len(args)
	for k, fn := range args {
		dtStart := time.Now()
		fk := float64(k) / float64(n)
		fmt.Printf("%d/%d %s...", k+1, n, fn)
		_ = flush.Flush()

		// Input
		dtStartI := time.Now()

		// Hints
		var hint jpegbw.HintData
		usedHint := false
		if useHints {
			data, err := ioutil.ReadFile(fn + ".hint")
			if err != nil {
				if hintRequired {
					return err
				}
				fmt.Printf("Missing hint file: %s.hint\n", fn)
			} else {
				err = json.Unmarshal(data, &hint)
				if err != nil {
					if hintRequired {
						return err
					}
					fmt.Printf("Invalid hint file: %s.hint\n", fn)
				} else {
					usedHint = true
					// info: fmt.Printf("Hint: %+v\n", hint)
				}
			}
		}

		// Image
		reader, err := os.Open(fn)
		if err != nil {
			return err
		}

		// Decode input
		m, _, err := image.Decode(reader)
		if err != nil {
			_ = reader.Close()
			return err
		}
		err = reader.Close()
		if err != nil {
			return err
		}
		bounds := m.Bounds()
		x := bounds.Max.X
		y := bounds.Max.Y
		xo := x
		yo := y
		if inf > 0 {
			x += inf
			y += 2 * inf
			fmt.Printf(" (%d/%d x %d/%d)...", xo, x, yo, y)
		} else {
			fmt.Printf(" (%d x %d)...", x, y)
		}
		dtEndI := time.Now()
		_ = flush.Flush()

		var pxdata [][][4]uint16
		for i := 0; i < x; i++ {
			pxdata = append(pxdata, [][4]uint16{})
			for j := 0; j < y; j++ {
				pxdata[i] = append(pxdata[i], [4]uint16{0, 0, 0, 0})
			}
		}

		// Convert
		all := float64(xo * yo)
		var (
			//at    [4]uint32
			timeF time.Duration
			timeH time.Duration
		)
		passes := 1
		if acm {
			passes = 2
		}
		var (
			loIs    []uint16
			hiIs    []uint16
			mults   []float64
			acmloI  uint16
			acmhiI  uint16
			acmmult float64
		)
		for pass := 0; pass < passes; pass++ {
			if pass == 1 {
				acmloI = uint16(0xffff)
				for ci := range loIs {
					if loIs[ci] < acmloI {
						acmloI = loIs[ci]
					}
					if hiIs[ci] > acmhiI {
						acmhiI = hiIs[ci]
					}
				}
				acmmult = 65535.0 / float64(acmhiI-acmloI)
				fmt.Printf(" ACM int: (%d, %d) mult: %f...", acmloI, acmhiI, acmmult)
				_ = flush.Flush()
			}
			for colidx, colrgba := range rgba {
				if noA && colidx == 3 {
					continue
				}
				var (
					loI  uint16
					hiI  uint16
					mult float64
				)
				r := ar[colidx]
				g := ag[colidx]
				b := ab[colidx]
				ga := aga[colidx]
				gaB := agaB[colidx]
				if pass == 0 {
					if inf <= 0 {
						getPixelFuncAry[colidx] = getPixelFunc
					}
					lo := alo[colidx]
					loi := aloi[colidx]
					hi := ahi[colidx]
					hii := ahii[colidx]

					if useHints && usedHint {
						loi = hint.LoIdx[colidx]
						hii = hint.HiIdx[colidx]
						// info: fmt.Printf("Using hint scale: %04x-%04x\n", loi, hii)
					}

					hist := make(jpegbw.IntHist)
					minGs := uint16(0xffff)
					maxGs := uint16(0)

					dtStartH := time.Now()
					if inf > 0 || loi == 0 || hii == 0xffff {
						for i := 0; i < xo; i++ {
							for j := 0; j < yo; j++ {
								pr, pg, pb, _ := m.At(i, j).RGBA()
								// debug2: fmt.Printf("(%d,%d,%d)\n", pr, pg, pb)
								gs := uint16(r*float64(pr) + g*float64(pg) + b*float64(pb))
								if gs < minGs {
									minGs = gs
								}
								if gs > maxGs {
									maxGs = gs
								}
								hist[gs]++
							}
						}
						// info: fmt.Printf("hist: %+v\n", hist.str())

						// Calculations
						histCum := make(jpegbw.FloatHist)
						sum := int64(0)
						for i := uint16(0); true; i++ {
							sum += hist[i]
							histCum[i] = (float64(sum) * 100.0) / all
							if i == 0xffff {
								break
							}
						}
						for i := uint16(1); true; i++ {
							prev := histCum[i-1]
							next := histCum[i]
							if loI == 0 && prev <= lo && lo <= next {
								loI = i
							}
							if prev <= hi && hi <= next {
								hiI = i
							}
							if i == 0xffff {
								break
							}
						}
						if loi > 0 && loi != loI {
							// info: fmt.Printf("Overwriting %s low index: %04x -> %04x\n", colrgba, loI, loi)
							loI = loi
						}
						if hii < 0xffff && hii != hiI {
							// info: fmt.Printf("Overwriting %s high index: %04x -> %04x\n", colrgba, hiI, hii)
							hiI = hii
						}
						if loI >= hiI {
							return fmt.Errorf("calculated integer range is empty: %d-%d", loI, hiI)
						}
					} else {
						loI = loi
						hiI = hii
					}
					mult = 65535.0 / float64(hiI-loI)
					// info: fmt.Printf("histCum: %+v\n", histCum.str())

					// In INF mode we need histogramScaled context
					if inf > 0 {
						b := 65535.0 / float64(x)
						histScaled := make(jpegbw.IntHist)
						maxHS := int64(0)
						if bhpow {
							for i := uint16(0); i < uint16(x); i++ {
								ff := (float64(i) * b) / 65535.0
								f := uint16(math.Pow(ff, shpow) * 65535.0)
								tf := (float64(i+1) * b) / 65535.0
								t := uint16(math.Pow(tf, shpow) * 65535.0)
								if t == f {
									t++
								}
								hv := int64(0)
								for h := f; h < t; h++ {
									hv += hist[h]
								}
								histScaled[i] = hv
								if hv > maxHS {
									maxHS = hv
								}
							}
						} else {
							for i := uint16(0); i < uint16(x); i++ {
								f := uint16(float64(i) * b)
								t := uint16(float64(i+1) * b)
								if t == f {
									t++
								}
								hv := int64(0)
								for h := f; h < t; h++ {
									hv += hist[h]
								}
								histScaled[i] = hv
								if hv > maxHS {
									maxHS = hv
								}
							}
						}
						fran := float64((hiI - loI) + 1)
						b2 := fran / float64(x)
						histScaled2 := make(jpegbw.IntHist)
						maxHS2 := int64(0)
						for i := uint16(0); i < uint16(x); i++ {
							f := loI + uint16(float64(i)*b2)
							t := uint16(float64(f) + b2)
							if t == f {
								t++
							}
							hv := int64(0)
							for h := f; h < t; h++ {
								hv += hist[h]
							}
							histScaled2[i] = hv
							if hv > maxHS2 {
								maxHS2 = hv
							}
						}
						prev := int64(0)
						next := int64(0)
						prevI := uint16(0xffff)
						for i := uint16(0); i < uint16(x); i++ {
							v := histScaled[i]
							if v > 0 {
								prev = v
								prevI = i
							} else {
								nextJ := uint16(0xffff)
								for j := i + 1; j < uint16(x); j++ {
									w := histScaled[j]
									if w > 0 {
										next = w
										nextJ = j
										break
									}
								}
								if prevI != 0xffff && nextJ != 0xffff {
									histScaled[i] = prev + int64((float64(i-prevI)/float64(nextJ-prevI))*float64(next-prev))
								}
							}
						}
						prev = 0
						next = 0
						prevI = uint16(0xffff)
						for i := uint16(0); i < uint16(x); i++ {
							v := histScaled2[i]
							if v > 0 {
								prev = v
								prevI = i
							} else {
								nextJ := uint16(0xffff)
								for j := i + 1; j < uint16(x); j++ {
									w := histScaled2[j]
									if w > 0 {
										next = w
										nextJ = j
										break
									}
								}
								if prevI != 0xffff && nextJ != 0xffff {
									histScaled2[i] = prev + int64((float64(i-prevI)/float64(nextJ-prevI))*float64(next-prev))
								}
							}
						}
						maxHSF := float64(maxHS)
						maxHSF2 := float64(maxHS2)
						finf := float64(inf * 2)
						// debug: fmt.Printf("histScaled: %+v\n", histScaled.str())
						ran := (hiI - loI) + 1
						ran4 := (ran + 1) / 4
						if ran == 0 {
							ran = 0xffff
						}
						if ran4 == 0 {
							ran4 = 0x4000
						}
						getPixelFunc = func(img *image.Image, i, j int) (uint32, uint32, uint32, uint32) {
							if i < x-inf && j < y-(2*inf) {
								// normal pixel
								return (*img).At(i, j).RGBA()
							} else if j < y-(2*inf) {
								// scale on the right: GS or GS, R, G, B
								if einf {
									g := (uint32(j) * uint32(ran)) / uint32(y-2*inf)
									d := g / uint32(ran4)
									r := uint32(hiI) - ((g % uint32(ran4)) << 2)
									switch d {
									case 0:
										return r, r, r, uint32(0xffff)
									case 1:
										return r, 0, 0, uint32(0xffff)
									case 2:
										return 0, r, 0, uint32(0xffff)
									default:
										return 0, 0, r, uint32(0xffff)
									}
								} else {
									g := uint32(hiI) - ((uint32(j) * uint32(ran)) / uint32(y-2*inf))
									return g, g, g, uint32(0xffff)
								}
							}
							// 2 histograms on the botton: scaled & absolute
							cv := float64((y-j)-1) / finf
							ncv := cv * 2.
							g := uint32(0xffff)
							if cv < .5 {
								hv := float64(histScaled[uint16(i)]) / maxHSF
								if ncv >= hv {
									g = uint32(0)
								}
							} else {
								ncv -= 1.
								hv2 := float64(histScaled2[uint16(i)]) / maxHSF2
								if ncv >= hv2 {
									g = uint32(0)
								}
							}
							return g, g, g, uint32(0xffff)
						}
						getPixelFuncAry[colidx] = getPixelFunc
					}

					dtEndH := time.Now()
					timeH += dtEndH.Sub(dtStartH)
					fmt.Printf(" %s: (%d, %d) int: (%d, %d) mult: %f...", colrgba, minGs, maxGs, loI, hiI, mult)
					// info: fmt.Printf("histCum: %+v\n", histCum.str())
					_ = flush.Flush()
					if acm {
						loIs = append(loIs, loI)
						hiIs = append(hiIs, hiI)
						mults = append(mults, mult)
					}
				}
				if !acm || pass == 1 {
					if acm {
						if acmFact > 0.999999 {
							loI = acmloI
							hiI = acmhiI
							mult = acmmult
						} else if acmFact <= 0.000001 {
							loI = loIs[colidx]
							hiI = hiIs[colidx]
							mult = mults[colidx]
						} else {
							loI = uint16(float64(loIs[colidx]) - acmFact*float64(loIs[colidx]-acmloI))
							hiI = uint16(float64(hiIs[colidx]) + acmFact*float64(acmhiI-hiIs[colidx]))
							mult = mults[colidx] - acmFact*(mults[colidx]-acmmult)
							fmt.Printf(" ACM(%f) int: (%d-%d->%d, %d-%d->%d) mult: %f-%f->%f...", acmFact, acmloI, loIs[colidx], loI, hiIs[colidx], acmhiI, hiI, acmmult, mults[colidx], mult)
							_ = flush.Flush()
						}
						getPixelFunc = getPixelFuncAry[colidx]
					}
					che := make(chan error)
					nThreads := 0
					ctxa := []jpegbw.FparCtx{}
					ctxInUse := make(map[int]bool)
					for i := 0; i < thrN; i++ {
						ctxa = append(ctxa, fctx[colidx].Cpy())
						ctxInUse[i] = false
					}

					// calculations for current color
					var cmtx = &sync.Mutex{}
					dtStartF := time.Now()
					for ii := 0; ii < x; ii++ {
						go func(c chan error, i int) {
							// debug: fmt.Printf("line: %d/%d\n", i, x)
							cmtx.Lock()
							cNum := -1
							for t := 0; t < thrN; t++ {
								if !ctxInUse[t] {
									cNum = t
									ctxInUse[cNum] = true
									break
								}
							}
							cmtx.Unlock()
							if cNum < 0 {
								// Sync
								c <- fmt.Errorf("no context copy available: i=%d", i)
								return
							}
							fi := float64(i) / float64(x)
							trace := 1.0
							cv := uint32(0)
							for j := 0; j < y; j++ {
								fj := float64(j) / float64(y)
								pr, pg, pb, pa := getPixelFunc(&m, i, j)
								switch colidx {
								case 0:
									cv = pr
								case 1:
									cv = pg
								case 2:
									cv = pb
								default:
									cv = pa
								}
								//if inf > 0 && (i >= xo || j >= yo) {
								if inf > 0 && j >= yo {
									switch colidx {
									case 0:
										pxdata[i][j][colidx] = uint16(pr)
									case 1:
										pxdata[i][j][colidx] = uint16(pg)
									case 2:
										pxdata[i][j][colidx] = uint16(pb)
									default:
										pxdata[i][j][colidx] = uint16(pa)
									}
									continue
								}
								gs := uint16(r*float64(pr) + g*float64(pg) + b*float64(pb))
								iv := int(gs) - int(loI)
								if iv < 0 {
									iv = 0
								}
								fv := float64(iv) * mult
								if fv > 65535.0 {
									fv = 65535.0
								}
								if gaB {
									fv = math.Pow(fv/65535.0, ga) * 65535.0
									if fv < 0.0 {
										fv = 0.0
									}
									if fv > 65535.0 {
										fv = 65535.0
									}
								}
								if bFun[colidx] {
									var e error
									cv, e := ctxa[cNum].FparF(
										[]complex128{
											complex(fv/65535.0, 0.0),
											complex(fi, fj),
											complex(float64(pr)/65535.0, float64(pg)/65535.0),
											complex(float64(pb)/65535.0, float64(pa)/65535.0),
											complex(fk, trace),
										},
									)
									if e != nil {
										// Sync
										cmtx.Lock()
										ctxInUse[cNum] = false
										cmtx.Unlock()
										c <- e
										return
									}
									if useImag[colidx] {
										fv = imag(cv)
									} else {
										fv = real(cv)
									}
									trace = fv
									// trace: fmt.Printf("trace is: %v\n", trace)
									fv *= 65535.0
									if fv < 0.0 {
										fv = 0.0
									}
									if fv > 65535.0 {
										fv = 65535.0
									}
								}
								if rev {
									delta := int(fv) - int(cv)
									set := int(cv) - delta
									if set < 0 {
										set = 0
									}
									if set > 0xffff {
										set = 0xffff
									}
									// fmt.Printf("curr = %d, new = %d, delta = %d, set = %d\n", cv, int(fv), delta, set)
									pxdata[i][j][colidx] = uint16(set)

								} else {
									pxdata[i][j][colidx] = uint16(fv)
								}
							}
							// Sync
							cmtx.Lock()
							ctxInUse[cNum] = false
							cmtx.Unlock()
							c <- nil
						}(che, ii)

						// Keep maximum number of threads
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
					dtEndF := time.Now()
					timeF += dtEndF.Sub(dtStartF)
				}
			}
		}

		// Hanlde contours algorithm
		contB := false
		for colidx := range rgba {
			if noA && colidx == 3 {
				continue
			}
			cont := acont[colidx]
			if cont > 0 {
				contB = true
				break
			}
		}
		if contB {
			dtContStart := time.Now()
			var tpxdata [][][4]uint16
			for i := 0; i < x; i++ {
				tpxdata = append(tpxdata, [][4]uint16{})
				for j := 0; j < y; j++ {
					d := pxdata[i][j]
					tpxdata[i] = append(tpxdata[i], [4]uint16{d[0], d[1], d[2], d[3]})
				}
			}
			for colidx := range rgba {
				if noA && colidx == 3 {
					continue
				}
				cont := acont[colidx] + 1
				if cont < 2 {
					continue
				}
				surf := asurf[colidx]
				edge := aedge[colidx]
				gcont := agcont[colidx]

				colidxF := colidx
				colidxT := colidx
				if gcont {
					colidxF = 0
					if noA {
						colidxT = 2
					} else {
						colidxT = 3
					}
				}
				che := make(chan error)
				nThreads := 0
				contourFunc := func(c chan error, i int) {
					contours := []uint16{}
					for t := uint16(1); t < cont; t++ {
						contours = append(contours, uint16((uint32(t)*uint32(0xffff))/uint32(cont)))
					}
					i1 := i - 1
					i2 := i + 1
					if i1 < 0 {
						i1 = 0
					}
					if i2 >= x {
						i2 = x - 1
					}
					co := false
					for ci := colidxF; ci <= colidxT; ci++ {
						di1 := tpxdata[i1][0][ci]
						di2 := tpxdata[i2][0][ci]
						dj1 := tpxdata[i][0][ci]
						dj2 := tpxdata[i][1][ci]
						for _, contour := range contours {
							if (di1 < contour && di2 >= contour) || (dj1 < contour && dj2 >= contour) || (di1 > contour && di2 <= contour) || (dj1 > contour && dj2 <= contour) {
								if edge == 0 || edge == 1 {
									pxdata[i][0][colidx] = uint16(0xffff * edge)
								} else if edge == 2 {
									pxdata[i][0][colidx] = tpxdata[i][0][colidx]
								} else if edge == 3 {
									pxdata[i][0][colidx] = uint16(0xffff) - tpxdata[i][0][colidx]
								}
								co = true
								break
							}
						}
					}
					if !co {
						if surf == 0 || surf == 1 {
							pxdata[i][0][colidx] = uint16(0xffff * surf)
						} else if surf == 2 {
							pxdata[i][0][colidx] = tpxdata[i][0][colidx]
						} else if surf == 3 {
							pxdata[i][0][colidx] = uint16(0xffff) - tpxdata[i][0][colidx]
						}
					}
					yp := y - 1
					co = false
					for ci := colidxF; ci <= colidxT; ci++ {
						di1 := tpxdata[i1][yp][ci]
						di2 := tpxdata[i2][yp][ci]
						dj1 := tpxdata[i][yp-1][ci]
						dj2 := tpxdata[i][yp][ci]
						for _, contour := range contours {
							if (di1 < contour && di2 >= contour) || (dj1 < contour && dj2 >= contour) || (di1 > contour && di2 <= contour) || (dj1 > contour && dj2 <= contour) {
								if edge == 0 || edge == 1 {
									pxdata[i][yp][colidx] = uint16(0xffff * edge)
								} else if edge == 2 {
									pxdata[i][yp][colidx] = tpxdata[i][yp][colidx]
								} else if edge == 3 {
									pxdata[i][yp][colidx] = uint16(0xffff) - tpxdata[i][yp][colidx]
								}
								co = true
								break
							}
						}
					}
					if !co {
						pxdata[i][yp][colidx] = uint16(0)
						if surf == 0 || surf == 1 {
							pxdata[i][yp][colidx] = uint16(0xffff * surf)
						} else if surf == 2 {
							pxdata[i][yp][colidx] = tpxdata[i][yp][colidx]
						} else if surf == 3 {
							pxdata[i][yp][colidx] = uint16(0xffff) - tpxdata[i][yp][colidx]
						}
					}
					for j := 1; j < yp; j++ {
						j1 := j - 1
						j2 := j + 1
						co = false
						for ci := colidxF; ci <= colidxT; ci++ {
							di1 := tpxdata[i1][j][ci]
							di2 := tpxdata[i2][j][ci]
							dj1 := tpxdata[i][j1][ci]
							dj2 := tpxdata[i][j2][ci]
							for _, contour := range contours {
								if (di1 < contour && di2 >= contour) || (dj1 < contour && dj2 >= contour) || (di1 > contour && di2 <= contour) || (dj1 > contour && dj2 <= contour) {
									if edge == 0 || edge == 1 {
										pxdata[i][j][colidx] = uint16(0xffff * edge)
									} else if edge == 2 {
										pxdata[i][j][colidx] = tpxdata[i][j][colidx]
									} else if edge == 3 {
										pxdata[i][j][colidx] = uint16(0xffff) - tpxdata[i][j][colidx]
									}
									co = true
								}
							}
						}
						if !co {
							if surf == 0 || surf == 1 {
								pxdata[i][j][colidx] = uint16(0xffff * surf)
							} else if surf == 2 {
								pxdata[i][j][colidx] = tpxdata[i][j][colidx]
							} else if surf == 3 {
								pxdata[i][j][colidx] = uint16(0xffff) - tpxdata[i][j][colidx]
							}
						}
					}
					c <- nil
				}
				for ii := 0; ii < x; ii++ {
					go contourFunc(che, ii)
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
			}
			dtContEnd := time.Now()
			contTime := dtContEnd.Sub(dtContStart)
			fmt.Printf(" contours (%+v)...", contTime)
		}

		if ir3cfg.enabled {
			dtIR3Start := time.Now()
			err = applyIR3(pxdata, x, y, thrN, ir3cfg)
			if err != nil {
				return err
			}
			dtIR3End := time.Now()
			ir3Time := dtIR3End.Sub(dtIR3Start)
			timeF += ir3Time
			fmt.Printf(" ir3 (%+v)...", ir3Time)
		}

		if monocfg.enabled {
			dtMonoStart := time.Now()
			err = applyMonoValue(pxdata, x, y, thrN, monocfg)
			if err != nil {
				return err
			}
			dtMonoEnd := time.Now()
			monoTime := dtMonoEnd.Sub(dtMonoStart)
			timeF += monoTime
			fmt.Printf(" monoval (%+v)...", monoTime)
		}

		if isovalcfg.enabled {
			dtIsoValStart := time.Now()
			err = applyIsoVal(pxdata, x, y, thrN, isovalcfg)
			if err != nil {
				return err
			}
			dtIsoValEnd := time.Now()
			isoValTime := dtIsoValEnd.Sub(dtIsoValStart)
			timeF += isoValTime
			fmt.Printf(" isoval (%+v)...", isoValTime)
		}

		// Final write to target
		var (
			target   *image.RGBA64
			targetGS *image.Gray16
		)
		if ogs {
			targetGS = image.NewGray16(image.Rect(0, 0, x, y))
		} else {
			target = image.NewRGBA64(image.Rect(0, 0, x, y))
		}
		dtStartF := time.Now()
		che := make(chan error)
		nThreads := 0
		var fCalc func(chan error, int)
		if ogs {
			fCalc = func(c chan error, i int) {
				for j := 0; j < y; j++ {
					px := pxdata[i][j]
					targetGS.Set(i, j, color.Gray16{uint16(float64(px[0])*gsr + float64(px[1])*gsg + float64(px[2])*gsb)})
				}
				c <- nil
			}
		} else {
			if noA {
				fCalc = func(c chan error, i int) {
					for j := 0; j < y; j++ {
						px := pxdata[i][j]
						//if i%100 == 0 && j%100 == 0 {
						//	fmt.Printf("(%d,%d) --> %v\n", i, j, px)
						//}
						target.Set(i, j, color.RGBA64{px[0], px[1], px[2], 0xffff})
					}
					c <- nil
				}
			} else {
				fCalc = func(c chan error, i int) {
					for j := 0; j < y; j++ {
						px := pxdata[i][j]
						//if i%100 == 0 && j%100 == 0 {
						//	fmt.Printf("(%d,%d) --> %v\n", i, j, px)
						//}
						//px[0] = uint16((uint32(px[0]) * uint32(px[3])) >> 0x10)
						//px[1] = uint16((uint32(px[1]) * uint32(px[3])) >> 0x10)
						//px[2] = uint16((uint32(px[2]) * uint32(px[3])) >> 0x10)
						target.Set(i, j, color.NRGBA64{px[0], px[1], px[2], px[3]})
					}
					c <- nil
				}
			}
		}
		for ii := 0; ii < x; ii++ {
			go fCalc(che, ii)
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
		dtEndF := time.Now()
		timeF += dtEndF.Sub(dtStartF)
		pps := (all / timeF.Seconds()) / 1048576.0

		// Eventual file name override
		ifn := fn
		if overB {
			ifn = strings.Replace(fn, overFrom, overTo, -1)
		}
		// info: fmt.Printf("filename: %s -> %s\n", fn, ifn)

		// Output name
		ary := strings.Split(ifn, "/")
		lAry := len(ary)
		last := ary[lAry-1]
		ary[lAry-1] = "co_" + last
		ofn := strings.Join(ary, "/")
		fi, err := os.Create(ofn)
		if err != nil {
			return err
		}
		lfn := strings.ToLower(ifn)
		// info: fmt.Printf("output filename: %s, lower case %s\n", ofn, lfn)

		// Output write
		dtStartO := time.Now()
		var (
			ierr error
			t    image.Image
		)
		if ogs {
			t = targetGS
		} else {
			t = target
		}
		if strings.Contains(lfn, ".png") {
			enc := png.Encoder{CompressionLevel: pngq}
			ierr = enc.Encode(fi, t)
		} else if strings.Contains(lfn, ".jpg") || strings.Contains(lfn, ".jpeg") {
			var jopts *jpeg.Options
			if jpegq >= 0 {
				jopts = &jpeg.Options{Quality: jpegq}
			}
			ierr = jpeg.Encode(fi, t, jopts)
		} else if strings.Contains(lfn, ".gif") {
			ierr = gif.Encode(fi, t, nil)
		}
		if ierr != nil {
			_ = fi.Close()
			return ierr
		}
		err = fi.Close()
		if err != nil {
			return err
		}
		dtEnd := time.Now()
		fmt.Printf(
			" %s (time %v, load %v, hist %v, calc %v, save %v, MPPS: %.3f)\n",
			ofn, dtEnd.Sub(dtStart), dtEndI.Sub(dtStartI), timeH, timeF, dtEnd.Sub(dtStartO), pps,
		)
	}
	return nil
}

func main() {
	dtStart := time.Now()
	if len(os.Args) > 1 {
		err := images2RGBA(os.Args[1:])
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
	} else {
		fmt.Printf("Please provide at least one image to convert\n")
		helpStr := `
Environment variables:
This program manipulates 4 channels R, G, B, A.
When you see X replace it with R, G, B or A.
NA - skip alpha calculation, alpha will be 1 everywhere
OGS - create grayscale output
GSR - when OGS output, use this amount of R to generate final GS pixel
GSG - when OGS output, use this amount of G to generate final GS pixel
GSB - when OGS output, use this amount of B to generate final GS pixel
(GSR+GSG+GSB) when OGS output - will be normalized to sum to 1, so their sum must be positive
ACM - all colors multiplier mode - will calculate minimum index in all RGBA (4 channels), max in RGBA and then mult from this range
ACM - you can set variable ACM value from ACM=0 (as if it is not specified) to ACM=1, ACM=0.5 will do average between ACM and non-ACM mode
HINT - use hints saved for every file "file.ext" - "file.ext.hint", if no hint is given warning is displayed
HINTREQ - make hint file required
Q - jpeg quality 1-100, will use library default if not specified
PQ - png quality 0-3 (0 is default): 0=DefaultCompression, 1=NoCompression, 2=BestSpeed, 3=BestCompression
XR - relative red usage for generating gray pixel, 1 if not specified
XG - relative green usage for generating gray pixel, 1 if not specified
XB - relative blue usage for generating gray pixel, 1 if not specified
(R+G+B) will be normalized to sum to 1, so their sum must be positive
R=0.2125 G=0.7154 B=0.0721 is a suggested configuration
XLO - when calculating intensity range, discard values than are in this lower %, for example 3
XHI - when calculating intensity range, discard values that are in this higher %, for example 3
XLOI - when calculating intensity range, discard values than are lower than this (range is 0000-FFFF)
XHII - when calculating intensity range, discard values that are higher than this (range is 0000-FFFF)
XGA - gamma default 1, which uses straight line (0,0) -> (1,1), if set uses (x,y)->(x,pow(x, GA)) mapping
XCONT - hanlde countour lines, RCONT=10 will draw 10 countour lines for red color
CONT - set countours to the same value for all R, G, B, A channels
EDGE - in countour algorithm, set edge mode: 0, 1, 2 (original), 3 (invert)
SURF - in countour algorithm, set surface (non-edge) mode: 0, 1, 2 (original), 3 (invert)
GCONT - set global contours mode, if set contour is detected when R, G, B or A detects contour
XGCONT - set global contour for one color say RGCONT=1 (red will detect contour when R, G, B, A has contour)
XEDGE, XSURF, XGCONT - set per color EDGE/SURF/GCONT params (unless global specified)
XF - function to apply on final 0-1 range, for example "sin(x1*2)+cos(x1*3)"
XC - function cache level (0-no cache, 1-1st arg caching, 2-1st and 2nd arg caching, ... 4 - 4 args caching)
LIB - if F is used and F calls external functions, thery need to be loaded for this C library
NF - set maximum number of distinct functions in the parser, if not set, default 128 is used
XI - use imaginary part of fuction return value instead of real, use like I=1
N - set number of CPUs to process data
O - eventual overwite file name config, example: ".jpg:.png"
INF - set additional info on image size is N when INF=N
EINF - more complex info.
HPOW - INF histogram 0-0x10000 --> 0-1 --> x. f(x) = pow(x, HPOW). Default 1
REV - reverse the calculation
MONOVAL / MVMODE - final mono-value stage, one of: luma, linear, hsv, hsl, oklch
  luma   - flatten the same weighted channel-value measure that jpeg/jpegbw grayscale uses
  linear - flatten linear-RGB luminance using MVR/MVG/MVB weights
  hsv    - keep hue and saturation, replace HSV value with MVT
  hsl    - keep hue and saturation, replace HSL lightness with MVT
  oklch  - keep OKLCh hue/chroma, replace OK lightness with MVT, optionally gamut-fit
MVT - target flattened value/lightness/luminance, 0-1, default 0.5
MVR/MVG/MVB - weights for luma/linear modes, default 0.2126/0.7152/0.0722, normalized internally
MVGAMUT - fit or clip, default fit. fit reduces chroma to stay inside gamut
MVZERO - gray or black, default gray. controls undefined-hue / zero-luma pixels in luma/linear modes
MVS - optional saturation override for hsv/hsl modes. if unset, original saturation is preserved
MVC - optional chroma override for oklch mode. if unset, original chroma is preserved
ISOVAL - equalize weighted RGB value/lightness across the image, modes: add or mul
  add - adds/subtracts the same delta to R,G,B so IVR*R+IVG*G+IVB*B reaches IVT (with clipping)
  mul - multiplies/divides R,G,B by the same factor so IVR*R+IVG*G+IVB*B reaches IVT (with clipping)
IVT - ISOVAL target value 0-1, default 0.5
IVR/IVG/IVB - ISOVAL value weights, default 0.2126/0.7152/0.0722, normalized internally
`
		fmt.Printf("%s\n", helpStr)
	}
	dtEnd := time.Now()
	fmt.Printf("Time: %v\n", dtEnd.Sub(dtStart))
}
