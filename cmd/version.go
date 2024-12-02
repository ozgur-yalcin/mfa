package cmd

import (
	"context"
	"flag"
	"log"

	"github.com/ozgur-yalcin/mfa/src/initialize"
)

type versionCommand struct {
	r        *rootCommand
	fs       *flag.FlagSet
	commands []Commander
	name     string
}

func newVersionCommand() *versionCommand {
	return &versionCommand{name: "version"}
}

func (c *versionCommand) Name() string {
	return c.name
}

func (c *versionCommand) Commands() []Commander {
	return c.commands
}

func (c *versionCommand) Init(cd *Ancestor) (err error) {
	c.fs = flag.NewFlagSet(c.name, flag.ExitOnError)
	return
}

func (c *versionCommand) Run(ctx context.Context, cd *Ancestor, args []string) (err error) {
	c.ShowVersion()
	return
}

func (c *versionCommand) ShowVersion() {
	log.Printf("mfa %s\n", initialize.Version)
}
