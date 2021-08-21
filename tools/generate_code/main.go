package main

import (
	"bytes"
	"fmt"
	logger "github.com/virtual-vgo/vvgo/pkg/logger/codegen"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
)

func main() {
	if len(os.Args) != 3 {
		_, _ = fmt.Fprintf(os.Stderr, "usage: %s PACKAGE_NAME OUTPUT_FILE.GO\n", os.Args[0])
		os.Exit(1)
	}
	packageName, outputName := os.Args[1], os.Args[2]

	file, err := os.OpenFile(outputName, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("os.OpenFile() failed: %v", err)
	}
	defer file.Close()

	var generator func(io.Writer) error
	switch packageName {
	case "pkg/logger":
		generator = logger.Generate
	default:
		log.Fatalf("I don't know how to generate for `%s`", packageName)
	}

	var buf bytes.Buffer
	_, _ = fmt.Fprintf(&buf, "// This code was automatically generated.\n// %s\n", strings.Join(os.Args, " "))

	if err := generator(&buf); err != nil {
		log.Fatalf("code generation failed: %v", err)
	}

	cmd := exec.Command("gofmt")
	cmd.Stdin = &buf
	cmd.Stdout = file
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Fatalf("cmd.Run() failed: %v", err)
	}

	fmt.Printf("generated %s/%s\n", packageName, outputName)
}
