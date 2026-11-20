//go:build !js || !wasm

package main

import "fmt"

func main() {
	fmt.Println("Build this browser runtime with GOOS=js GOARCH=wasm.")
}
