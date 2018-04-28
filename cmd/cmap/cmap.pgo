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
	"math/cmplx"
	"os"
	"runtime"
	"runtime/pprof"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/andybons/gogif"
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
	hits  []color.RGBA
	types []int // 0 - contour, -1 - value below, 1 - value above, -2 - none of the above
}

type pixRect [][]hitInfo

type drawConfigItem struct {
	fz    bool       // false: draw complex plane data (z), true draw function data f(z)
	rim   string     // can be "r" - real, "i" - imag, "m" - modulo
	v     float64    // value to draw
	col   color.RGBA // color to use
	nextv string     // value increment (if many frames) - this is a FparF function definition it receives "v" and 0-1 fraction (for frames 1-n)
	cinc  []float64  // color increment
	lh    bool       // false - contours only, true - draw also lo/hi values with blended color
}

type drawConfig struct {
	items []drawConfigItem
	n     int
}

func (dc *drawConfig) initFromEnv() (bool, error) {
	var fctx jpegbw.FparCtx
	s := os.Getenv("U")
	if s == "" {
		return false, nil
	}
	ary := strings.Split(strings.TrimSpace(s), "|")
	if len(ary) < 2 {
		return false, fmt.Errorf("required at least two elements separated by '|': %s", s)
	}
	n, err := strconv.Atoi(strings.TrimSpace(ary[0]))
	if err != nil {
		return false, err
	}
	dc.n = n
	for idx, item := range ary[1:] {
		item := strings.TrimSpace(item)
		//fz;r;3.14;255:128:192:255;0.01;0.01:-0.01:0:0;0
		ary := strings.Split(item, ";")
		if len(ary) != 7 {
			return false, fmt.Errorf("single item must have 6 ',' values: fz;r;v;col;nextv(x1,x2);cinc;lh: '%s', got %d for %d item", item, len(ary), idx+1)
		}
		itemAry := []string{}
		for _, el := range ary {
			itemAry = append(itemAry, strings.TrimSpace(el))
		}
		var dci drawConfigItem
		if itemAry[0] == "fz" {
			dci.fz = true
		} else if itemAry[0] == "z" {
			dci.fz = false
		} else {
			return false, fmt.Errorf("item %d: '%s' fz value incorrect: '%s' must be 'z' or 'fz'", idx+1, item, itemAry[0])
		}
		if itemAry[1] == "r" || itemAry[1] == "i" || itemAry[1] == "m" {
			dci.rim = itemAry[1]
		} else {
			return false, fmt.Errorf("item %d: '%s' rim value incorrect: '%s' must be 'r', 'i' or 'm'", idx+1, item, itemAry[1])
		}
		v, err := strconv.ParseFloat(itemAry[2], 64)
		if err != nil {
			return false, err
		}
		dci.v = v
		colA := strings.Split(itemAry[3], ":")
		if len(colA) != 4 {
			return false, fmt.Errorf("item %d: '%s' col value incorrect: '%s' must be 4 0-255 uint8 values ':' separated", idx+1, item, itemAry[3])
		}
		r, err := strconv.Atoi(strings.TrimSpace(colA[0]))
		if err != nil {
			return false, err
		}
		g, err := strconv.Atoi(strings.TrimSpace(colA[1]))
		if err != nil {
			return false, err
		}
		b, err := strconv.Atoi(strings.TrimSpace(colA[2]))
		if err != nil {
			return false, err
		}
		a, err := strconv.Atoi(strings.TrimSpace(colA[3]))
		if err != nil {
			return false, err
		}
		if r < 0 || r > 0xff || g < 0 || g > 0xff || b < 0 || b > 0xff || a < 0 || a > 0xff {
			return false, fmt.Errorf("item %d: '%s' col value incorrect: '%s' all r,g,b,g values must be from 0-255 range", idx+1, item, itemAry[3])
		}
		dci.col = color.RGBA{uint8(r), uint8(g), uint8(b), uint8(a)}
		fdef := itemAry[4]
		err = fctx.FparFunction(fdef)
		if err != nil {
			return false, err
		}
		err = fctx.FparOK(2)
		if err != nil {
			return false, err
		}
		dci.nextv = fdef
		colA = strings.Split(itemAry[5], ":")
		if len(colA) != 4 {
			return false, fmt.Errorf("item %d: '%s' colInc value incorrect: '%s' must be 4 float values ':' separated", idx+1, item, itemAry[5])
		}
		ri, err := strconv.ParseFloat(strings.TrimSpace(colA[0]), 64)
		if err != nil {
			return false, err
		}
		gi, err := strconv.ParseFloat(strings.TrimSpace(colA[1]), 64)
		if err != nil {
			return false, err
		}
		bi, err := strconv.ParseFloat(strings.TrimSpace(colA[2]), 64)
		if err != nil {
			return false, err
		}
		ai, err := strconv.ParseFloat(strings.TrimSpace(colA[3]), 64)
		if err != nil {
			return false, err
		}
		dci.cinc = []float64{ri, gi, bi, ai}
		if itemAry[6] == "1" {
			dci.lh = true
		} else if itemAry[6] == "0" {
			dci.lh = false
		} else {
			return false, fmt.Errorf("item %d: '%s' lh value incorrect: '%s' must be '1' or '0'", idx+1, item, itemAry[6])
		}
		dc.items = append(dc.items, dci)
	}
	return true, nil
}

