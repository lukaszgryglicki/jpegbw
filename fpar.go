package jpegbw

/*
#cgo LDFLAGS: -ldl -lm -lbyname -L./
#include "byname.h"
*/
import "C"

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"unsafe"
)

// FparCtx - context for expression parser
type FparCtx struct {
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
	cidents  map[string]*C.char
}

// Cpy - copies one context to the another, it is partially shallow copy (we copy references to maps not maps)
func (ctx *FparCtx) Cpy() FparCtx {
	// debug: fmt.Printf("copying context\n")
	// We just copy references to maps, not maps, but init is only called from single thread and then map is only read not modified
	return FparCtx{
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
		cidents:  ctx.cidents,
	}
}

// Init - initialize context, allocate internal C structs
func (ctx *FparCtx) Init(lib string, n uint) bool {
	// debug: fmt.Printf("init library: %s,%d\n", lib, n)
	clib := C.CString(lib)
	defer C.free(unsafe.Pointer(clib))
	return C.init(clib, C.size_t(n)) == 1
}

// Tidy - free memory, release context, deallocate insternal C structs
func (ctx *FparCtx) Tidy() {
	ctx.freeCIdents()
	C.tidy()
}

func (ctx *FparCtx) freeCIdents() {
	if ctx.cidents != nil {
		for _, cident := range ctx.cidents {
			C.free(unsafe.Pointer(cident))
		}
		ctx.cidents = nil
	}
}

func (ctx *FparCtx) makeCIdents() {
	ctx.freeCIdents()
	ctx.cidents = make(map[string]*C.char)
}

func (ctx *FparCtx) zeroVect() []float64 {
	vec := []float64{}
	for i := 0; i < ctx.nvar; i++ {
		vec = append(vec, 0.0)
	}
	return vec
}

func (ctx *FparCtx) er(e error) {
	// debug: if e != nil {
	// debug:   fmt.Printf("Setting error: %v, current context error: %v, position: %s\n", e, ctx.err, ctx.pos())
	// debug: }
	if e != nil && ctx.err == nil {
		ctx.err = e
	}
}

func (ctx *FparCtx) makeDigits() {
	ctx.digits = make(map[string]struct{})
	for i := 0; i <= 9; i++ {
		ctx.digits[fmt.Sprintf("%d", i)] = struct{}{}
	}
	ctx.digits["."] = struct{}{}
}

func (ctx *FparCtx) makeAlphas() {
	alphas := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ_"
	lAlphas := len(alphas)
	ctx.alphas = make(map[string]struct{})
	for i := 0; i < lAlphas; i++ {
		ctx.alphas[alphas[i:i+1]] = struct{}{}
	}
}

func (ctx *FparCtx) pos() string {
	s1 := ""
	s2 := ""
	if ctx.position > 0 {
		s1 = ctx.buffer[:ctx.position-1]
	}
	if ctx.position < ctx.maxpos {
		s2 = ctx.buffer[ctx.position:]
	}
	return fmt.Sprintf("'%s;%s' (%d/%d,ch=%s)", s1, s2, ctx.position, ctx.maxpos, ctx.ch)
}

func (ctx *FparCtx) isDigit() bool {
	_, ok := ctx.digits[ctx.ch]
	// debug2: fmt.Printf("isDigit: position: %s -> %t\n", ctx.pos(), ok)
	return ok
}

func (ctx *FparCtx) isAlpha() bool {
	_, ok := ctx.alphas[ctx.ch]
	// debug2: fmt.Printf("isAlpha: position: %s -> %t\n", ctx.pos(), ok)
	return ok
}

// FparFunction - set function expression
func (ctx *FparCtx) FparFunction(def string) error {
	if def == "" {
		return fmt.Errorf("FparFunction: empty function definition")
	}
	ctx.rbuffer = strings.Replace(def, ";", "", -1) + ";"
	ctx.buffer = strings.ToLower(ctx.rbuffer)
	ctx.maxpos = len(ctx.buffer)
	ctx.makeDigits()
	ctx.makeAlphas()
	ctx.makeCIdents()
	ctx.err = nil
	return nil
}

// FparOK - check if definition is correct
func (ctx *FparCtx) FparOK(nvar int) error {
	if nvar < 1 {
		return fmt.Errorf("FparOK: must be positive, got %v", nvar)
	}
	ctx.nvar = nvar
	_, _ = ctx.FparF(ctx.zeroVect())
	return ctx.err
}

func (ctx *FparCtx) skipBlanks() {
	if ctx.ch == ";" || ctx.position >= ctx.maxpos {
		// debug2: fmt.Printf("skipBlanks: position: %s: out of range or ;\n", ctx.pos())
		return
	}
	for strings.TrimSpace(ctx.ch) == "" && ctx.position < ctx.maxpos {
		// debug2: fmt.Printf("skipBlanks: position: %s ...\n", ctx.pos())
		ctx.readNextChar()
	}
}

