package main

import (
	"io"
	"os"
)

func main() {
	if len(os.Args) != 2 {
		panic("Required output file")
	}

	f, err := os.OpenFile(os.Args[1], os.O_APPEND|os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	out := io.MultiWriter(os.Stdout, f)

	io.Copy(out, os.Stdin)
}
