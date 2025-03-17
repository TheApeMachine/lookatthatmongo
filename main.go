package main

import (
	"github.com/charmbracelet/log"
	"github.com/theapemachine/lookatthatmongo/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
