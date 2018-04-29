package main

import (
	"fmt"
	"jpegbw"
	"math/cmplx"
	"os"
	"strconv"
)

func fCall(f string, za []string) error {
	var fzc jpegbw.FparCtx
	var zca [4]jpegbw.FparCtx

	if len(za) > 4 {
		return fmt.Errorf("maximum 4 arguments are allowed")
	}

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
		ok := fzc.Init(lib, uint(nf))
		if !ok {
			return fmt.Errorf("LIB init failed for: %s", lib)
		}
		defer func() { fzc.Tidy() }()
	}
	err := fzc.FparFunction(f)
	if err != nil {
		return err
	}
	err = fzc.FparOK(4)
	if err != nil {
		return err
	}

	var zar [4]complex128
	for i := 0; i < 4; i++ {
		def := "0"
		if len(za) > i {
			def = za[i]
		}
		err := zca[i].FparFunction(def)
		if err != nil {
			return err
		}
		err = zca[i].FparOK(1)
		if err != nil {
			return err
		}
		zar[i], err = zca[i].FparF([]complex128{complex(0.0, 0.0)})
		if err != nil {
			return err
		}
	}
	fz, err := fzc.FparF(zar[:])
	s := "f("
	s2 := "|'" + f + "'("
	csvr := ""
	for i := 0; i < 4; i++ {
		if len(za) > i {
			s += fmt.Sprintf("%v+%vi, ", real(zar[i]), imag(zar[i]))
			s2 += za[i] + ", "
			csvr += fmt.Sprintf("%v,%v,%v,", real(zar[i]), imag(zar[i]), cmplx.Abs(zar[i]))
		} else {
			break
		}
	}
	l := len(s)
	if len(za) > 0 {
		s = s[:l-2]
	}
	s += ") = "
	s += fmt.Sprintf("%v+%vi", real(fz), imag(fz))
	l2 := len(s2)
	if len(za) > 0 {
		s2 = s2[:l2-2]
	}
	s2 += ")| = "
	s2 += fmt.Sprintf("%v", cmplx.Abs(fz))
	fmt.Printf("%s\n%s\n", s, s2)
	csvr += fmt.Sprintf("%v,%v,%v\n", real(fz), imag(fz), cmplx.Abs(fz))
	fmt.Fprintf(os.Stderr, csvr)
	return err
}

func main() {
	if len(os.Args) >= 2 {
		err := fCall(os.Args[1], os.Args[2:])
		if err != nil {
			fmt.Printf("%v\n", err)
		}
		return
	}
	fmt.Printf("LIB=libtet.so %s 'csin(x1)*ccos(x2)*cpow(x3, x4)' 1 -2 _3 -_4\n", os.Args[0])
}
