package main

import (
	"os"

	"harvest/internal/commands"
)

func main() {
	os.Exit(commands.Run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}
