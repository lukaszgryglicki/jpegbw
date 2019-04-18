package main

import (
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"math"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"
)

func srFrame(ch chan error, s, md, jpegq int, pngq png.CompressionLevel, gs, inpl bool, args []string) {
	var ma [][]*image.Image
	for i := 0; i < s; i++ {
		var t []*image.Image
		for j := 0; j < s; j++ {
			var u *image.Image
			t = append(t, u)
		}
		ma = append(ma, t)
	}
	k := 0
	px := -1
	py := -1
	x := -1
	y := -1
	for i := 0; i < s; i++ {
		for j := 0; j < s; j++ {
			reader, err := os.Open(args[k])
			if err != nil {
				ch <- err
				return
			}
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
			x = bounds.Max.X
			y = bounds.Max.Y
			if px == -1 {
				px = x
			}
			if py == -1 {
				py = y
			}
			if px != x {
				ch <- fmt.Errorf("first image x: %d, %d image x: %d (must be the same)", px, k+1, x)
				return
			}
			if py != y {
				ch <- fmt.Errorf("first image y: %d, %d image y: %d (must be the same)", py, k+1, y)
				return
			}
			ma[i][j] = &m
			k++
		}
	}
	// Detect motion
	var motion [][][2]int
	for i := 0; i < s; i++ {
		var row [][2]int
		for j := 0; j < s; j++ {
			row = append(row, [2]int{0, 0})
		}
		motion = append(motion, row)
	}
	ss := s * s
	for im := 1; im < ss; im++ {
		si := im % s
		sj := im / s
		minI := 0
		minJ := 0
		// fmt.Printf("(%d,%d,%d)\n", im, si, sj)
		minMetric := float64(1e13)
		for mdi := -md; mdi <= md; mdi++ {
			for mdj := -md; mdj <= md; mdj++ {
				metric := 0.0
				for i := 0; i < x; i++ {
					ii := i + mdi
					if ii < 0 {
						ii = 0
					}
					if ii >= x {
						ii = x - 1
					}
					for j := 0; j < y; j++ {
						jj := j + mdj
						if jj < 0 {
							jj = 0
						}
						if jj >= y {
							jj = y - 1
						}
						rr, rg, rb, _ := (*ma[0][0]).At(i, j).RGBA()
						cr, cg, cb, _ := (*ma[si][sj]).At(ii, jj).RGBA()
						r := rr + rg + rb
						c := cr + cg + cb
						metric += math.Abs(float64(r) - float64(c))
					}
				}
				metric /= float64(x * y)
				//fmt.Printf("(%d,%d) -> %f\n", mdi, mdj, metric)
				if metric < minMetric {
					minMetric = metric
					minI = mdi
					minJ = mdj
				}
			}
		}
		//fmt.Printf("Final (%d,%d,%d) --> (%d,%d,%f)\n", im, si, sj, minI, minJ, minMetric)
		motion[si][sj] = [2]int{minI, minJ}
	}
	fmt.Printf("%+v\n", motion)
	var (
		target   *image.RGBA64
		targetGS *image.Gray16
	)
	if gs {
		targetGS = image.NewGray16(image.Rect(0, 0, s*x, s*y))
	} else {
		target = image.NewRGBA64(image.Rect(0, 0, s*x, s*y))
	}
	if gs {
		for i := 0; i < x; i++ {
			for j := 0; j < y; j++ {
				for si := 0; si < s; si++ {
					for sj := 0; sj < s; sj++ {
						ii := i + motion[si][sj][0]
						if ii < 0 {
							ii = 0
						}
						if ii >= x {
							ii = x - 1
						}
						jj := j + motion[si][sj][1]
						if jj < 0 {
							jj = 0
						}
						if jj >= y {
							jj = y - 1
						}
						targetGS.Set(s*i+si, s*j+sj, (*ma[si][sj]).At(ii, jj))
					}
				}
			}
		}
	} else {
		for i := 0; i < x; i++ {
			for j := 0; j < y; j++ {
				for si := 0; si < s; si++ {
					for sj := 0; sj < s; sj++ {
						ii := i + motion[si][sj][0]
						if ii < 0 {
							ii = 0
						}
						if ii >= x {
							ii = x - 1
						}
						jj := j + motion[si][sj][1]
						if jj < 0 {
							jj = 0
						}
						if jj >= y {
							jj = y - 1
						}
						target.Set(s*i+si, s*j+sj, (*ma[si][sj]).At(ii, jj))
					}
				}
			}
		}
	}
	var (
		t   image.Image
		err error
	)
	if gs {
		t = targetGS
	} else {
		t = target
	}
	var ofn string
	if inpl {
		ofn = args[0]
	} else {
		ary := strings.Split(args[0], "/")
		lAry := len(ary)
		last := ary[lAry-1]
		ary[lAry-1] = "sr_" + last
		ofn = strings.Join(ary, "/")
	}
	fi, err := os.Create(ofn)
	if err != nil {
		ch <- err
		return
	}
	lfn := strings.ToLower(args[0])
	if strings.Contains(lfn, ".png") {
		enc := png.Encoder{CompressionLevel: pngq}
		err = enc.Encode(fi, t)
	} else if strings.Contains(lfn, ".jpg") || strings.Contains(lfn, ".jpeg") {
		var jopts *jpeg.Options
		if jpegq >= 0 {
			jopts = &jpeg.Options{Quality: jpegq}
		}
		err = jpeg.Encode(fi, t, jopts)
	} else if strings.Contains(lfn, ".gif") {
		err = gif.Encode(fi, t, nil)
	}
	if err != nil {
		_ = fi.Close()
		ch <- err
		return
	}
	err = fi.Close()
	ch <- err
	return
}

