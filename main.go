package main

import (
	"flag"
	"github.com/YogurtRaisin/gurl/timing"
	"os"
)

var (
	trgOpt string
)

func argsInit() {
	flag.StringVar(&trgOpt, "t", "none", "target url")
	flag.StringVar(&trgOpt, "target", "none", "target url")

	flag.Parse()
}

func main() {
	argsInit()

	if trgOpt == "none" {
		flag.Usage()
		os.Exit(1)
	}

	timing.Timing(trgOpt)
}