func firstColor(ha []color.RGBA, ty []int) color.RGBA {
	for i, col := range ha {
		if ty[i] != 0 {
			continue
		}
		return col
	}
	for i, col := range ha {
		if ty[i] == -2 {
			continue
		}
		return col
	}
	for _, col := range ha {
		return col
	}
	return color.RGBA{uint8(0xff), uint8(0xff), uint8(0xff), uint8(0xff)}
}

func mergeColors(ha []color.RGBA, ty []int) (uint8, uint8, uint8, uint8) {
	r, g, b, a, n := 0, 0, 0, 0, 0
	for i, col := range ha {
		if ty[i] != 0 {
			continue
		}
		r += int(col.R)
		g += int(col.G)
		b += int(col.B)
		a += int(col.A)
		n++
	}
	if n == 0 {
		for i, col := range ha {
			if ty[i] == -2 {
				continue
			}
			r += int(col.R)
			g += int(col.G)
			b += int(col.B)
			a += int(col.A)
			n++
		}
		if n == 0 {
			for _, col := range ha {
				r += int(col.R)
				g += int(col.G)
				b += int(col.B)
				a += int(col.A)
				n++
			}
		}
	}
	if n == 0 {
		return uint8(0xff), uint8(0xff), uint8(0xff), uint8(0xff)
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
			s += fmt.Sprintf("[%5d,%5d]=%8.3f+%8.3fi(%8.3f) ", i, j, real(cr[i][j]), imag(cr[i][j]), cmplx.Abs(cr[i][j]))
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
			l := len(p[i][j].hits)
			if l > 0 {
				s += fmt.Sprintf("hit[%d,%d]: ", i, j)
				for _, hit := range p[i][j].hits {
					s += fmt.Sprintf("%v ", hit)
				}
				s += "\n"
			}
		}
	}
	return s
}

func colorLoHi(c color.RGBA) (color.RGBA, color.RGBA) {
	lo := color.RGBA{
		uint8(0xff - (c.R >> 1)),
		uint8(0xff - (c.G >> 1)),
		uint8(0xff - (c.B >> 1)),
		uint8(0xff),
	}
	hi := color.RGBA{
		uint8(0xff - ((c.G + c.B) >> 2)),
		uint8(0xff - ((c.R + c.B) >> 2)),
		uint8(0xff - ((c.R + c.G) >> 2)),
		uint8(0xff),
	}
	// debug: fmt.Printf("%v --> (%v, %v)\n", c, lo, hi)
	return lo, hi
}

