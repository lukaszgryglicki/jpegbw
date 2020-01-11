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
		if v > 3 {
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
		if v > 3 {
			return fmt.Errorf("SURF must be from [0, 1, 2, 3]")
		}
		surf = uint16(v)
		bGlobSurf = true
		asurf = [4]uint16{surf, surf, surf, surf}
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

	// Additional in-image  info: R,G,B,Gs scale to the right and RGB histogram on the bottom
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
				if v > 3 {
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
				if v > 3 {
					return fmt.Errorf("SURF must be from [0, 1, 2, 3]")
				}
				surf = uint16(v)
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

		fmt.Printf(
			"Final %s RGB multiplier: %f(%f, %f, %f), range %f%% - %f%%, idx range: %04x-%04x, cont: %d, surf/edge: %d/%d, quality: %d, gamma: (%v, %f), cache: %d, threads: %d, override: %v,%s,%s\n",
			colrgba, fact, ar[colidx], ag[colidx], ab[colidx], alo[colidx], ahi[colidx], aloi[colidx], ahii[colidx], acont[colidx], asurf[colidx], aedge[colidx],
			jpegq, agaB[colidx], aga[colidx], acl[colidx], thrN, overB, overFrom, overTo,
		)
	}

	// Flushing before endline
	flush := bufio.NewWriter(os.Stdout)

	// Function extracting image data
	var getPixelFunc func(img *image.Image, i, j int) (uint32, uint32, uint32, uint32)
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
		for colidx, colrgba := range rgba {
			if noA && colidx == 3 {
				continue
			}
			r := ar[colidx]
			g := ag[colidx]
			b := ab[colidx]
			lo := alo[colidx]
			loi := aloi[colidx]
			hi := ahi[colidx]
			hii := ahii[colidx]
			ga := aga[colidx]
			gaB := agaB[colidx]

			if useHints && usedHint {
				loi = hint.LoIdx[colidx]
				hii = hint.HiIdx[colidx]
				// info: fmt.Printf("Using hint scale: %04x-%04x\n", loi, hii)
			}

			hist := make(jpegbw.IntHist)
			minGs := uint16(0xffff)
			maxGs := uint16(0)

			dtStartH := time.Now()
			loI := uint16(0)
			hiI := uint16(0)
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
			mult := 65535.0 / float64(hiI-loI)
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
			}

			dtEndH := time.Now()
			timeH += dtEndH.Sub(dtStartH)
			fmt.Printf(" %s: (%d, %d) int: (%d, %d) mult: %f...", colrgba, minGs, maxGs, loI, hiI, mult)
			// info: fmt.Printf("histCum: %+v\n", histCum.str())
			_ = flush.Flush()

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
					for j := 0; j < y; j++ {
						fj := float64(j) / float64(y)
						pr, pg, pb, pa := getPixelFunc(&m, i, j)
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
						pxdata[i][j][colidx] = uint16(fv)
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
					di1 := tpxdata[i1][0][colidx]
					di2 := tpxdata[i2][0][colidx]
					dj1 := tpxdata[i][0][colidx]
					dj2 := tpxdata[i][1][colidx]
					co := false
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
					di1 = tpxdata[i1][yp][colidx]
					di2 = tpxdata[i2][yp][colidx]
					dj1 = tpxdata[i][yp-1][colidx]
					dj2 = tpxdata[i][yp][colidx]
					co = false
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
						di1 = tpxdata[i1][j][colidx]
						di2 = tpxdata[i2][j][colidx]
						dj1 = tpxdata[i][j1][colidx]
						dj2 = tpxdata[i][j2][colidx]
						co = false
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
XEDGE, XSURF - set per color EDGE/SURF params (unless global specified)
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
`
		fmt.Printf("%s\n", helpStr)
	}
	dtEnd := time.Now()
	fmt.Printf("Time: %v\n", dtEnd.Sub(dtStart))
}
