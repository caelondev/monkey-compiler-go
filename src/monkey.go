package src

import (
	"flag"
	"fmt"
	"os"

	"github.com/caelondev/monkey-compiler-go/src/build"
	"github.com/caelondev/monkey-compiler-go/src/repl"
	"github.com/caelondev/monkey-compiler-go/src/run"
)

func Main() {
	buildFlag := flag.String("build", "", "compile source file")
	runBCFlag := flag.String("run-bc", "", "run bytecode file")
	disassembleFlag := flag.String("disassemble-bc", "", "disassemble bytecode file")
	flag.Parse()

	args := flag.Args() // remaining positional args

	if *buildFlag != "" {
		build.BuildFile(*buildFlag)
		return
	}

	if *runBCFlag != "" {
		run.RunBytecode(*runBCFlag)
		return
	}

	if *disassembleFlag != "" {
		build.DisassembleFile(*disassembleFlag)
		return
	}

	if len(args) == 0 {
		repl.Start(os.Stdin, os.Stdout)
		return
	}

	if len(args) == 1 {
		run.RunFile(args[0])
		return
	}

	fmt.Println("Usage: monkey [filepath]")
	os.Exit(1)
}
