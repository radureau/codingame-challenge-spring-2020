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
	return printElapsedTimeSince(time.Now(), name)
}

func printElapsedTimeSince(since time.Time, name string) func() {
	start := since
	return func() {
		elapsed := time.Since(start)
		// codingame environment seems to multiply cpu time by a factor of 4
		if os.Getenv("USER") == "__USER__" {
			elapsed *= 4
		}
		debug(fmt.Sprintf("\t%s\ttook %dms", name, elapsed.Milliseconds()))
	}
}
