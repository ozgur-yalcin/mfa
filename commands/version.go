package commands

import (
	"context"
	"flag"
	"log"

	"github.com/ozgur-yalcin/mfa/internal/initialize"
)

type versionCommand struct {
	r        *rootCommand
	name     string
	use      string
	commands []Commander
	fs       *flag.FlagSet
}

func (c *versionCommand) Name() string {
	return c.name
}

func (c *versionCommand) Use() string {
	return c.use
}

func (c *versionCommand) Init(cd *Ancestor) error {
	c.fs = flag.NewFlagSet(c.name, flag.ExitOnError)
	return nil
}

func (c *versionCommand) Run(ctx context.Context, cd *Ancestor, args []string) error {
	c.ShowVersion()
	return nil
}

func (c *versionCommand) Commands() []Commander {
	return c.commands
}

func newVersionCommand() *versionCommand {
	return &versionCommand{
		name: "version",
		use:  "version",
	}
}

func (c *versionCommand) ShowVersion() {
	log.Printf("mfa %s\n", initialize.Version)
}
