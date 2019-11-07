package main

import (
	"flag"
	"io"
	"os"
)

var (
	append = flag.Bool("a", false, "Append to the end of file")
)

func main() {
	flag.Parse()
	fFlag := os.O_APPEND | os.O_CREATE | os.O_WRONLY
	if !*append {
		fFlag |= os.O_TRUNC
	}

	args := flag.Args()
	if len(args) != 1 {
		panic("Required output file")
	}
	f, err := os.OpenFile(args[0], fFlag, 0644)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	out := io.MultiWriter(os.Stdout, f)

	io.Copy(out, os.Stdin)
}
