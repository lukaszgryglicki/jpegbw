package main

import (
	"fmt"
	"jpegbw"
	"os"
	"runtime"
	"strconv"
	"time"
)

func cmap(ofn, f string) error {
	var fctx jpegbw.FparCtx

	// LIB, NF
	lib := os.Getenv("LIB")
	if lib == "" {
		return fmt.Errorf("you must specify dynamic library for functions via LIB env variable")
	}
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
	ch := make(chan error)
	dtStart := time.Now()
	for ii := 0; ii < x; ii++ {
		go func(ch chan error, i int) {
			for j := 0; j < y; j++ {
			}
			ch <- nil
		}(ch, ii)

		nThreads++
		if nThreads == thrN {
			e := <-ch
			if e != nil {
				return e
			}
			nThreads--
		}
	}
	for nThreads > 0 {
		e := <-ch
		if e != nil {
			return e
		}
		nThreads--
	}
	dtEnd := time.Now()
	pps := (all / dtEnd.Sub(dtStart).Seconds()) / 1048576.0
	fmt.Printf("Processing: %v, MPPS: %.3f\n", dtEnd.Sub(dtStart), pps)
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
`
		fmt.Printf("%s\n", helpStr)
	}
	dtEnd := time.Now()
	fmt.Printf("Time: %v\n", dtEnd.Sub(dtStart))
}
