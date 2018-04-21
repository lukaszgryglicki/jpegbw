package main

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
	"strconv"
	"strings"
)

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
func images2BW(args []string) error {
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
	fmt.Printf("Final RGB multiplier: %f(%f, %f, %f), range %f%% - %f%%, quality: %d, gamma: (%v, %f)\n", fact, r, g, b, lo, hi, jpegq, gaB, ga)

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
		for i := 0; i < x; i++ {
			for j := 0; j < y; j++ {
				// target.Set(i, j, m.At(i, j))
				pr, pg, pb, _ := m.At(i, j).RGBA()
				// fmt.Printf("%d,%d,%d\n", pr, pg, pb)
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
		for i := 0; i < x; i++ {
			for j := 0; j < y; j++ {
				pr, pg, pb, _ := m.At(i, j).RGBA()
				// fmt.Printf("%d,%d,%d\n", pr, pg, pb)
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
				gs = uint16(fv)
				pixel := color.Gray16{gs}
				target.Set(i, j, pixel)
			}
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
