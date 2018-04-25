package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"jpegbw"
	"math"
	"math/cmplx"
	"os"
	"runtime"
	"strconv"
	"sync"
	"time"
)

type scanline struct {
	idx  int
	line []complex128
	minr float64
	mini float64
	minm float64
	maxr float64
	maxi float64
	maxm float64
	err  error
}

type complexRect [][]complex128

type hitInfo struct {
	hitsR []color.RGBA
	hitsI []color.RGBA
	hitsM []color.RGBA
}

type pixRect [][]hitInfo

func firstColor(ra, ia, ma []color.RGBA) color.RGBA {
	for _, col := range ra {
		if col.A == 0 {
			continue
		}
		return col
	}
	for _, col := range ia {
		if col.A == 0 {
			continue
		}
		return col
	}
	for _, col := range ma {
		if col.A == 0 {
			continue
		}
		return col
	}
	return color.RGBA{uint8(0), uint8(0), uint8(0), uint8(0xff)}
}

func mergeColors(ra, ia, ma []color.RGBA) (uint8, uint8, uint8, uint8) {
	r, g, b, a, n := 0, 0, 0, 0, 0
	for _, col := range ra {
		if col.A == 0 {
			continue
		}
		r += int(col.R)
		g += int(col.G)
		b += int(col.B)
		a += int(col.A)
		n++
	}
	for _, col := range ia {
		if col.A == 0 {
			continue
		}
		r += int(col.R)
		g += int(col.G)
		b += int(col.B)
		a += int(col.A)
		n++
	}
	for _, col := range ma {
		if col.A == 0 {
			continue
		}
		r += int(col.R)
		g += int(col.G)
		b += int(col.B)
		a += int(col.A)
		n++
	}
	if n == 0 {
		return uint8(0), uint8(0), uint8(0), uint8(0xff)
	}
	if n > 1 {
		r /= n
		g /= n
		b /= n
		a /= n
		// debug: fmt.Printf("Merged from %d colors: (%v,%v,%v,%v)\n", n, r, g, b, a)
	}
	return uint8(r), uint8(g), uint8(b), uint8(a)
}

func (cr complexRect) str() string {
	xl := len(cr)
	s := ""
	s += fmt.Sprintf("X length: %5d\n", xl)
	for i := 0; i < xl; i++ {
		yl := len(cr[i])
		s += fmt.Sprintf("Y[%5d] length: %d: [", i, yl)
		for j := 0; j < yl; j++ {
			s += fmt.Sprintf("[%5d,%5d]=%8.3f+%8.3fi ", i, j, real(cr[i][j]), imag(cr[i][j]))
		}
		s += "\n"
	}
	return s
}

func makePixData(x, y int) pixRect {
	var matrix pixRect
	for i := 0; i < x; i++ {
		row := []hitInfo{}
		for j := 0; j < y; j++ {
			row = append(row, hitInfo{})
		}
		matrix = append(matrix, row)
	}
	return matrix
}

func (p pixRect) str(x, y int) string {
	s := ""
	for i := 0; i < x; i++ {
		for j := 1; j < y; j++ {
			h := false
			lR := len(p[i][j].hitsR)
			lI := len(p[i][j].hitsI)
			lM := len(p[i][j].hitsM)
			if lR > 0 {
				s += fmt.Sprintf("Re[%d,%d] ", i, j)
				h = true
			}
			if lI > 0 {
				s += fmt.Sprintf("Im[%d,%d] ", i, j)
				h = true
			}
			if lM > 0 {
				s += fmt.Sprintf("Mo[%d,%d] ", i, j)
				h = true
			}
			if h {
				s += "\n"
			}
		}
	}
	return s
}

func calculateHits(px pixRect, data complexRect, x, y int, val float64, re, im, mo color.RGBA) {
	// info: fmt.Printf("calculateHits: %f (%v, %v, %v)\n", val, re, im, mo)
	for i := 0; i < x; i++ {
		for j := 1; j < y; j++ {
			pv := data[i][j-1]
			v := data[i][j]
			pvr := real(pv)
			vr := real(v)
			pvi := imag(pv)
			vi := imag(v)
			pvm := cmplx.Abs(pv)
			vm := cmplx.Abs(v)
			if (pvr <= val && vr > val) || (pvr >= val && vr < val) {
				px[i][j].hitsR = append(px[i][j].hitsR, re)
			}
			if (pvi <= val && vi > val) || (pvi >= val && vi < val) {
				px[i][j].hitsI = append(px[i][j].hitsI, im)
			}
			if (pvm <= val && vm > val) || (pvm >= val && vm < val) {
				px[i][j].hitsM = append(px[i][j].hitsM, mo)
			}
		}
	}
	for j := 0; j < y; j++ {
		for i := 1; i < x; i++ {
			pv := data[i-1][j]
			v := data[i][j]
			pvr := real(pv)
			vr := real(v)
			pvi := imag(pv)
			vi := imag(v)
			pvm := cmplx.Abs(pv)
			vm := cmplx.Abs(v)
			if (pvr <= val && vr > val) || (pvr >= val && vr < val) {
				px[i][j].hitsR = append(px[i][j].hitsR, re)
			}
			if (pvi <= val && vi > val) || (pvi >= val && vi < val) {
				px[i][j].hitsI = append(px[i][j].hitsI, im)
			}
			if (pvm <= val && vm > val) || (pvm >= val && vm < val) {
				px[i][j].hitsM = append(px[i][j].hitsM, mo)
			}
		}
	}
}