func (ctx *FparCtx) readNextChar() {
	if ctx.position < ctx.maxpos && ctx.ch != ";" {
		ctx.ch = ctx.buffer[ctx.position : ctx.position+1]
		ctx.position++
		// debug: } else {
		// debug:   fmt.Printf("readNextChar: position: %s: out of range or ;\n", ctx.pos())
	}
}

func (ctx *FparCtx) readNumber() float64 {
	digitStr := ""
	for ctx.isDigit() {
		digitStr += ctx.ch
		ctx.readNextChar()
	}
	f, err := strconv.ParseFloat(digitStr, 64)
	// debug: fmt.Printf("readNumber: position: %s -> (%s,%f,%v)\n", ctx.pos(), digitStr, f, err)
	ctx.er(err)
	return f
}

func (ctx *FparCtx) readIdent() string {
	ctx.skipBlanks()
	ident := ""
	if ctx.isAlpha() {
		for ctx.isAlpha() || ctx.isDigit() {
			ident += ctx.ch
			ctx.readNextChar()
		}
	} else {
		ctx.er(fmt.Errorf("readIdent: expected function name or variable: position: %s", ctx.pos()))
	}
	ctx.skipBlanks()
	// debug: fmt.Printf("readIdent: position: %s -> %s\n", ctx.pos(), ident)
	return ident
}

func (ctx *FparCtx) callFunction(ident string) (float64, bool) {
	// debug: fmt.Printf("callFunction: position: %s %s(...)\n", ctx.pos(), ident)
	ctx.skipBlanks()
	res := 10
	if ctx.ch == "(" {
		// cident := C.CString(ident)
		// defer C.free(unsafe.Pointer(cident))
		var cident *C.char
		cident, ok := ctx.cidents[ident]
		if !ok {
			cident = C.CString(ident)
			ctx.cidents[ident] = cident
			// debug: fmt.Printf("callFunction: position: %s added '%s' to cache\n", ctx.pos(), ident)
		}
		res1 := ctx.expression()
		ctx.skipBlanks()
		v := 0.0
		if ctx.ch == ")" {
			v = float64(C.byname(cident, C.double(res1), (*C.int)(unsafe.Pointer(&res))))
			// info: fmt.Printf("callFunction: position: %s %s(%f) -> (%f,%d)\n", ctx.pos(), ident, res1, v, res)
			ctx.readNextChar()
			ctx.skipBlanks()
			if res != 0 {
				ctx.er(fmt.Errorf("error %d calling 1 argument function %s(%f): position: %s", res, ident, res1, ctx.pos()))
				return 0.0, false
			}
		} else if ctx.ch == "," {
			ctx.skipBlanks()
			res2 := ctx.expression()
			ctx.skipBlanks()
			if ctx.ch == ")" {
				v = float64(C.byname2(cident, C.double(res1), C.double(res2), (*C.int)(unsafe.Pointer(&res))))
				// info: fmt.Printf("callFunction: position: %s %s(%f,%f) -> (%f,%d)\n", ctx.pos(), ident, res1, res2, v, res)
				ctx.readNextChar()
				ctx.skipBlanks()
				if res != 0 {
					ctx.er(fmt.Errorf("error %d calling 2 arguments function %s(%f,%f): position: %s", res, ident, res1, res2, ctx.pos()))
					return 0.0, false
				}
			} else if ctx.ch == "," {
				ctx.skipBlanks()
				res3 := ctx.expression()
				ctx.skipBlanks()
				if ctx.ch == ")" {
					v = float64(C.byname3(cident, C.double(res1), C.double(res2), C.double(res3), (*C.int)(unsafe.Pointer(&res))))
					// info: fmt.Printf("callFunction: position: %s %s(%f,%f,%f) -> (%f,%d)\n", ctx.pos(), ident, res1, res2, res3, v, res)
					ctx.readNextChar()
					ctx.skipBlanks()
					if res != 0 {
						ctx.er(fmt.Errorf("error %d calling 3 arguments function %s(%f,%f,%f): position: %s", res, ident, res1, res2, res3, ctx.pos()))
						return 0.0, false
					}
				} else if ctx.ch == "," {
					ctx.skipBlanks()
					res4 := ctx.expression()
					ctx.skipBlanks()
					if ctx.ch == ")" {
						v = float64(C.byname4(cident, C.double(res1), C.double(res2), C.double(res3), C.double(res4), (*C.int)(unsafe.Pointer(&res))))
						// info: fmt.Printf("callFunction: position: %s %s(%f,%f,%f,%f) -> (%f,%d)\n", ctx.pos(), ident, res1, res2, res3, res4, v, res)
						ctx.readNextChar()
						ctx.skipBlanks()
						if res != 0 {
							ctx.er(fmt.Errorf("error %d calling 4 arguments function %s(%f,%f,%f,%f): position: %s", res, ident, res1, res2, res3, res4, ctx.pos()))
							return 0.0, false
						}
					} else {
						ctx.er(fmt.Errorf("expected: ')' after 4 arguments function %s(%f,%f,%f: position: %s", ident, res1, res2, res3, ctx.pos()))
						return 0.0, false
					}
				} else {
					ctx.er(fmt.Errorf("expected: ')' after 3 arguments function %s(%f,%f,: position: %s", ident, res1, res2, ctx.pos()))
					return 0.0, false
				}
			} else {
				ctx.er(fmt.Errorf("expected: ')' after 2 arguments function %s(%f,: position: %s", ident, res1, ctx.pos()))
				return 0.0, false
			}
		} else {
			ctx.er(fmt.Errorf("expected: ')' after 1 argument function %s: position: %s", ident, ctx.pos()))
			return 0.0, false
		}
		return v, true
	}
	ctx.er(fmt.Errorf("callFunction: expected '(' after %s: position: %s", ident, ctx.pos()))
	return 0.0, false
}

