//go:build ignore

package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "gen: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// build wasm binary
	buildCmd := exec.Command("go", "build", "-o", filepath.Join("static", "main.wasm"), "./wasm/")
	buildCmd.Env = append(os.Environ(), "GOOS=js", "GOARCH=wasm")
	buildCmd.Stdout = os.Stdout
	buildCmd.Stderr = os.Stderr
	if err := buildCmd.Run(); err != nil {
		return fmt.Errorf("go build wasm: %w", err)
	}

	// locate wasm_exec.js in GOROOT
	goroot := runtime.GOROOT()
	candidates := []string{
		filepath.Join(goroot, "lib", "wasm", "wasm_exec.js"),  // Go 1.24+?
		filepath.Join(goroot, "misc", "wasm", "wasm_exec.js"), // older Go
	}
	var src string
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			src = c
			break
		}
	}
	if src == "" {
		return fmt.Errorf("wasm_exec.js not found in GOROOT (%s); tried:\n  %s", goroot, strings.Join(candidates, "\n  "))
	}

	// copy wasm_exec.js
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(filepath.Join("static", "wasm_exec.js"))
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	// Close explicitly; deferred Close will be a harmless no-op on already-closed file.
	return out.Close()
}
