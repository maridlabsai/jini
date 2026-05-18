package main

import (
	"os"

	"github.com/maridlabsai/jini/internal/app"
)

func main() {
	os.Exit(app.RunInteractive(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}