func cmap(ofn, f string) error {
	var fctx jpegbw.FparCtx

	// LIB, NF
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
		ok := fctx.Init(lib, uint(nf))
		if !ok {
			return fmt.Errorf("LIB init failed for: %s", lib)
		}
		defer func() { fctx.Tidy() }()
	}
	err := fctx.FparFunction(f)
	if err != nil {
		return err
	}
	err = fctx.FparOK(1)
	if err != nil {
		return err
	}

	// x, y resolution
	x := 1000
	y := 1000

	// X
	xs := os.Getenv("X")
	if xs != "" {
		v, err := strconv.Atoi(xs)
		if err != nil {
			return err
		}
		if v < 1 || v > 0xffff {
			return fmt.Errorf("X must be from 1-65535 range")
		}
		x = v
	} else {
		fmt.Printf("Default X resolution used: %d\n", x)
	}

	// Y
	ys := os.Getenv("Y")
	if ys != "" {
		v, err := strconv.Atoi(ys)
		if err != nil {
			return err
		}
		if v < 1 || v > 0xffff {
			return fmt.Errorf("Y must be from 1-65535 range")
		}
		y = v
	} else {
		fmt.Printf("Default Y resolution used: %d\n", y)
	}
	all := float64(x * y)

	// K
	kinc := 0x10
	xk := os.Getenv("K")
	if xk != "" {
		v, err := strconv.Atoi(xk)
		if err != nil {
			return err
		}
		if v < 1 || v > 0xff {
			return fmt.Errorf("K must be from 1-255 range")
		}
		kinc = v
	} else {
		fmt.Printf("Default K lines increment used resolution used: %d\n", kinc)
	}

	// R0
	r0 := -1.0
	r0s := os.Getenv("R0")
	if r0s != "" {
		v, err := strconv.ParseFloat(r0s, 64)
		if err != nil {
			return err
		}
		r0 = v
	} else {
		fmt.Printf("Default R0 used: %f\n", r0)
	}

	// R1
	r1 := 1.0
	r1s := os.Getenv("R1")
	if r1s != "" {
		v, err := strconv.ParseFloat(r1s, 64)
		if err != nil {
			return err
		}
		r1 = v
	} else {
		fmt.Printf("Default R1 used: %f\n", r1)
	}
	if r0 >= r1 {
		return fmt.Errorf("r0 must be less than r1: r0=%f r1=%f", r0, r1)
	}
	dr := r1 - r0

	// I0
	i0 := -1.0
	i0s := os.Getenv("I0")
	if i0s != "" {
		v, err := strconv.ParseFloat(i0s, 64)
		if err != nil {
			return err
		}
		i0 = v
	} else {
		fmt.Printf("Default I0 used: %f\n", i0)
	}

	// R1
	i1 := 1.0
	i1s := os.Getenv("I1")
	if i1s != "" {
		v, err := strconv.ParseFloat(i1s, 64)
		if err != nil {
			return err
		}
		i1 = v
	} else {
		fmt.Printf("Default I1 used: %f\n", i1)
	}
	if i0 >= i1 {
		return fmt.Errorf("i0 must be less than i1: i0=%f i1=%f", i0, i1)
	}
	di := i1 - i0

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

	fmt.Printf("(%d x %d) Real: [%f,%f] Imag: [%f,%f] Threads: %d\n", x, y, r0, r1, i0, i1, thrN)

	// Run
	nThreads := 0
	ch := make(chan scanline)
	dtStart := time.Now()

	// Need thrN contexts
	var cmtx = &sync.Mutex{}
	ctxa := []jpegbw.FparCtx{}
	ctxInUse := make(map[int]bool)
	for i := 0; i < thrN; i++ {
		ctxa = append(ctxa, fctx.Cpy())
		ctxInUse[i] = false
	}

	// Output array
	var (
		data         complexRect
		complexPlane complexRect
	)
	for i := 0; i < x; i++ {
		cr := r0 + (float64(i)/float64(x-1))*dr
		row := []complex128{}
		data = append(data, []complex128{})
		for j := 0; j < y; j++ {
			ci := i0 + (float64(j)/float64(y-1))*di
			z := complex(cr, ci)
			row = append(row, z)
		}
		complexPlane = append(complexPlane, row)
	}
	minr := math.MaxFloat64
	mini := math.MaxFloat64
	minm := math.MaxFloat64
	maxr := -math.MaxFloat64
	maxi := -math.MaxFloat64
	maxm := -math.MaxFloat64
	for ii := 0; ii < x; ii++ {
		go func(ch chan scanline, i int) {
			var line []complex128
			minr := math.MaxFloat64
			mini := math.MaxFloat64
			minm := math.MaxFloat64
			maxr := -math.MaxFloat64
			maxi := -math.MaxFloat64
			maxm := -math.MaxFloat64
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
				ch <- scanline{err: fmt.Errorf("no context copy available: i=%d", i)}
				return
			}
			cr := r0 + (float64(i)/float64(x-1))*dr
			for j := 0; j < y; j++ {
				ci := i0 + (float64(j)/float64(y-1))*di
				z := complex(cr, ci)
				fz, e := ctxa[cNum].FparF([]complex128{z})
				// debug: fmt.Printf("'%s'[%d,%d](%v) = %v\n", f, i, j, z, fz)
				line = append(line, fz)
				if e != nil {
					cmtx.Lock()
					ctxInUse[cNum] = false
					cmtx.Unlock()
					ch <- scanline{err: e}
					return
				}
				fzr := real(fz)
				fzi := imag(fz)
				fzm := cmplx.Abs(fz)
				if fzr > maxr {
					maxr = fzr
				}
				if fzi > maxi {
					maxi = fzi
				}
				if fzm > maxm {
					maxm = fzm
				}
				if fzr < minr {
					minr = fzr
				}
				if fzi < mini {
					mini = fzi
				}
				if fzm < minm {
					minm = fzm
				}
			}
			cmtx.Lock()
			ctxInUse[cNum] = false
			cmtx.Unlock()
			ch <- scanline{idx: i, line: line, minr: minr, mini: mini, minm: minm, maxr: maxr, maxi: maxi, maxm: maxm, err: nil}
		}(ch, ii)

		nThreads++
		if nThreads == thrN {
			line := <-ch
			if line.err != nil {
				return line.err
			}
			data[line.idx] = line.line
			if line.maxr > maxr {
				maxr = line.maxr
			}
			if line.maxi > maxi {
				maxi = line.maxi
			}
			if line.maxm > maxm {
				maxm = line.maxm
			}
			if line.minr < minr {
				minr = line.minr
			}
			if line.mini < mini {
				mini = line.mini
			}
			if line.minm < minm {
				minm = line.minm
			}
			nThreads--
		}
	}
	for nThreads > 0 {
		line := <-ch
		if line.err != nil {
			return line.err
		}
		data[line.idx] = line.line
		if line.err != nil {
			return line.err
		}
		data[line.idx] = line.line
		if line.maxr > maxr {
			maxr = line.maxr
		}
		if line.maxi > maxi {
			maxi = line.maxi
		}
		if line.maxm > maxm {
			maxm = line.maxm
		}
		if line.minr < minr {
			minr = line.minr
		}
		if line.mini < mini {
			mini = line.mini
		}
		if line.minm < minm {
			minm = line.minm
		}
		nThreads--
	}

	// Info
	// debug: fmt.Printf("Matrix\n%s\n", data.str())

	// Prepare structure to hold hits info
	px := makePixData(x, y)

	// Minumum
	mi := math.Min(minr, mini)
	mi = math.Min(mi, minm)
	// Maximum
	ma := math.Min(maxr, maxi)
	ma = math.Min(ma, maxm)

	// Info
	dm := (ma - mi) / 255.0
	fmt.Printf("Values range: %v - %v, modulo range: %f - %f, lines range: %f - %f\n", complex(minr, mini), complex(maxr, maxi), minm, maxm, mi, ma)

	// Calculate hits
	last := false
	kp := 0
	kpb := false
	for k := 0; k < 0x100; k += kinc {
		v := mi + float64(k)*dm
		if v > 0.0 && !kpb {
			kp = k
			kpb = true
		}
	}
	// info: fmt.Printf("First positive modulo k: %d\n", kp)
	for k := 0; k < 0x100; k += kinc {
		v := mi + float64(k)*dm
		calculateHits(
			px, data, x, y, v,
			color.RGBA{uint8(0xff - k), uint8(k), uint8(k), 0xff},
			color.RGBA{uint8(k), uint8(k), uint8(0xff - k), 0xff},
			color.RGBA{uint8(k), uint8(0xff - (k + kp)), uint8(k), 0xff},
		)
		if k == 0xff {
			last = true
		}
	}
	if !last {
		// Max must be shown
		calculateHits(
			px, data, x, y, ma,
			color.RGBA{uint8(0), uint8(0xff), uint8(0xff), 0xff},
			color.RGBA{uint8(0xff), uint8(0xff), uint8(0), 0xff},
			color.RGBA{uint8(0xff), uint8(0), uint8(0xff), 0xff},
		)
	}
	// Function 0's Re, IM, Modulo
	// Re = 0 red almost white
	// Im = 0 blue almost white
	// Mod = 0 green almost white (it means complex zero, function retuned (0+0i)
	calculateHits(
		px, data, x, y, 0.0,
		color.RGBA{uint8(0xff), uint8(0xc0), uint8(0xce), 0xff},
		color.RGBA{uint8(0xce), uint8(0xce), uint8(0xff), 0xff},
		color.RGBA{uint8(0xce), uint8(0xff), uint8(0xce), 0xff},
	)
	// Complex plane axes and unit circle
	// Re = 0 and Im = 0 white
	calculateHits(
		px, complexPlane, x, y, 0.0,
		color.RGBA{uint8(0xff), uint8(0xff), uint8(0xff), 0xff},
		color.RGBA{uint8(0xff), uint8(0xff), uint8(0xff), 0xff},
		color.RGBA{uint8(0x0), uint8(0x0), uint8(0x0), 0}, // Complex Modulo 0 is just a point at (0,0)
	)
	// Modulo unit circle white
	calculateHits(
		px, complexPlane, x, y, 1.0,
		color.RGBA{uint8(0x0), uint8(0x0), uint8(0x0), 0},
		color.RGBA{uint8(0x0), uint8(0x0), uint8(0x0), 0},
		color.RGBA{uint8(0xff), uint8(0xff), uint8(0xff), 0xff},
	)
	// debug: fmt.Printf("Hits\n%s\n", px.str(x, y))

	// Output
	target := image.NewRGBA(image.Rect(0, 0, x, y))
	for i := 0; i < x; i++ {
		for j := 0; j < y; j++ {
			r, g, b, a := mergeColors(px[i][j].hitsR, px[i][j].hitsI, px[i][j].hitsM)
			pixel := color.RGBA{r, g, b, a}
			//pixel := firstColor(px[i][j].hitsR, px[i][j].hitsI, px[i][j].hitsM)
			target.Set(i, (y-j)-1, pixel)
		}
	}
	fout, err := os.Create(ofn)
	if err != nil {
		return err
	}
	defer func() { _ = fout.Close() }()
	err = png.Encode(fout, target)
	if err != nil {
		return err
	}

	dtEnd := time.Now()
	pps := (all / dtEnd.Sub(dtStart).Seconds()) / 1048576.0
	fmt.Printf("Processed in: %v, MPPS: %.3f, %d\n", dtEnd.Sub(dtStart), pps, nThreads)
	fmt.Printf("Real values from minimum to max are: red --> cyan/teal\n")
	fmt.Printf("Imag values from minimum to max are: blue --> yellow\n")
	fmt.Printf("Modulo values from minimum to max are: green --> pink\n")
	fmt.Printf("Re = 0 red almost white\n")
	fmt.Printf("Im = 0 blue almost white\n")
	fmt.Printf("Mod = 0 green almost white\n")
	fmt.Printf("Complex plane Re = 0, Im = 0 and modulo unit circle: white\n")
	return nil
}

func main() {
	dtStart := time.Now()
	if len(os.Args) >= 3 {
		err := cmap(os.Args[1], os.Args[2])
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		}
	} else {
		helpStr := `
Parameters required: output_file_name.png 'function definition'
Example: LIB="/usr/local/lib/libjpegbw.so" out.png 'csin(x1)'

Environment variables:
LIB - if F is used and F calls external functions, thery need to be loaded for this C library
NF - set maximum number of distinct functions in the parser, if not set, default 128 is used
N - set number of CPUs to process data
X - x resoultion - output image width
Y - y resoultion - output image width
R0 - Real from
R1 - Real to
I0 - Imag from
I1 - Imag to
K - increment value to next line: 0-255, default 16
`
		fmt.Printf("%s\n", helpStr)
	}
	dtEnd := time.Now()
	fmt.Printf("Time: %v\n", dtEnd.Sub(dtStart))
}
