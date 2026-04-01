// Package main prints the current UTC time in ISO 8601 format.
// Used by Makefile for cross-platform date generation (replaces Unix date -u).
package main

import (
	"fmt"
	"time"
)

func main() {
	fmt.Print(time.Now().UTC().Format(time.RFC3339))
}
