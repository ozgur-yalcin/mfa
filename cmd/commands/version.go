package commands

import (
	"context"
	"fmt"

	"github.com/ozgur-yalcin/mfa/internal/initialize"
)

type versionCommand struct {
	r        *rootCommand
	name     string
	use      string
	commands []Commander
}

func (c *versionCommand) Name() string {
	return c.name
}

func (c *versionCommand) Use() string {
	return c.use
}

func (c *versionCommand) Init(cd *Ancestor) error {
	cmd := cd.Command
	cmd.Short = "show version"
	cmd.Long = "show version"
	return nil
}

func (c *versionCommand) Args(ctx context.Context, cd *Ancestor, args []string) error {
	return nil
}

func (c *versionCommand) PreRun(cd, runner *Ancestor) error {
	c.r = cd.Root.Commander.(*rootCommand)
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
	versionCmd := &versionCommand{
		name: "version",
		use:  "version",
	}
	return versionCmd
}

func (c *versionCommand) ShowVersion() {
	fmt.Printf("mfa %s\n", initialize.Version)
}
