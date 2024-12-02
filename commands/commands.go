package commands

import (
	"context"
	"flag"
)

type Commander interface {
	Name() string
	Init(*Ancestor) error
	Run(ctx context.Context, cd *Ancestor, args []string) error
	Commands() []Commander
}

type Ancestor struct {
	Commander Commander
	Command   *flag.FlagSet
	Root      *Ancestor
	Parent    *Ancestor
	ancestors []*Ancestor
}

type Exec struct {
	c *Ancestor
}

func newExec() (*Exec, error) {
	return New(&rootCommand{
		name: "mfa",
		commands: []Commander{
			newGenCommand(),
			newQrCommand(),
			newAddCommand(),
			newDelCommand(),
			newSetCommand(),
			newListCommand(),
			newVersionCommand(),
		},
	})
}

func (c *Ancestor) init() error {
	var ancestors []*Ancestor
	{
		cd := c
		for cd != nil {
			ancestors = append(ancestors, cd)
			cd = cd.Parent
		}
	}
	return nil
}

func (c *Ancestor) run() error {
	c.Command = flag.NewFlagSet(c.Commander.Name(), flag.ContinueOnError)
	if err := c.Commander.Init(c); err != nil {
		return err
	}
	for _, cc := range c.ancestors {
		if err := cc.run(); err != nil {
			return err
		}
	}
	return nil
}
