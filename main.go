package main

import (
	"os"

	"github.com/caelondev/monkey-compiler-go/src/repl"
)







func main() {
	repl.Start(os.Stdin, os.Stdout)
}
