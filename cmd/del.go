package cmd

import (
	"context"
	"errors"
	"flag"
	"log"
	"strings"

	"github.com/ozgur-yalcin/mfa/src/database"
	"github.com/ozgur-yalcin/mfa/src/initialize"
)

type delCommand struct {
	r        *rootCommand
	fs       *flag.FlagSet
	commands []Commander
	name     string
}

func newDelCommand() *delCommand {
	return &delCommand{name: "del"}
}

func (c *delCommand) Name() string {
	return c.name
}

func (c *delCommand) Commands() []Commander {
	return c.commands
}

func (c *delCommand) Init(cd *Ancestor) {
	c.fs = flag.NewFlagSet(c.name, flag.ExitOnError)
}

func (c *delCommand) Run(ctx context.Context, cd *Ancestor, args []string) (err error) {
	initialize.Init()
	if err := c.fs.Parse(args); err != nil {
		return err
	}
	var issuer, user string
	if pairs := strings.SplitN(c.fs.Arg(0), ":", 2); len(pairs) == 2 {
		issuer = pairs[0]
		user = pairs[1]
	} else {
		issuer = c.fs.Arg(0)
	}
	if issuer == "" {
		return errors.New("issuer cannot be empty")
	}
	if err := c.delAccount(issuer, user); err != nil {
		return err
	}
	log.Println("accounts deleted successfully")
	return
}

func (c *delCommand) delAccount(issuer string, user string) (err error) {
	db, err := database.LoadDatabase()
	if err != nil {
		return err
	}
	if err := db.Open(); err != nil {
		return err
	}
	defer db.Close()
	accounts, err := db.ListAccounts(issuer, user)
	if err != nil {
		return err
	}
	if len(accounts) == 0 {
		return errors.New("account not found")
	} else if len(accounts) > 0 {
		return db.DelAccount(issuer, user)
	}
	return
}
