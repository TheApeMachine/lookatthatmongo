package main

import (
	"github.com/charmbracelet/log"
	"github.com/theapemachine/lookatthatmongo/cmd"
)

/*
main is the entry point of the application. It executes the root command
and handles any fatal errors that might occur during execution.
*/
func main() {
	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
