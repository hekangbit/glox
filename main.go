package main

import (
	"bufio"
	"errors"
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
	case 3:
		if os.Args[1] == "-D" {
			DebugFlag = true
		}
		if err := RunFile(os.Args[2]); err != nil {
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
	return Run(string(src))
}

func Run(source string) error {
	CompilerInit()
	if DebugFlag {
		DumpTokens(source)
	}
	ok, chunk := Compile(source)
	if !ok {
		return errors.New("glox compile fail")
	}
	Interprete(chunk)
	return nil
}
