package commands

import (
	"context"
	"flag"
	"log"
	"strings"

	"github.com/ozgur-yalcin/mfa/internal/database"
	"github.com/ozgur-yalcin/mfa/internal/initialize"
)

type removeCommand struct {
	r        *rootCommand
	fs       *flag.FlagSet
	commands []Commander
	name     string
}

func (c *removeCommand) Name() string {
	return c.name
}

func (c *removeCommand) Init(cd *Ancestor) error {
	c.fs = flag.NewFlagSet(c.name, flag.ExitOnError)
	return nil
}

func (c *removeCommand) Run(ctx context.Context, cd *Ancestor, args []string) error {
	initialize.Init()
	if err := c.fs.Parse(args); err != nil {
		log.Fatal(err)
	}
	var accountName, userName string
	if pairs := strings.SplitN(c.fs.Arg(0), ":", 2); len(pairs) == 2 {
		accountName = pairs[0]
		userName = pairs[1]
	} else {
		accountName = c.fs.Arg(0)
	}
	if accountName == "" {
		log.Fatal("account name cannot be empty")
	}
	if err := c.removeAccount(accountName, userName); err != nil {
		log.Fatal(err)
	}
	log.Println("accounts deleted successfully")
	return nil
}

func (c *removeCommand) Commands() []Commander {
	return c.commands
}

func newRemoveCommand() *removeCommand {
	return &removeCommand{name: "remove"}
}

func (c *removeCommand) removeAccount(accountName string, userName string) error {
	db, err := database.LoadDatabase()
	if err != nil {
		log.Fatal(err)
	}
	if err := db.Open(); err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	if err := db.RemoveAccount(accountName, userName); err != nil {
		log.Fatal(err)
	}
	return nil
}
