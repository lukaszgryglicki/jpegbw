package main

import (
	"encoding/json"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io/ioutil"
	"os"
	"runtime"
	"strconv"
	"time"

	"github.com/lukaszgryglicki/jpegbw"
)

func hist(args []string) error {
	// Parse env
	rgba := [4]string{"R", "G", "B", "A"}
	var (
		ahi [4]float64
		alo [4]float64
	)
	// No alpha processing
	noA := os.Getenv("NA") != ""

	// No histogram file write
	wH := os.Getenv("WH") != ""

	// Number of frames to merge histogram data (MF moving average MF MA)
	n := len(args)
	mfS := os.Getenv("MF")
	mf := 16
	if mfS != "" {
		m, err := strconv.Atoi(mfS)
		if err != nil {
			return err
		}
		if m < 1 || m > n {
			return fmt.Errorf("MF must be from 1-%d range", n)
		}
		mf = m
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

	// Process colors
	for c, colrgba := range rgba {
		if noA && c == 3 {
			continue
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

		alo[c] = lo
		ahi[c] = hi
	}

	// Iterate given files
	ch := make(chan error)
	nThreads := 0
	allHist := [][4]jpegbw.IntHist{}
	allN := []float64{}
	for range args {
		allHist = append(allHist, [4]jpegbw.IntHist{nil, nil, nil, nil})
		allN = append(allN, 0.0)
	}
	for k, fn := range args {
		go func(ch chan error, fn string, k int) {
			// Input
			reader, err := os.Open(fn)
			if err != nil {
				ch <- err
				return
			}

			// Decode input
			m, _, err := image.Decode(reader)
			if err != nil {
				_ = reader.Close()
				ch <- err
				return
			}
			err = reader.Close()
			if err != nil {
				ch <- err
				return
			}
			bounds := m.Bounds()
			x := bounds.Max.X
			y := bounds.Max.Y

			// Data structure for pixels
			var pxdata [][][4]uint16
			for i := 0; i < x; i++ {
				pxdata = append(pxdata, [][4]uint16{})
				for j := 0; j < y; j++ {
					pxdata[i] = append(pxdata[i], [4]uint16{0, 0, 0, 0})
				}
			}

			// Get pixel data
			for i := 0; i < x; i++ {
				for j := 0; j < y; j++ {
					pr, pg, pb, pa := m.At(i, j).RGBA()
					pxdata[i][j] = [4]uint16{uint16(pr), uint16(pg), uint16(pb), uint16(pa)}
				}
			}

			// Convert
			all := float64(x * y)
			allN[k] = all
			var fh jpegbw.FileHist

			// Process RGBA histograms
			for c := 0; c < 4; c++ {
				if noA && c == 3 {
					continue
				}
				lo := alo[c]
				hi := ahi[c]

				hist := make(jpegbw.IntHist)
				minGs := uint16(0xffff)
				maxGs := uint16(0)

				for i := 0; i < x; i++ {
					for j := 0; j < y; j++ {
						gs := pxdata[i][j][c]
						if gs < minGs {
							minGs = gs
						}
						if gs > maxGs {
							maxGs = gs
						}
						hist[gs]++
					}
				}
				//fmt.Printf("hist(%d): %+v\n", c, hist.Str())
				// info: fmt.Printf("hist: %+v\n", hist.Str())

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
					ch <- fmt.Errorf("%s:%s calculated integer range is empty: %d-%d", fn, rgba[c], loI, hiI)
					return
				}
				// info: fmt.Printf("histCum: %+v\n", histCum.Str())
				// info: mult := 65535.0 / float64(hiI-loI)
				// info: fmt.Printf("%s:%s %04x - %04x -> range(%f%%-%f%%): %04x - %04x, mult: %f\n", fn, rgba[c], minGs, maxGs, lo, hi, loI, hiI, mult)

				// Update all hist - no mutex needed
				allHist[k][c] = hist

				// Write histogram data
				if wH {
					fh.Hist[c] = hist
					fh.HistCum[c] = histCum
				}
			}
			if wH {
				fh.Fn = fn
				err = fh.WriteHist()
				if err != nil {
					ch <- err
					return
				}
			}
			ch <- nil
			return
		}(ch, fn, k)
		nThreads++
		if nThreads == thrN {
			err := <-ch
			nThreads--
			if err != nil {
				return err
			}
		}
	}
	for nThreads > 0 {
		err := <-ch
		nThreads--
		if err != nil {
			return err
		}
	}

	// Create moving histograms
	mf2 := mf >> 1
	for k := 0; k < n; k++ {
		f := k - mf2
		t := k + mf2
		if f == t {
			t++
		}
		if f < 0 {
			f = 0
		}
		if t > n {
			t = n
		}
		go func(ch chan error, k, f, t int) {
			var hint jpegbw.HintData
			hint.From = f
			hint.To = t
			hint.Curr = k
			for c := 0; c < 4; c++ {
				if noA && c == 3 {
					continue
				}
				lo := alo[c]
				hi := ahi[c]
				hint.LoPerc[c] = lo
				hint.HiPerc[c] = hi
				hist := make(jpegbw.IntHist)
				minV := uint16(0xffff)
				maxV := uint16(0)
				all := 0.0
				for ma := f; ma < t; ma++ {
					all += allN[ma]
					for idx, val := range allHist[ma][c] {
						v, ok := hist[idx]
						if ok {
							hist[idx] = v + val
						} else {
							hist[idx] = val
						}
						if idx < minV {
							minV = idx
						}
						if idx > maxV {
							maxV = idx
						}
					}
				}
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
					ch <- fmt.Errorf("%s:%s calculated integer range is empty: %d-%d", args[k], rgba[c], loI, hiI)
					return
				}
				hint.Mult[c] = 65535.0 / float64(hiI-loI)
				hint.Min[c] = minV
				hint.Max[c] = maxV
				hint.LoIdx[c] = loI
				hint.HiIdx[c] = hiI
				// info: fmt.Printf("> %s:%s[%d-%d]: %04x-%04x -> range(%f%%-%f%%): %04x - %04x, mult: %f\n", args[k], rgba[c], f, t, minV, maxV, lo, hi, loI, hiI, hint.Mult[c])
			}
			// Write hint
			fn := args[k] + ".hint"
			jsonBytes, err := json.Marshal(hint)
			if err != nil {
				ch <- err
				return
			}
			err = ioutil.WriteFile(fn, jsonBytes, 0644)
			ch <- err
			return
		}(ch, k, f, t)
		nThreads++
		if nThreads == thrN {
			err := <-ch
			nThreads--
			if err != nil {
				return err
			}
		}
	}
	for nThreads > 0 {
		err := <-ch
		nThreads--
		if err != nil {
			return err
		}
	}
	return nil
}

func main() {
	dtStart := time.Now()
	if len(os.Args) > 1 {
		err := hist(os.Args[1:])
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
WH - write *.hist files
MF - merge frames (calculate histogram from MF frames), moving histogram, default 16
XLO - when calculating intensity range, discard values than are in this lower %, for example 3
XHI - when calculating intensity range, discard values that are in this higher %, for example 3
N - set number of CPUs to process data
`
		fmt.Printf("%s\n", helpStr)
	}
	dtEnd := time.Now()
	fmt.Printf("Time: %v\n", dtEnd.Sub(dtStart))
}
