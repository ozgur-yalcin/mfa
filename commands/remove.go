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

func newRemoveCommand() *removeCommand {
	return &removeCommand{name: "remove"}
}

func (c *removeCommand) Name() string {
	return c.name
}

func (c *removeCommand) Commands() []Commander {
	return c.commands
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
	var issuer, user string
	if pairs := strings.SplitN(c.fs.Arg(0), ":", 2); len(pairs) == 2 {
		issuer = pairs[0]
		user = pairs[1]
	} else {
		issuer = c.fs.Arg(0)
	}
	if issuer == "" {
		log.Fatal("account name cannot be empty")
	}
	if err := c.removeAccount(issuer, user); err != nil {
		log.Fatal(err)
	}
	log.Println("accounts deleted successfully")
	return nil
}

func (c *removeCommand) removeAccount(issuer string, user string) error {
	db, err := database.LoadDatabase()
	if err != nil {
		log.Fatal(err)
	}
	if err := db.Open(); err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	accounts, err := db.ListAccounts(issuer, user)
	if err != nil {
		log.Fatal(err)
	}
	if len(accounts) == 0 {
		log.Fatal("account not found")
	} else if len(accounts) > 0 {
		return db.RemoveAccount(issuer, user)
	}
	return nil
}
