package main

/*
#cgo LDFLAGS: -ldl
#include <stdlib.h>
#include <stdio.h>
#include <string.h>
#include <dlfcn.h>

#define MAXFN 2048

static void* handle = 0;
static double (**fptra)(double) = 0;
static char** fnames = 0;
static int nptrs = 0;

double byname(char* fname, double arg) {
  int i;
  double (*fptr)(double) = 0;
  if (!handle) {
    printf("byname %s: library not open\n", fname);
    return 0.0;
  }
  for (i=0;i<nptrs;i++) {
    if (!strcmp(fnames[i], fname)) {
      fptr = fptra[i];
      // printf("found %s at %p, n=%d\n", fname, fptr, nptrs);
    }
  }
  if (!fptr) {
    if (nptrs >= MAXFN) {
      printf("byname %s: function table full\n", fname);
      return 0.0;
    }
    fptr = (double (*)(double))dlsym(handle, fname);
    if (!fptr) {
      printf("byname %s: function not found\n", fname);
      return 0.0;
    }
    fptra[nptrs] = fptr;
    fnames[nptrs] = (char*)malloc((strlen(fname)+1)*sizeof(char));
    strcpy(fnames[nptrs], fname);
    nptrs ++;
  }
  return (*fptr)(arg);
}

int init(char* lib) {
  handle = dlopen(lib, RTLD_LAZY);
  if (!handle) {
    printf("%s load result: %p\n", lib, handle);
    return 0;
  }
  fptra = malloc(MAXFN*sizeof(void*));
  fnames = (char**)malloc(MAXFN*sizeof(char*));
  if (!fptra || !fnames) {
    printf("%s malloc failed\n", lib);
    return 0;
  }
  return 1;
}
*/
import "C"

import (
	"bufio"
	"fmt"
	"image"
	"image/color"
	"image/gif"
	"image/jpeg"
	"image/png"
	"math"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"unsafe"
)

// fpar: start
type fparCtx struct {
	buffer   string
	rbuffer  string
	position int
	ch       string
	maxpos   int
	err      error
	arg      []float64
	nvar     int
	digits   map[string]struct{}
	alphas   map[string]struct{}
}

func (ctx *fparCtx) cpy() fparCtx {
	return fparCtx{
		buffer:   ctx.buffer,
		rbuffer:  ctx.rbuffer,
		position: ctx.position,
		ch:       ctx.ch,
		maxpos:   ctx.maxpos,
		err:      nil,
		arg:      []float64{},
		nvar:     ctx.nvar,
		digits:   ctx.digits,
		alphas:   ctx.alphas,
	}
}

func (ctx *fparCtx) init(lib string) bool {
	clib := C.CString(lib)
	defer C.free(unsafe.Pointer(clib))
	res := C.init(clib)
	return res == 1
}

func (ctx *fparCtx) zeroVect() []float64 {
	vec := []float64{}
	for i := 0; i < ctx.nvar; i++ {
		vec = append(vec, 0.0)
	}
	return vec
}

func (ctx *fparCtx) er(e error) {
	if e != nil && ctx.err == nil {
		ctx.err = e
	}
}

func (ctx *fparCtx) makeDigits() {
	ctx.digits = make(map[string]struct{})
	for i := 0; i <= 9; i++ {
		ctx.digits[fmt.Sprintf("%d", i)] = struct{}{}
	}
	ctx.digits["."] = struct{}{}
}

func (ctx *fparCtx) makeAlphas() {
	alphas := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ_"
	lAlphas := len(alphas)
	ctx.alphas = make(map[string]struct{})
	for i := 0; i < lAlphas; i++ {
		ctx.alphas[alphas[i:i+1]] = struct{}{}
	}
}

func (ctx *fparCtx) isDigit() bool {
	_, ok := ctx.digits[ctx.ch]
	return ok
}

func (ctx *fparCtx) isAlpha() bool {
	_, ok := ctx.alphas[ctx.ch]
	return ok
}

func (ctx *fparCtx) fparFunction(def string) error {
	if def == "" {
		return fmt.Errorf("fparFunction: empty function definition")
	}
	ctx.rbuffer = strings.Replace(def, ";", "", -1) + ";"
	ctx.buffer = strings.ToLower(ctx.rbuffer)
	ctx.maxpos = len(ctx.buffer)
	ctx.makeDigits()
	ctx.makeAlphas()
	ctx.err = nil
	return nil
}

func (ctx *fparCtx) fparOK(nvar int) error {
	if nvar < 1 {
		return fmt.Errorf("fparOK: must be positive, got %v", nvar)
	}
	ctx.nvar = nvar
	_, _ = ctx.fparF(ctx.zeroVect())
	return ctx.err
}

func (ctx *fparCtx) skipBlanks() {
	if ctx.ch == ";" || ctx.position >= ctx.maxpos {
		return
	}
	for strings.TrimSpace(ctx.ch) == "" && ctx.position < ctx.maxpos {
		ctx.readNextChar()
	}
}

