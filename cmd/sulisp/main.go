package main

import (
	"bufio"
	"bytes"
	"errors"
	"flag"
	"fmt"
	"go/format"
	"go/token"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/ninedraft/sulisp/lexer"
	"github.com/ninedraft/sulisp/parser"
	"github.com/ninedraft/sulisp/translator"
)

func main() {
	flag.Usage = usage
	verbose := flag.Bool("v", false, "verbose output")
	output := flag.String("o", "", "output executable file")
	run := flag.Bool("run", false, "run compiled executable")
	printAnnotated := flag.Bool("annotated", false, "print generated AST")
	flag.Parse()

	if !*verbose {
		log.SetOutput(io.Discard)
	}

	if flag.NArg() == 0 {
		flag.Usage()
		return
	}

	input := flag.Arg(0)

	inputFile, errInput := os.Open(input)
	if errInput != nil {
		panic("opening input file: " + errInput.Error())
	}
	defer inputFile.Close()

	lex := &lexer.Lexer{
		File:  input,
		Input: bufio.NewReader(inputFile),
	}

	log.Println("lexing input")
	if errLex := lex.Run(); errLex != nil {
		panic("lexing input: " + errLex.Error())
	}

	pr := &parser.Parser{Tokens: lex.Tokens}

	log.Println("parsing input")
	root, errParse := pr.Parse()
	if errParse != nil {
		panic("parsing input: " + errParse.Error())
	}

	log.Println("generating go code")
	ast, errTranslate := translator.TranslateFile(root)
	if errTranslate != nil {
		panic("translating input: " + errTranslate.Error())
	}

	fset := token.NewFileSet()
	formatted := &bytes.Buffer{}
	errFormat := format.Node(formatted, fset, ast)
	if errFormat != nil {
		panic("formatting translated code: " + errFormat.Error())
	}

	if *printAnnotated {
		fmt.Println(formatted)
		return
	}

	wd, errWd := os.MkdirTemp("", "sulisp")
	if errWd != nil {
		panic("creating temp dir: " + errWd.Error())
	}
	defer os.RemoveAll(wd)

	log.Println("generating go package in", wd)
	if err := writeProject(wd, formatted.Bytes()); err != nil {
		panic("writing project: " + err.Error())
	}

	if *output == "" {
		*output = binOutput(wd)
	}

	outputAbs, errAbs := filepath.Abs(*output)
	if errAbs != nil {
		panic("getting absolute output path: " + errAbs.Error())
	}

	log.Println("building executable in", outputAbs, "to", outputAbs)

	errBuild := build(wd, outputAbs)
	if errBuild != nil {
		panic("building executable: " + errBuild.Error())
	}

	if *run {
		errRun := runBin(outputAbs)
		if errRun != nil {
			panic("running executable: " + errRun.Error())
		}
	}
}

func writeProject(wd string, generated []byte) error {
	return errors.Join(
		os.WriteFile(filepath.Join(wd, "generated.go"), generated, 0600),
		os.WriteFile(filepath.Join(wd, "go.mod"), []byte("module main\n"), 0600),
	)
}

func binOutput(wd string) string {
	name := filepath.Base(wd)
	if runtime.GOOS == "windows" {
		name += ".exe"
	}

	return filepath.Join(wd, name)
}

func runBin(binary string) error {
	cmd := exec.Command(binary)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Env = append(cmd.Env, os.Environ()...)
	return cmd.Run()
}

func build(wd, output string) error {
	cmd := exec.Command("go", "build")
	if output != "" {
		cmd.Args = append(cmd.Args, "-o", output)
	}
	cmd.Dir = wd
	cmd.Args = append(cmd.Args, wd)
	cmd.Env = append(cmd.Env, os.Environ()...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	return cmd.Run()
}

func usage() {
	println := func(args ...any) {
		fmt.Fprintln(flag.CommandLine.Output(), args...)
	}
	println("Usage: sulisp [flags] <input file>")
	flag.PrintDefaults()
}
