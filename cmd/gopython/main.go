package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/tamnd/gopython/py"
)

func main() {
	version := flag.Bool("version", false, "print version")
	flag.Parse()

	if *version {
		fmt.Println(py.Version)
		return
	}

	fmt.Fprintf(os.Stdout, "gopython %s\n", py.Version)
	fmt.Fprintln(os.Stdout, py.Status())
}
