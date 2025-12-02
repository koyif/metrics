package main

import (
	"github.com/koyif/metrics/pkg/exitcheck"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
	singlechecker.Main(exitcheck.Analyzer)
}