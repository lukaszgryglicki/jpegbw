package main

import (
	"bufio"
	"fmt"
	"os"
	"runtime"
	"strings"
)

func gengo(args []string) error {
	var defa []string
	var ldefa []int
	defs := os.Getenv("DEFINE")
	if defs != "" {
		ary := strings.Split(defs, " ")
		for _, def := range ary {
			marker := "// " + def + ": "
			defa = append(defa, marker)
			ldefa = append(ldefa, len(marker))
		}
	}

	nThreads := 0
	thrN := runtime.NumCPU()
	runtime.GOMAXPROCS(thrN)
	che := make(chan error)

	for _, fn := range args {
		ary := strings.Split(fn, ".")
		lAry := len(ary)
		last := ary[lAry-1]
		if last != "pgo" {
			return fmt.Errorf("filename must end with .pgo: %s", fn)
		}
		ary[lAry-1] = "go"
		ofn := strings.Join(ary, ".")

		go func(c chan error, fn, ofn string) {
			fin, err := os.Open(fn)
			if err != nil {
				c <- err
				return
			}
			defer func() { _ = fin.Close() }()

			fout, err := os.Create(ofn)
			if err != nil {
				c <- err
				return
			}
			defer func() { _ = fout.Close() }()

			scanner := bufio.NewScanner(fin)
			for scanner.Scan() {
				lineOri := scanner.Text()
				line := strings.TrimSpace(lineOri)
				lLine := len(line)
				for i, marker := range defa {
					lMarker := ldefa[i]
					if lLine >= lMarker && line[:lMarker] == marker {
						idx := strings.Index(lineOri, line)
						newLine := lineOri[:idx] + line[lMarker:] + "\n"
						_, err := fout.WriteString(newLine)
						if err != nil {
							c <- err
							return
						}
					} else {
						_, err := fout.WriteString(lineOri + "\n")
						if err != nil {
							c <- err
							return
						}
					}
				}
			}

			err = scanner.Err()
			if err != nil {
				c <- err
				return
			}
			c <- nil
		}(che, fn, ofn)

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
	return nil
}

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("%s: required filename parameter\n", os.Args[0])
		return
	}
	err := gengo(os.Args[1:])
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}
