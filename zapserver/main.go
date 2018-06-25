// Zap Collection Server
package main

import (
	"flag"
	"log"
	"fmt"
	"os"
	"os/signal"
	"runtime/pprof"
)

func Usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS]\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "\nOptions:\n")
	flag.PrintDefaults()
}

func parseFlags() {
	flag.Usage = Usage
	flag.Parse()
	if *showHelp {
		flag.Usage()
		os.Exit(0)
	}
}

func main() {
	parseFlags()
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Kill, os.Interrupt)

	err := runLab()
	if err != nil {
		panic(err)
	}

	// Wait for CTRL-C or other kill signal
	s := <-signalChan
	fmt.Println("Server stopping on", s, "signal")
	if *memprofile != "" {
		f, err := os.Create(*memprofile)
		if err != nil {
			log.Fatal(err)
		}

		pprof.WriteHeapProfile(f)
		f.Close()
		fmt.Println("Saved memory profile")
		fmt.Println("Analyze with: go tool pprof $GOPATH/bin/zapserver", *memprofile)
	}
}
