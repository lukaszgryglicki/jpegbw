package main

import (
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"jpegbw"
	"os"
	"runtime"
	"strconv"
	"time"
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
	n := len(args)
	ch := make(chan error)
	nThreads := 0
	for k, fn := range args {
		fmt.Printf("File %d/%d %s\n", k+1, n, fn)
		go func(ch chan error, fn string, k int) {
			// Input
			reader, err := os.Open(fn)
			if err != nil {
				ch <- err
				return
			}

			// Decode input
			m, _, err := image.Decode(reader)
			_ = reader.Close()
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
					ch <- fmt.Errorf("%s:%s calculated integer range is empty: %d-%d", fn, rgba[c], loI, hiI)
					return
				}
				mult := 65535.0 / float64(hiI-loI)
				// info: fmt.Printf("histCum: %+v\n", histCum.Str())
				fmt.Printf("%s:%s range(%f%%-%f%%): %04x - %04x, mult: %f\n", fn, rgba[c], lo, hi, loI, hiI, mult)
				fh.Hist[c] = hist
				fh.HistCum[c] = histCum
				fh.Fn = fn
				ch <- fh.WriteHist()
				return
			}
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
	}
	dtEnd := time.Now()
	fmt.Printf("Time: %v\n", dtEnd.Sub(dtStart))
}
