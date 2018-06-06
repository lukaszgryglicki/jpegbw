package main

import (
	"bufio"
	"fmt"
	"image"
	"image/color"
	"image/gif"
	"image/jpeg"
	"image/png"
	"jpegbw"
	"math"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

type intHist map[uint16]int
type floatHist map[uint16]float64

func (m intHist) str() string {
	s := ""
	for i := uint16(0); true; i++ {
		v := m[i]
		if v > 0 {
			s += fmt.Sprintf("%d => %d\n", i, m[i])
		}
		if i == 0xffff {
			break
		}
	}
	return s
}

func (m floatHist) str() string {
	s := ""
	prev := -1.0
	for i := uint16(0); true; i++ {
		v := m[i]
		if v > 0.00001 && v < 99.99999 && math.Abs(v-prev) > 0.00001 {
			s += fmt.Sprintf("%d => %.5f%%\n", i, m[i])
		}
		prev = v
		if i == 0xffff {
			break
		}
	}
	return s
}

// images2RGBA: convert given images to bw: iname.ext -> co_iname.ext, dir/iname.ext -> dir/co_iname.ext
// Other parameters are set via env variables (see main() function it describes all env params):
func images2RGBA(args []string) error {
	// Quality
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
		ahi     [4]float64
		aga     [4]float64
		agaB    [4]bool
	)

	// Process colors
	for colidx, colrgba := range rgba {
		fun := os.Getenv(colrgba + "F")
		lib := ""
		bFun[colidx] = false
		if fun != "" {
			lib = os.Getenv(colrgba + "LIB")
			if lib != "" {
				nf := 128
				nfs := os.Getenv(colrgba + "NF")
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
				ok := fctx[colidx].Init(lib, uint(nf))
				if !ok {
					return fmt.Errorf("LIB init failed for: %s", lib)
				}
				defer func() { fctx[colidx].Tidy() }()
			}
			err := fctx[colidx].FparFunction(fun)
			if err != nil {
				return err
			}
			err = fctx[colidx].FparOK(5)
			if err != nil {
				return err
			}
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
		ar[colidx] = r
		ag[colidx] = g
		ab[colidx] = b
		alo[colidx] = lo
		ahi[colidx] = hi
		agaB[colidx] = gaB
		aga[colidx] = ga

		fmt.Printf(
			"Final %s RGB multiplier: %f(%f, %f, %f), range %f%% - %f%%, quality: %d, gamma: (%v, %f), threads: %d, override: %v,%s,%s\n",
			colrgba, fact, ar[colidx], ag[colidx], ab[colidx], alo[colidx], ahi[colidx],
			jpegq, agaB[colidx], aga[colidx], thrN, overB, overFrom, overTo,
		)
	}

	// Flushing before endline
	flush := bufio.NewWriter(os.Stdout)

	// Iterate given files
	n := len(args)
	for k, fn := range args {
		dtStart := time.Now()
		fk := float64(k) / float64(n)
		fmt.Printf("%d/%d %s...", k+1, n, fn)
		_ = flush.Flush()

		// Input
		dtStartI := time.Now()
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
		bounds := m.Bounds()
		x := bounds.Max.X
		y := bounds.Max.Y
		dtEndI := time.Now()
		fmt.Printf(" (%d x %d)...", x, y)
		_ = flush.Flush()

		// Output
		target := image.NewRGBA64(image.Rect(0, 0, x, y))
		var pxdata [][][4]uint16
		for i := 0; i < x; i++ {
			pxdata = append(pxdata, [][4]uint16{})
			for j := 0; j < y; j++ {
				pxdata[i] = append(pxdata[i], [4]uint16{0, 0, 0, 0})
			}
		}

		// Convert
		all := float64(x * y)
		var (
			//at    [4]uint32
			timeF time.Duration
			timeH time.Duration
		)
		for colidx, colrgba := range rgba {
			r := ar[colidx]
			g := ag[colidx]
			b := ab[colidx]
			lo := alo[colidx]
			hi := ahi[colidx]
			ga := aga[colidx]
			gaB := agaB[colidx]

			hist := make(intHist)
			minGs := uint16(0xffff)
			maxGs := uint16(0)

			dtStartH := time.Now()
			for i := 0; i < x; i++ {
				for j := 0; j < y; j++ {
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
			histCum := make(floatHist)
			sum := 0
			for i := uint16(0); true; i++ {
				sum += hist[i]
				histCum[i] = (float64(sum) * 100.0) / all
				if i == 0xffff {
					break
				}
			}
			loI := uint16(0)
			hiI := uint16(0)
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
			if loI >= hiI {
				_ = reader.Close()
				return fmt.Errorf("calculated integer range is empty: %d-%d", loI, hiI)
			}
			mult := 65535.0 / float64(hiI-loI)
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
						pr, pg, pb, pa := m.At(i, j).RGBA()
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
						_ = reader.Close()
						return e
					}
					nThreads--
				}
			}
			for nThreads > 0 {
				e := <-che
				if e != nil {
					_ = reader.Close()
					return e
				}
				nThreads--
			}
			dtEndF := time.Now()
			timeF += dtEndF.Sub(dtStartF)
		}

		// Final write to target
		dtStartF := time.Now()
		che := make(chan error)
		nThreads := 0
		for ii := 0; ii < x; ii++ {
			go func(c chan error, i int) {
				for j := 0; j < y; j++ {
					px := pxdata[i][j]
					target.Set(i, j, color.RGBA64{px[0], px[1], px[2], px[3]})
				}
				c <- nil
			}(che, ii)
			nThreads++
			if nThreads == thrN {
				e := <-che
				if e != nil {
					_ = reader.Close()
					return e
				}
				nThreads--
			}
		}
		for nThreads > 0 {
			e := <-che
			if e != nil {
				_ = reader.Close()
				return e
			}
			nThreads--
		}
		dtEndF := time.Now()
		timeF += dtEndF.Sub(dtStartF)
		pps := (all / timeF.Seconds()) / 1048576.0

		// Close reader
		err = reader.Close()
		if err != nil {
			return err
		}

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
		var ierr error
		if strings.Contains(lfn, ".png") {
			ierr = png.Encode(fi, target)
		} else if strings.Contains(lfn, ".jpg") || strings.Contains(lfn, ".jpeg") {
			if jpegq < 0 {
				ierr = jpeg.Encode(fi, target, nil)
			} else {
				ierr = jpeg.Encode(fi, target, &jpeg.Options{Quality: jpegq})
			}
		} else if strings.Contains(lfn, ".gif") {
			ierr = gif.Encode(fi, target, nil)
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
		}
	} else {
		fmt.Printf("Please provide at least one image to convert\n")
		helpStr := `
Environment variables:
This program manipulates 4 channels R, G, B, A.
When you see X replace it with R, G, B or A.
Q - jpeg quality 1-100, will use library default if not specified
XR - relative red usage for generating gray pixel, 1 if not specified
XG - relative green usage for generating gray pixel, 1 if not specified
XB - relative blue usage for generating gray pixel, 1 if not specified
(R+G+B) will be normalized to sum to 1, so their sum must be positive
R=0.2125 G=0.7154 B=0.0721 is a suggested configuration
XLO - when calculating intensity range, discard values than are in this lower %, for example 3
XHI - when calculating intensity range, discard values that are in this higher %, for example 3
XGA - gamma default 1, which uses straight line (0,0) -> (1,1), if set uses (x,y)->(x,pow(x, GA)) mapping
XF - function to apply on final 0-1 range, for example "sin(x1*2)+cos(x1*3)"
XLIB - if F is used and F calls external functions, thery need to be loaded for this C library
XNF - set maximum number of distinct functions in the parser, if not set, default 128 is used
XI - use imaginary part of fuction return value instead of real, use like I=1
N - set number of CPUs to process data
O - eventual overwite file name config, example: ".jpg:.png"
`
		fmt.Printf("%s\n", helpStr)
	}
	dtEnd := time.Now()
	fmt.Printf("Time: %v\n", dtEnd.Sub(dtStart))
}
