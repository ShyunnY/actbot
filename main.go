package main

import (
	"fmt"
	"os"

	"github.com/ShyunnY/actbot/internal"
)

func main() {
	if err := internal.Setup(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
	}
}
