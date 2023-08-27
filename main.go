package main

import (
	"fmt"
	"os"

	"github.com/vigo/statoo/v2/app"
)

func main() {
	cmd := app.NewCLIApplication()
	if err := cmd.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
