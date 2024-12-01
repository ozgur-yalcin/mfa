package main

import (
	"log"
	"os"

	"github.com/ozgur-yalcin/mfa/cmd/commands"
)

func main() {
	log.SetFlags(0)
	log.SetOutput(os.Stdout)
	err := commands.Execute(os.Args[1:])
	if err != nil {
		log.Fatalf("Error: %s", err)
	}
}