func sr(scaleS string, args []string) error {
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

	// JPEG Quality
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

	// PNG Quality
	pngqStr := os.Getenv("PQ")
	pngq := png.DefaultCompression
	if pngqStr != "" {
		v, err := strconv.Atoi(pngqStr)
		if err != nil {
			return err
		}
		if v < 0 || v > 3 {
			return fmt.Errorf("PQ must be from 0-3 range")
		}
		pngq = png.CompressionLevel(-v)
	}

	// Motion detect area
	mStr := os.Getenv("M")
	md := 1
	if mStr != "" {
		v, err := strconv.Atoi(mStr)
		if err != nil {
			return err
		}
		if v < 0 || v > 32 {
			return fmt.Errorf("M must be from 0-32 range")
		}
		md = v
	}

	// Grayscale
	gs := os.Getenv("GS") != ""

	// In-place mode
	inpl := os.Getenv("INPL") != ""

	// Scale
	scale, err := strconv.Atoi(scaleS)
	if err != nil {
		return err
	}
	if scale < 2 {
		return fmt.Errorf("scale must be >=2: %s", scaleS)
	}
	sscale := scale * scale
	n := len(args)
	ch := make(chan error)
	nThreads := 0
	for i := 0; i < n; i++ {
		to := i + sscale
		if to > n {
			break
		}
		go srFrame(ch, scale, md, jpegq, pngq, gs, inpl, args[i:to])
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
	if len(os.Args) > 5 {
		err := sr(os.Args[1], os.Args[2:])
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
	} else {
		fmt.Printf("Please provide scale >= 2 and at least 4 images to convert\n")
		helpStr := `
Parameters: scale and then scale^2 filenames, example: 2 im01.png im02.png im03.png im04.png, will generate sr_im01.png
Environment variables:
Q - jpeg quality 1-100, will use library default if not specified
PQ - png quality 0-3 (0 is default): 0=DefaultCompression, 1=NoCompression, 2=BestSpeed, 3=BestCompression
GS - set grayscale mode
INPL - set in-place mode (will overwrite input files)
N - set number of CPUs to process data
M - motion detect range araound given pixel, default 1, note that this means <1-p-1>-> 3^2 = 9 checks. (2*M+1)^2
`
		fmt.Printf("%s\n", helpStr)
	}
	dtEnd := time.Now()
	fmt.Printf("Time: %v\n", dtEnd.Sub(dtStart))
}
