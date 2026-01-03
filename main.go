//go:build !(js && wasm)
// +build !js !wasm

package main

import "github.com/cwbudde/go-sq-tool/cmd"

func main() {
	cmd.Execute()
}
