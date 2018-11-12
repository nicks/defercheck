package main

import (
	"github.com/nicks/defercheck/internal/defercheck"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
	singlechecker.Main(defercheck.Analyzer)
}