func (ctx *fparCtx) readNextChar() {
	if ctx.position < ctx.maxpos && ctx.ch != ";" {
		ctx.ch = ctx.buffer[ctx.position : ctx.position+1]
		ctx.position++
	}
}

func (ctx *fparCtx) readNumber() float64 {
	digitStr := ""
	for ctx.isDigit() {
		digitStr += ctx.ch
		ctx.readNextChar()
	}
	f, err := strconv.ParseFloat(digitStr, 64)
	ctx.er(err)
	return f
}

func (ctx *fparCtx) readIdent() string {
	ctx.skipBlanks()
	ident := ""
	if ctx.isAlpha() {
		for ctx.isAlpha() || ctx.isDigit() {
			ident += ctx.ch
			ctx.readNextChar()
		}
	} else {
		ctx.er(fmt.Errorf("readIdent: expected function name or variable: position: (%d/%d,ch=%s)", ctx.position, ctx.maxpos, ctx.ch))
	}
	ctx.skipBlanks()
	return ident
}

func (ctx *fparCtx) double(in float64) float64 {
	return 2.0 * in
}

func (ctx *fparCtx) callFunction(ident string) (float64, bool) {
	ctx.skipBlanks()
	if ctx.ch == "(" {
		cident := C.CString(ident)
		defer C.free(unsafe.Pointer(cident))
		v := float64(C.byname(cident, C.double(ctx.expression())))
		ctx.skipBlanks()
		if ctx.ch == ")" {
			ctx.readNextChar()
			ctx.skipBlanks()
		} else {
			ctx.er(fmt.Errorf("expected: ')' after %s: position: (%d/%d,ch=%s)", ident, ctx.position, ctx.maxpos, ctx.ch))
		}
		return v, true
	}
	ctx.er(fmt.Errorf("callFunction: expected '(' after %s: position: (%d/%d,ch=%s)", ident, ctx.position, ctx.maxpos, ctx.ch))
	return 0.0, false
}

func (ctx *fparCtx) argVal(ident string) (float64, bool) {
	if ident == "" {
		return 0.0, false
	}
	if ident[:1] == "x" {
		num, err := strconv.Atoi(ident[1:])
		if err != nil || num < 1 || num > ctx.nvar {
			return 0.0, false
		}
		return ctx.arg[num-1], true
	}
	return 0.0, false
}

func (ctx *fparCtx) factor() float64 {
	f := 0.0
	minus := 1.0
	ctx.readNextChar()
	ctx.skipBlanks()
	for ctx.ch == "+" || ctx.ch == "-" {
		if ctx.ch == "-" {
			minus *= -1.0
		}
		ctx.readNextChar()
	}
	if ctx.isDigit() {
		f = ctx.readNumber()
		ctx.skipBlanks()
	} else if ctx.ch == "(" {
		f = ctx.expression()
		ctx.skipBlanks()
		if ctx.ch == ")" {
			ctx.readNextChar()
			ctx.skipBlanks()
		} else {
			ctx.er(fmt.Errorf("expected: ')': position: (%d/%d,ch=%s)", ctx.position, ctx.maxpos, ctx.ch))
		}
	} else {
		ident := ctx.readIdent()
		arg, isArg := ctx.argVal(ident)
		if isArg {
			f = arg
		} else {
			val, gotVal := ctx.callFunction(ident)
			if gotVal {
				f = val
			} else {
				ctx.er(fmt.Errorf("don't know what to do with '%s': position: (%d/%d,ch=%s)", ident, ctx.position, ctx.maxpos, ctx.ch))
			}
		}
	}
	ctx.skipBlanks()
	return f * minus
}

func (ctx *fparCtx) exponential() float64 {
	f := ctx.factor()
	for ctx.ch == "^" {
		f = math.Pow(f, ctx.exponential())
	}
	return f
}

func (ctx *fparCtx) term() float64 {
	f := ctx.exponential()
	for {
		switch ctx.ch {
		case "*":
			f *= ctx.exponential()
		case "/":
			f /= ctx.exponential()
		default:
			return f
		}
	}
}

func (ctx *fparCtx) expression() float64 {
	t := ctx.term()
	for {
		switch ctx.ch {
		case "+":
			t += ctx.term()
		case "-":
			t -= ctx.term()
		default:
			return t
		}
	}
}

func (ctx *fparCtx) fparF(args []float64) (float64, error) {
	ctx.err = nil
	ctx.arg = args
	ctx.position = 0
	ctx.ch = ""
	e := ctx.expression()
	if ctx.ch != ";" {
		ctx.er(fmt.Errorf("fparF: garbage in function expression"))
	}
	return e, ctx.err
}

// fpar: end

