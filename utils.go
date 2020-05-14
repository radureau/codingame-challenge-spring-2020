package main

import (
	"fmt"
	"os"
	"time"
)

func debug(args ...interface{}) {
	fmt.Fprintln(os.Stderr, args...)
}

func printElapsedTime(name string) func() {
	start := time.Now()
	return func() {
		elapsed := time.Since(start)
		// codingame environment seems to multiply cpu time by a factor of 4
		debug(fmt.Sprintf("\t%s\ttook %dms", name, elapsed.Milliseconds()*4))
	}
}