func calculateHits(px pixRect, data complexRect, thrN int, lh bool, x, y int, val float64, rib string, colC, colL, colH, colU color.RGBA) {
	// info: fmt.Printf("calculateHits: lh=%v val=%f, rib=%s, C=%v, L=%v, H=%v, U=%v\n", lh, val, rib, colC, colL, colH, colU)
	var sel func(complex128) float64
	if rib == "r" {
		sel = func(z complex128) float64 { return real(z) }
	} else if rib == "i" {
		sel = func(z complex128) float64 { return imag(z) }
	} else if rib == "m" {
		sel = cmplx.Abs
	} else {
		return
	}
	n := 0
	ch := make(chan struct{})

	if !lh {
		for ii := 0; ii < x; ii++ {
			go func(ch chan struct{}, i int) {
				for j := 1; j < y; j++ {
					pv := sel(data[i][j-1])
					v := sel(data[i][j])
					if (pv <= val && v > val) || (pv >= val && v < val) {
						px[i][j].hits = append(px[i][j].hits, colC)
						px[i][j].types = append(px[i][j].types, 0)
					}
				}
				ch <- struct{}{}
			}(ch, ii)
			n++
			if n == thrN {
				<-ch
				n--
			}
		}
		for n > 0 {
			<-ch
			n--
		}
		for jj := 0; jj < y; jj++ {
			go func(ch chan struct{}, j int) {
				for i := 1; i < x; i++ {
					pv := sel(data[i-1][j])
					v := sel(data[i][j])
					if (pv <= val && v > val) || (pv >= val && v < val) {
						px[i][j].hits = append(px[i][j].hits, colC)
						px[i][j].types = append(px[i][j].types, 0)
					}
				}
				ch <- struct{}{}
			}(ch, jj)
			n++
			if n == thrN {
				<-ch
				n--
			}
		}
		for n > 0 {
			<-ch
			n--
		}
		return
	}

	// Hits info
	m1 := make([]int, x*y)
	m2 := make([]int, x*y)

	for ii := 0; ii < x; ii++ {
		iiy := ii * y
		go func(ch chan struct{}, i, iy int) {
			for j := 1; j < y; j++ {
				arg := iy + j
				pv := sel(data[i][j-1])
				v := sel(data[i][j])
				if (pv <= val && v > val) || (pv >= val && v < val) {
					m1[arg] = 0
				} else if pv <= val && v <= val {
					m1[arg] = -1
				} else if pv >= val && v >= v {
					m1[arg] = 1
				} else {
					m1[arg] = -2
				}
			}
			ch <- struct{}{}
		}(ch, ii, iiy)
		n++
		if n == thrN {
			<-ch
			n--
		}
	}
	for n > 0 {
		<-ch
		n--
	}
	for jj := 0; jj < y; jj++ {
		go func(ch chan struct{}, j int) {
			for i := 1; i < x; i++ {
				arg := i*y + j
				pv := sel(data[i-1][j])
				v := sel(data[i][j])
				if (pv <= val && v > val) || (pv >= val && v < val) {
					m2[arg] = 0
				} else if pv <= val && v <= val {
					m2[arg] = -1
				} else if pv >= val && v >= v {
					m2[arg] = 1
				} else {
					m2[arg] = -2
				}
			}
			ch <- struct{}{}
		}(ch, jj)
		n++
		if n == thrN {
			<-ch
			n--
		}
	}
	for n > 0 {
		<-ch
		n--
	}
	ca := colC.A
	la := colL.A
	ha := colH.A
	ua := colU.A
	for ii := 0; ii < x; ii++ {
		iiy := ii * y
		go func(ch chan struct{}, i, iy int) {
			for j := 1; j < y; j++ {
				arg := iy + j
				v1 := m1[arg]
				v2 := m2[arg]
				if v1 == 0 || v2 == 0 {
					if ca != 0 {
						px[i][j].hits = append(px[i][j].hits, colC)
						px[i][j].types = append(px[i][j].types, 0)
					}
				} else if v1 == -1 && v2 == -1 {
					if la != 0 {
						px[i][j].hits = append(px[i][j].hits, colL)
						px[i][j].types = append(px[i][j].types, -1)
					}
				} else if v1 == 1 && v2 == 1 {
					if ha != 0 {
						px[i][j].hits = append(px[i][j].hits, colH)
						px[i][j].types = append(px[i][j].types, 1)
					}
				} else {
					if ua != 0 {
						px[i][j].hits = append(px[i][j].hits, colU)
						px[i][j].types = append(px[i][j].types, -2)
					}
				}
			}
			ch <- struct{}{}
		}(ch, ii, iiy)
		n++
		if n == thrN {
			<-ch
			n--
		}
	}
	for n > 0 {
		<-ch
		n--
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
	err = fctx.FparOK(2)
	if err != nil {
		return err
	}

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

	// Merge colors or use first hit's color?
	mergeCols := os.Getenv("FC") == ""

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

	// User defined draw config
	var dc drawConfig
	dcMode, err := dc.initFromEnv()
	if err != nil {
		return err
	}

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
	dmr := (maxr - minr) / 255.0
	dmi := (maxi - mini) / 255.0
	dmm := (maxm - minm) / 255.0
	// Info
	fmt.Printf("Values range: %v - %v, modulo range: %f - %f\n", complex(minr, mini), complex(maxr, maxi), minm, maxm)

	// Zero color
	cc := color.RGBA{uint8(255), uint8(255), uint8(255), uint8(0)}
	ccL := cc
	ccH := cc
	// Do we want Lo/Hi blended colors in addition to contour?
	lh := os.Getenv("LH") != ""

	// Flushing before endline
	flush := bufio.NewWriter(os.Stdout)

	if dcMode {
		// GIF and JPG frames
		saveGIF := os.Getenv("NOGIF") == ""
		saveFrames := os.Getenv("JPG") != ""
		if !saveGIF && !saveFrames {
			return fmt.Errorf("you need to save GIF or separate frames as JPEGs")
		}

		lfn := strings.ToLower(ofn)
		if saveGIF && !strings.Contains(lfn, ".gif") {
			return fmt.Errorf("only .gif files can be used for user mode video-like output: %s", ofn)
		}
		var images []*image.Paletted
		var delays []int
		fmt.Printf("%d frames\n", dc.n)
		for f := 0; f < dc.n; f++ {
			ff := 0.0
			if dc.n > 1 {
				ff = float64(f) / float64(dc.n-1)
			}
			fmt.Printf("%d ", f+1)
			_ = flush.Flush()
			// Prepare structure to hold hits info
			px := makePixData(x, y)
			for _, item := range dc.items {
				var fc jpegbw.FparCtx
				r := float64(item.col.R) + float64(f)*item.cinc[0]
				g := float64(item.col.G) + float64(f)*item.cinc[1]
				b := float64(item.col.B) + float64(f)*item.cinc[2]
				a := float64(item.col.A) + float64(f)*item.cinc[3]
				if r < 0.0 {
					r = 0.0
				}
				if g < 0.0 {
					g = 0.0
				}
				if b < 0.0 {
					b = 0.0
				}
				if a < 0.0 {
					a = 0.0
				}
				if r > 255.0 {
					r = 255.0
				}
				if g > 255.0 {
					g = 255.0
				}
				if b > 255.0 {
					b = 255.0
				}
				if a > 255.0 {
					a = 255.0
				}
				c := color.RGBA{uint8(r), uint8(g), uint8(b), uint8(a)}
				if item.lh {
					ccL, ccH = colorLoHi(c)
				}
				err := fc.FparFunction(item.nextv)
				if err != nil {
					return err
				}
				err = fc.FparOK(2)
				if err != nil {
					return err
				}
				fz, err := fc.FparF([]complex128{complex(item.v, 0.0), complex(ff, 0.0)})
				if err != nil {
					return err
				}
				v := real(fz)
				if item.fz {
					calculateHits(px, data, thrN, item.lh, x, y, v, item.rim, c, ccL, ccH, cc)
				} else {
					calculateHits(px, complexPlane, thrN, item.lh, x, y, v, item.rim, c, ccL, ccH, cc)
				}
			}
			target := image.NewRGBA(image.Rect(0, 0, x, y))
			if mergeCols {
				for i := 0; i < x; i++ {
					for j := 0; j < y; j++ {
						r, g, b, a := mergeColors(px[i][j].hits, px[i][j].types)
						pixel := color.RGBA{r, g, b, a}
						target.Set(i, (y-j)-1, pixel)
					}
				}
			} else {
				for i := 0; i < x; i++ {
					for j := 0; j < y; j++ {
						target.Set(i, (y-j)-1, firstColor(px[i][j].hits, px[i][j].types))
					}
				}
			}
			// save single frame
			if saveFrames {
				f, err := os.Create(fmt.Sprintf("frame%05d.jpg", f))
				if err != nil {
					return err
				}
				if jpegq < 0 {
					err = jpeg.Encode(f, target, nil)
				} else {
					err = jpeg.Encode(f, target, &jpeg.Options{Quality: jpegq})
				}
				_ = f.Close()
				if err != nil {
					return err
				}
			}

			if saveGIF {
				// Add GIF frame
				bounds := target.Bounds()
				palettedImage := image.NewPaletted(bounds, nil)
				quantizer := gogif.MedianCutQuantizer{NumColor: 0x10000}
				quantizer.Quantize(palettedImage, bounds, target, image.ZP)
				images = append(images, palettedImage)
				delays = append(delays, 0)
			}
		}
		if saveGIF {
			fout, err := os.Create(ofn)
			if err != nil {
				return err
			}
			defer func() { _ = fout.Close() }()
			err = gif.EncodeAll(fout, &gif.GIF{Image: images, Delay: delays})
			if err != nil {
				return err
			}
		}
		return nil
	}

	// Prepare structure to hold hits info
	px := makePixData(x, y)

	// Calculate hits
	// Real hits
	last := false
	for k := 0; k < 0x100; k += kinc {
		v := minr + float64(k)*dmr
		c := color.RGBA{uint8(0xff - k), uint8(k), uint8(k), 0xff}
		if lh {
			ccL, ccH = colorLoHi(c)
		}
		calculateHits(px, data, thrN, lh, x, y, v, "r", c, ccL, ccH, cc)
		if k == 0xff {
			last = true
		}
	}
	if !last {
		// Max must be shown
		c := color.RGBA{uint8(0), uint8(0xff), uint8(0xff), 0xff}
		if lh {
			ccL, ccH = colorLoHi(c)
		}
		calculateHits(px, data, thrN, lh, x, y, maxr, "r", c, ccL, ccH, cc)
	}

	// Imag hits
	last = false
	for k := 0; k < 0x100; k += kinc {
		v := mini + float64(k)*dmi
		c := color.RGBA{uint8(k), uint8(k), uint8(0xff - k), 0xff}
		if lh {
			ccL, ccH = colorLoHi(c)
		}
		calculateHits(px, data, thrN, lh, x, y, v, "i", c, ccL, ccH, cc)
		if k == 0xff {
			last = true
		}
	}
	if !last {
		// Max must be shown
		c := color.RGBA{uint8(0xff), uint8(0xff), uint8(0), 0xff}
		if lh {
			ccL, ccH = colorLoHi(c)
		}
		calculateHits(px, data, thrN, lh, x, y, maxi, "i", c, ccL, ccH, cc)
	}

	// Modulo/Abs hits
	last = false
	for k := 0; k < 0x100; k += kinc {
		v := minm + float64(k)*dmm
		c := color.RGBA{uint8(k), uint8(0xff - k), uint8(k), 0xff}
		if lh {
			ccL, ccH = colorLoHi(c)
		}
		calculateHits(px, data, thrN, lh, x, y, v, "m", c, ccL, ccH, cc)
		if k == 0xff {
			last = true
		}
	}
	if !last {
		// Max must be shown
		c := color.RGBA{uint8(0xff), uint8(0), uint8(0xff), 0xff}
		if lh {
			ccL, ccH = colorLoHi(c)
		}
		calculateHits(px, data, thrN, lh, x, y, maxm, "m", c, ccL, ccH, cc)
	}

	// Function 0's Re, IM, Modulo
	// Re = 0 dark red
	// Im = 0 dark blue
	// Mod = 0 dark green (it means complex zero, function retuned (0+0i)
	c := color.RGBA{uint8(0x80), uint8(0), uint8(0), 0xff}
	if lh {
		ccL, ccH = colorLoHi(c)
	}
	calculateHits(px, data, thrN, lh, x, y, 0.0, "r", c, ccL, ccH, cc)
	c = color.RGBA{uint8(0), uint8(0), uint8(0x80), 0xff}
	if lh {
		ccL, ccH = colorLoHi(c)
	}
	calculateHits(px, data, thrN, lh, x, y, 0.0, "i", c, ccL, ccH, cc)
	c = color.RGBA{uint8(0), uint8(0x80), uint8(0), 0xff}
	if lh {
		ccL, ccH = colorLoHi(c)
	}
	calculateHits(px, data, thrN, lh, x, y, 0.0, "m", c, ccL, ccH, cc)

	// Complex plane axes and unit circle
	// Re = 0 and Im = 0 black
	c = color.RGBA{uint8(0), uint8(0), uint8(0), 0xff}
	if lh {
		ccL, ccH = colorLoHi(c)
	}
	calculateHits(px, complexPlane, thrN, lh, x, y, 0.0, "r", c, ccL, ccH, cc)
	c = color.RGBA{uint8(0), uint8(0), uint8(0), 0xff}
	if lh {
		ccL, ccH = colorLoHi(c)
	}
	calculateHits(px, complexPlane, thrN, lh, x, y, 0.0, "i", c, ccL, ccH, cc)
	// Modulo unit circle white
	c = color.RGBA{uint8(0), uint8(0), uint8(0), 0xff}
	if lh {
		ccL, ccH = colorLoHi(c)
	}
	calculateHits(px, complexPlane, thrN, lh, x, y, 1.0, "m", c, ccL, ccH, cc)

	// debug: fmt.Printf("Hits\n%s\n", px.str(x, y))

	// Output
	target := image.NewRGBA(image.Rect(0, 0, x, y))
	if mergeCols {
		for i := 0; i < x; i++ {
			for j := 0; j < y; j++ {
				r, g, b, a := mergeColors(px[i][j].hits, px[i][j].types)
				pixel := color.RGBA{r, g, b, a}
				target.Set(i, (y-j)-1, pixel)
			}
		}
	} else {
		for i := 0; i < x; i++ {
			for j := 0; j < y; j++ {
				target.Set(i, (y-j)-1, firstColor(px[i][j].hits, px[i][j].types))
			}
		}
	}
	fout, err := os.Create(ofn)
	if err != nil {
		return err
	}
	defer func() { _ = fout.Close() }()
	var ierr error
	lfn := strings.ToLower(ofn)
	if strings.Contains(lfn, ".png") {
		ierr = png.Encode(fout, target)
	} else if strings.Contains(lfn, ".jpg") || strings.Contains(lfn, ".jpeg") {
		if jpegq < 0 {
			ierr = jpeg.Encode(fout, target, nil)
		} else {
			ierr = jpeg.Encode(fout, target, &jpeg.Options{Quality: jpegq})
		}
	} else if strings.Contains(lfn, ".gif") {
		ierr = gif.Encode(fout, target, nil)
	}
	if ierr != nil {
		return ierr
	}

	dtEnd := time.Now()
	pps := (all / dtEnd.Sub(dtStart).Seconds()) / 1048576.0
	fmt.Printf("Processed in: %v, MPPS: %.3f, %d\n", dtEnd.Sub(dtStart), pps, nThreads)
	fmt.Printf("Real values from minimum to max are: red --> cyan/teal\n")
	fmt.Printf("Imag values from minimum to max are: blue --> yellow\n")
	fmt.Printf("Modulo values from minimum to max are: green --> pink\n")
	fmt.Printf("Re = 0 dark red\n")
	fmt.Printf("Im = 0 dark blue\n")
	fmt.Printf("Mod = 0 dark green\n")
	fmt.Printf("Complex plane Re = 0, Im = 0 and modulo unit circle: black\n")
	return nil
}

func main() {
	dtStart := time.Now()
	if len(os.Args) >= 3 {
		prof := os.Getenv("PR")
		if prof != "" {
			f, err := os.Create(prof)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}
			_ = pprof.StartCPUProfile(f)
			defer pprof.StopCPUProfile()
		}
		err := cmap(os.Args[1], os.Args[2])
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		}
	} else {
		helpStr := `
Parameters required: output_file_name.png 'function definition'
Example: LIB="/usr/local/lib/libjpegbw.so" out.png 'csin(x1)'
PNG, JPG and GIF outputs are supported

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
FC - use first hit color instead of merging color from all hits
Q - image quality 1-100
U - define own contours to display, possibly with movement 
LH - draw lo/hi values (blended color of coutour chart - slows down a lot)
PR - dump CPU profile to a given file
--- for user defined contours
NOGIF - skip final animation GIF
JPG - save each frame in JPG file framexxxxx.jpg, xxxxx = frame number

User defined contours:
Provide U="n_frames|def1|def2|def3|...|defK"
n_frames - how many GIF animation frames and/or JPEG frames generate
def1..K - K definitions of countours (has nothing in commont with n_frames)
each definitions is:
"fz;rim;v;col;nextv(x1,x2);cinc;lh"
"fz;rim;v;rC:rG:rb:cA;nextv(x1,x2);ciR:ciB:ciG:ciA;lh"
where:
fz can be:
  z - check complex plane (function complex arg) value to match "v"
  fz - check function complex value to match "v"
rim can be:
  r - check if real part of "z" or "fz" (defined above) match "v"
  i - check if imaginary part of "z" or "fz" (defined above) match "v"
  m - check if complex modulo/abs of "z" or "fz" (defined above) match "v"
v - value to draw its countour, for example re(f(z)) = v "fz,r,v" or im(z) = v "z,i,v" etc.
col - if match then use this col as a color, defined as "r:b:b:a"
  r - red part of color, range 0-255
  g - green part of color, range 0-255
  b - blue part of color, range 0-255
  a - alpha part of color, range 0-255
nextv(x1,x2) - function to get "v" value in next frames, to increement from v to v+2 on all frames on each frame use "x1+x2*2"
  x1 receives start "v"
  x2 is changing from 0 to 1 for frames 1-n_frames
  function's return real value is used
cinc - increase color by this value on each step (this is a float number that will be rounded to int from 0-255 range but after adding
actual color can change by +1 after 40 steps or 1 step, it depends, format "ri:gi:bi:ai"
  ri - red color increment, if any color overflows < 0 or > 255 it saturates to this value.
  gi - red color increment
  bi - red color increment
  ai - red color increment


Example final definition:
  "100|fz;r;0.5;255:0:0:255;-0.01;-0.005:0:0:0|fz;i;0.5;0:0:255:255;-0.01;0:0:-0.005:0|fz;m;1;0:255:0:255;-0.01;0:-0.005:0:0"
Example test call:
  LIB="libtet.so" U="11|fz;r;-1;255:0:0:255;x1+2*x2;0:0:0:0;1" ./cmap out.gif "x1"
  LIB="libtet.so" U="1|z;r;0;255:0:0:255;x1;0:0:0:0;1|z;i;0;0:0:255:255;x1;0:0:0:0;1|z;m;1;0:255:0:255;x1;0:0:0:0;1" ./cmap out.gif "x1"
`
		fmt.Printf("%s\n", helpStr)
	}
	dtEnd := time.Now()
	fmt.Printf("Time: %v\n", dtEnd.Sub(dtStart))
}
