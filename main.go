package main

import (
	"bufio"
	"fmt"
	"os"
)

func Init() {
	ScannerInit()
}

func main() {
	Init()
	switch len(os.Args) {
	case 1:
		Repl()
	case 2:
		if err := RunFile(os.Args[1]); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(64)
		}
	default:
		fmt.Fprintln(os.Stderr, "Usage: glox [path]")
		os.Exit(64)
	}
}

func Repl() {
	in := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("> ")
		if !in.Scan() {
			break
		}
		line := in.Text()
		// TODO: interpreter line
		fmt.Println("got:", line)
	}
}

func RunFile(path string) error {
	src, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return run(string(src))
}

func run(source string) error {
	CompilerInit()
	DumpTokens(source)
	ok, chunk := Compile(source)
	fmt.Printf("Compile success: %v\n", ok)
	if ok {
		DisassembleChunk(chunk, "test chunk")
	}
	return nil
}
