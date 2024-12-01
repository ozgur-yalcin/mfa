package commands

import (
	"context"

	"github.com/spf13/cobra"
)

type simpleCommand struct {
	use      string
	name     string
	short    string
	long     string
	args     func(ctx context.Context, cd *Ancestor, rootCmd *rootCommand, args []string) error
	run      func(ctx context.Context, cd *Ancestor, rootCmd *rootCommand, args []string) error
	withc    func(cmd *cobra.Command, r *rootCommand)
	initc    func(cd *Ancestor) error
	commands []Commander
	rootCmd  *rootCommand
}

func (c *simpleCommand) Name() string {
	return c.name
}

func (c *simpleCommand) Use() string {
	return c.use
}

func (c *simpleCommand) Init(cd *Ancestor) error {
	c.rootCmd = cd.Root.Commander.(*rootCommand)
	cmd := cd.Command
	cmd.Short = c.short
	cmd.Long = c.long
	if c.use != "" {
		cmd.Use = c.use
	}
	if c.withc != nil {
		c.withc(cmd, c.rootCmd)
	}
	return nil
}

func (c *simpleCommand) Args(ctx context.Context, cd *Ancestor, args []string) error {
	if c.args == nil {
		return nil
	}
	return c.args(ctx, cd, c.rootCmd, args)
}

func (c *simpleCommand) PreRun(cd, runner *Ancestor) error {
	if c.initc != nil {
		return c.initc(cd)
	}
	return nil
}

func (c *simpleCommand) Run(ctx context.Context, cd *Ancestor, args []string) error {
	if c.run == nil {
		return nil
	}
	return c.run(ctx, cd, c.rootCmd, args)
}

func (c *simpleCommand) Commands() []Commander {
	return c.commands
}
