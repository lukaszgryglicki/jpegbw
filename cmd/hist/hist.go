package main

import (
	"fmt"
	"os"
	"time"
)

func hist(files []string) error {
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
