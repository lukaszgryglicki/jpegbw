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

// images2BW: convert given images to bw: iname.ext -> bw_iname.ext, dir/iname.ext -> dir/bw_iname.ext
// Other parameters are set via env variables (see main() function it describes all env params):
func images2BW(args []string) error {
	// F, LIB processing
	var fctx jpegbw.FparCtx
	fun := os.Getenv("F")
	lib := ""
	bFun := false
	if fun != "" {
		lib = os.Getenv("LIB")
		if lib != "" {
			ok := fctx.Init(lib)
			if !ok {
				return fmt.Errorf("LIB init failed for: %s", lib)
			}
			defer func() { fctx.Tidy() }()
		}
		err := fctx.FparFunction(fun)
		if err != nil {
			return err
		}
		err = fctx.FparOK(4)
		if err != nil {
			return err
		}
		bFun = true
	}

	// ENV
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

	// R red
	rS := os.Getenv("R")
	r := 1.0
	if rS != "" {
		v, err := strconv.ParseFloat(rS, 64)
		if err != nil {
			return err
		}
		r = v
	}

	// G green
	gS := os.Getenv("G")
	g := 1.0
	if gS != "" {
		v, err := strconv.ParseFloat(gS, 64)
		if err != nil {
			return err
		}
		g = v
	}

	// B blue
	bS := os.Getenv("B")
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
		return fmt.Errorf("r+g+b is <= 0: %v", fact)
	}
	r /= fact
	g /= fact
	b /= fact

	// LO
	loS := os.Getenv("LO")
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
	hiS := os.Getenv("HI")
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
	gaS := os.Getenv("GA")
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
	fmt.Printf(
		"Final RGB multiplier: %f(%f, %f, %f), range %f%% - %f%%, quality: %d, gamma: (%v, %f), threads: %d, override: %v,%s,%s\n",
		fact, r, g, b, lo, hi, jpegq, gaB, ga, thrN, overB, overFrom, overTo,
	)

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
		target := image.NewGray16(image.Rect(0, 0, x, y))

		// Convert
		hist := make(intHist)
		minGs := uint16(0xffff)
		maxGs := uint16(0)

		dtStartH := time.Now()
		for i := 0; i < x; i++ {
			for j := 0; j < y; j++ {
				// target.Set(i, j, m.At(i, j))
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
		all := float64(x * y)
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
		fmt.Printf(" gray: (%d, %d) int: (%d, %d) mult: %f...", minGs, maxGs, loI, hiI, mult)
		// info: fmt.Printf("histCum: %+v\n", histCum.str())
		_ = flush.Flush()

		che := make(chan error)
		nThreads := 0
		ctxa := []jpegbw.FparCtx{}
		ctxInUse := make(map[int]bool)
		for i := 0; i < thrN; i++ {
			ctxa = append(ctxa, fctx.Cpy())
			ctxInUse[i] = false
		}
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
				for j := 0; j < y; j++ {
					fj := float64(j) / float64(y)
					pr, pg, pb, _ := m.At(i, j).RGBA()
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
					if bFun {
						var e error
						fv, e = ctxa[cNum].FparF([]float64{fv / 65535.0, fi, fj, fk})
						if e != nil {
							// Sync
							cmtx.Lock()
							ctxInUse[cNum] = false
							cmtx.Unlock()
							c <- e
							return
						}
						fv *= 65535.0
						if fv < 0.0 {
							fv = 0.0
						}
						if fv > 65535.0 {
							fv = 65535.0
						}
					}
					gs = uint16(fv)
					pixel := color.Gray16{gs}
					target.Set(i, j, pixel)
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
		pps := (all / dtEndF.Sub(dtStartF).Seconds()) / 1048576.0

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
		ary[lAry-1] = "bw_" + last
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
			ofn, dtEnd.Sub(dtStart), dtEndI.Sub(dtStartI), dtEndH.Sub(dtStartH), dtEndF.Sub(dtStartF), dtEnd.Sub(dtStartO), pps,
		)
	}
	return nil
}

func main() {
	dtStart := time.Now()
	if len(os.Args) > 1 {
		err := images2BW(os.Args[1:])
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		}
	} else {
		fmt.Printf("Please provide at least one image to convert\n")
		helpStr := `
Environment variables:
Q - jpeg quality 1-100, will use library default if not specified
R - relative red usage for generating gray pixel, 1 if not specified
G - relative green usage for generating gray pixel, 1 if not specified
B - relative blue usage for generating gray pixel, 1 if not specified
(R+G+B) will be normalized to sum to 1, so their sum must be positive
R=0.2125 G=0.7154 B=0.0721 is a suggested configuration
LO - when calculating intensity range, discard values than are in this lower %, for example 3
HI - when calculating intensity range, discard values that are in this higher %, for example 3
GA - gamma default 1, which uses straight line (0,0) -> (1,1), if set uses (x,y)->(x,pow(x, GA)) mapping
F - function to apply on final 0-1 range, for example "sin(x1*2)+cos(x1*3)"
LIB - if F is used and F calls external functions, thery need to be loaded for this C library
N - set number of CPUs to process data
O - eventual overwite file name config, example: ".jpg:.png"
`
		fmt.Printf("%s\n", helpStr)
	}
	dtEnd := time.Now()
	fmt.Printf("Time: %v\n", dtEnd.Sub(dtStart))
}
