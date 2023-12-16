package main

import (
	"log"
	"os"

	"github.com/ninedraft/sulisp/repl"
)

func main() {
	signals := make(chan repl.Signal)

	log.SetFlags(0)

	if err := repl.Run(os.Stdout, os.Stdin, signals); err != nil {
		panic(err)
	}
}
