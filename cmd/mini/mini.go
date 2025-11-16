package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/ninedraft/sulisp/compiler"
	"github.com/ninedraft/sulisp/interpreter/bytecode"
	"github.com/ninedraft/sulisp/language/object"
	"github.com/ninedraft/sulisp/lexer"
	"github.com/ninedraft/sulisp/parser"
)

func main() {
	flag.Parse()

	filename := flag.Arg(0)
	if filename == "" {
		panic("need input source code file to run")
	}

	file, err := os.Open(filename)
	if err != nil {
		panic("opening source code file: " + err.Error())
	}
	defer file.Close()

	lexer := lexer.NewLexer(filename, bufio.NewReader(file))
	pkg, err := parser.New(lexer).Parse()
	if err != nil {
		panic("parsing source code: " + err.Error())
	}

	comp := compiler.New()
	comp.HandleEffect("io", "println", 1)

	tape, err := comp.Compile(pkg)
	if err != nil {
		panic("compile: " + err.Error())
	}

	vm := bytecode.NewVM(tape)

	vm.RegisterHostEffect("io", "println",
		func(ctx *bytecode.CallCtx, args []object.Object) {
			strs := make([]string, 0, len(args))

			for _, arg := range args {
				strs = append(strs, arg.Inspect())
			}

			fmt.Println(strings.Join(strs, " "))

			ctx.Invoke()
		})

	vm.Run()

	if vm.Err != nil {
		panic("running: " + vm.Err.Error())
	}
}
