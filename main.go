package main

import (
	"fmt"
	"os"

	"github.com/XotoX1337/GoThrough/cmd"
)

func main() {
	cmd.SetVersion(Version)
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
