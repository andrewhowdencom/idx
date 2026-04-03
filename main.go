package main

import (
	"os"

	"github.com/andrewhowdencom/idx/internal/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
