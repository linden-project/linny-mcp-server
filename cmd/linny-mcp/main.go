package main

import (
	"os"
)

func main() {
	os.Exit(Run(os.Args, os.Stdout, os.Stderr))
}
