package main

import (
	"log"
	"os"

	"github.com/ozgur-yalcin/mfa/cmd"
)

func main() {
	log.SetFlags(0)
	log.SetOutput(os.Stdout)
	err := cmd.Execute(os.Args[1:])
	if err != nil {
		log.Fatal(err)
	}
}