func (ctx *FparCtx) argVal(ident string) (float64, bool) {
	if ident == "" {
		// debug: fmt.Printf("argVal: position: %s '' -> 0,false\n", ctx.pos())
		return 0.0, false
	}
	if ident[:1] == "x" {
		num, err := strconv.Atoi(ident[1:])
		if err != nil || num < 1 || num > ctx.nvar {
			// debug: fmt.Printf("argVal: position: %s ident=%s -> (%d,%v) -> 0,false\n", ctx.pos(), ident, num, err)
			return 0.0, false
		}
		// debug: fmt.Printf("argVal: position: %s ident=%s -> x%d -> %f,true\n", ctx.pos(), ident, num, ctx.arg[num-1])
		return ctx.arg[num-1], true
	}
	// debug: fmt.Printf("argVal: position: %s ident=%s -> 0,false\n", ctx.pos(), ident)
	return 0.0, false
}

func (ctx *FparCtx) factor() float64 {
	f := 0.0
	minus := 1.0
	ctx.readNextChar()
	ctx.skipBlanks()
	for ctx.ch == "+" || ctx.ch == "-" {
		if ctx.ch == "-" {
			// debug: fmt.Printf("factor: position: %s minus\n", ctx.pos())
			minus *= -1.0
		}
		ctx.readNextChar()
	}
	if ctx.isDigit() {
		f = ctx.readNumber()
		ctx.skipBlanks()
	} else if ctx.ch == "(" {
		// debug: fmt.Printf("factor: position: %s new expression in (\n", ctx.pos())
		f = ctx.expression()
		ctx.skipBlanks()
		if ctx.ch == ")" {
			// debug: fmt.Printf("factor: position: %s expression in ) finished\n", ctx.pos())
			ctx.readNextChar()
			ctx.skipBlanks()
		} else {
			ctx.er(fmt.Errorf("expected: ')': position: %s", ctx.pos()))
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
				ctx.er(fmt.Errorf("don't know what to do with '%s': position: %s", ident, ctx.pos()))
			}
		}
	}
	ctx.skipBlanks()
	// debug: fmt.Printf("factor: position: %s -> %f\n", ctx.pos(), f * minus)
	return f * minus
}

func (ctx *FparCtx) exponential() float64 {
	f := ctx.factor()
	for ctx.ch == "^" {
		// debug: fmt.Printf("exponential: position: %s %f ^ ...\n", ctx.pos(), f)
		f = math.Pow(f, ctx.exponential())
		// debug: fmt.Printf("exponential: position: %s -> %f\n", ctx.pos(), f)
	}
	return f
}

func (ctx *FparCtx) term() float64 {
	f := ctx.exponential()
	for {
		switch ctx.ch {
		case "*":
			// debug: fmt.Printf("term: position: %s %f * ...\n", ctx.pos(), f)
			f *= ctx.exponential()
			// debug: fmt.Printf("term: position: %s -> %f\n", ctx.pos(), f)
		case "/":
			// debug: fmt.Printf("term: position: %s %f / ...\n", ctx.pos(), f)
			f /= ctx.exponential()
			// debug: fmt.Printf("term: position: %s -> %f\n", ctx.pos(), f)
		default:
			return f
		}
	}
}

func (ctx *FparCtx) expression() float64 {
	t := ctx.term()
	for {
		switch ctx.ch {
		case "+":
			// debug: fmt.Printf("expression: position: %s %f + ...\n", ctx.pos(), t)
			t += ctx.term()
			// debug: fmt.Printf("expression: position: %s -> %f\n", ctx.pos(), t)
		case "-":
			// debug: fmt.Printf("expression: position: %s %f - ...\n", ctx.pos(), t)
			t -= ctx.term()
			// debug: fmt.Printf("expression: position: %s -> %f\n", ctx.pos(), t)
		default:
			return t
		}
	}
}

// FparF - call user defined function
func (ctx *FparCtx) FparF(args []float64) (float64, error) {
	ctx.err = nil
	ctx.arg = args
	ctx.position = 0
	ctx.ch = ""
	// debug: fmt.Printf("FparF: position: %s f(%v) ...\n", ctx.pos(), args)
	e := ctx.expression()
	if ctx.ch != ";" {
		ctx.er(fmt.Errorf("FparF: garbage in function expression"))
	}
	// info: fmt.Printf("FparF: position: %s f(%v) = %f\n", ctx.pos(), args, e)
	return e, ctx.err
}