// images2BW: convert given images to bw: iname.ext -> bw_iname.ext, dir/iname.ext -> dir/bw_iname.ext
// Other parameters are set via env variable:
// Q - jpeg quality 1-100, will use library default if not specified
// R - relative red usage for generating gray pixel, 1 if not specified
// G - relative green usage for generating gray pixel, 1 if not specified
// B - relative blue usage for generating gray pixel, 1 if not specified
// (R+G+B) will be normalized to sum to 1, so their sum must be positive
// R=0.2125 G=0.7154 B=0.0721 is a suggested configuration
// LO - when calculating intensity range, discard values than are in this lower %, for example 3
// HI - when calculating intensity range, discard values that are in this higher %, for example 3
// GA - gamma default 1, which uses straight line (0,0) -> (1,1), if set uses (x,y)->(x,pow(x, GA)) mapping
// F - function to apply on final 0-1 range, for example "sin(x1*2)+cos(x1*3)"
// LIB - if F is used and F calls external functions, thery need to be loaded for this C library
func images2BW(args []string) error {
	// F, LIB processing
	var fctx fparCtx
	fun := os.Getenv("F")
	lib := ""
	bFun := false
	if fun != "" {
		lib = os.Getenv("LIB")
		if lib != "" {
			ok := fctx.init(lib)
			if !ok {
				return fmt.Errorf("LIB init failed for: %s", lib)
			}
		}
		err := fctx.fparFunction(fun)
		if err != nil {
			return err
		}
		err = fctx.fparOK(3)
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
	thrN := runtime.NumCPU()
	runtime.GOMAXPROCS(thrN)
	fmt.Printf("Final RGB multiplier: %f(%f, %f, %f), range %f%% - %f%%, quality: %d, gamma: (%v, %f), threads: %d\n", fact, r, g, b, lo, hi, jpegq, gaB, ga, thrN)

	// Flushing before endline
	flush := bufio.NewWriter(os.Stdout)

	// Iterate given files
	n := len(args)
	for i, fn := range args {
		fmt.Printf("%d/%d %s...", i+1, n, fn)
		_ = flush.Flush()

		// Input
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

		// Output
		target := image.NewGray16(image.Rect(0, 0, x, y))

		// Convert
		hist := make(map[uint16]int)
		minGs := uint16(0xffff)
		maxGs := uint16(0)
		var mtx = &sync.Mutex{}

		ch := make(chan bool)
		nThreads := 0
		for ii := 0; ii < x; ii++ {
			go func(c chan bool, i int) {
				for j := 0; j < y; j++ {
					// target.Set(i, j, m.At(i, j))
					pr, pg, pb, _ := m.At(i, j).RGBA()
					// fmt.Printf("%d,%d,%d\n", pr, pg, pb)
					gs := uint16(r*float64(pr) + g*float64(pg) + b*float64(pb))
					mtx.Lock()
					if gs < minGs {
						minGs = gs
					}
					if gs > maxGs {
						maxGs = gs
					}
					hist[gs]++
					mtx.Unlock()
				}
				// Sync
				c <- true
			}(ch, ii)

			// Keep maximum number of threads
			nThreads++
			if nThreads == thrN {
				<-ch
				nThreads--
			}
		}
		for nThreads > 0 {
			<-ch
			nThreads--
		}

		// Calculations
		all := float64(x * y)
		histCum := make(map[uint16]float64)
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
			return fmt.Errorf("calculated integer range is empty: %d-%d", loI, hiI)
		}
		mult := 65535.0 / float64(hiI-loI)
		fmt.Printf(" gray: (%d, %d) int: (%d, %d) mult: %f...", minGs, maxGs, loI, hiI, mult)
		_ = flush.Flush()

		che := make(chan error)
		nThreads = 0
		ctxa := []fparCtx{}
		ctxInUse := make(map[int]bool)
		for i := 0; i < thrN; i++ {
			ctxa = append(ctxa, fctx.cpy())
			ctxInUse[i] = false
		}
		var cmtx = &sync.Mutex{}
		for ii := 0; ii < x; ii++ {
			go func(c chan error, i int) {
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
						fv, e = ctxa[cNum].fparF([]float64{fv / 65535.0, fi, fj})
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

		// Close reader
		err = reader.Close()
		if err != nil {
			return err
		}

		// Output name
		ary := strings.Split(fn, "/")
		lAry := len(ary)
		last := ary[lAry-1]
		ary[lAry-1] = "bw_" + last
		ofn := strings.Join(ary, "/")
		fi, err := os.Create(ofn)
		if err != nil {
			return err
		}
		lfn := strings.ToLower(fn)

		// Output write
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
			return ierr
		}
		err = fi.Close()
		if err != nil {
			return err
		}
		fmt.Printf(" %s\n", ofn)
	}
	return nil
}

func main() {
	if len(os.Args) > 1 {
		err := images2BW(os.Args[1:])
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
	} else {
		fmt.Printf("Please provide at least one image to convert\n")
	}
}
