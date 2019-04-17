package main

import (
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"runtime"
	"strconv"
	"time"
)

func srFrame(ch chan error, s, ss int, args []string) {
	fmt.Printf("%d %d %+v\n", s, ss, args)
	///////////
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
			x := bounds.Max.X
			y := bounds.Max.Y
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
			k++
		}
	}
	ch <- nil
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
		go srFrame(ch, scale, sscale, args[i:to])
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
N - set number of CPUs to process data
`
		fmt.Printf("%s\n", helpStr)
	}
	dtEnd := time.Now()
	fmt.Printf("Time: %v\n", dtEnd.Sub(dtStart))
}
