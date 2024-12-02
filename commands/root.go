package commands

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/ozgur-yalcin/mfa/internal/initialize"
)

type rootCommand struct {
	fs       *flag.FlagSet
	commands []Commander
	name     string
}

func (r *rootCommand) Name() string {
	return r.name
}

func (r *rootCommand) Init(cd *Ancestor) error {
	r.fs = flag.NewFlagSet(r.name, flag.ExitOnError)
	return nil
}

func (r *rootCommand) Run(ctx context.Context, cd *Ancestor, args []string) error {
	slog.Debug(fmt.Sprintf("mfa version %q finishing with parameters %q", initialize.Version, os.Args))
	return nil
}

func (r *rootCommand) Commands() []Commander {
	return r.commands
}

func (r *Exec) Execute(ctx context.Context, args []string) (*Ancestor, error) {
	if err := r.c.init(); err != nil {
		return nil, err
	}
	cd := r.c
	if len(args) > 0 {
		for _, subcmd := range r.c.ancestors {
			if subcmd.Commander.Name() == args[0] {
				cd = subcmd
				break
			}
		}
	}
	if err := cd.Command.Parse(args); err != nil {
		return cd, err
	}
	if err := cd.Commander.Run(ctx, cd, cd.Command.Args()[1:]); err != nil {
		return cd, err
	}
	return cd, nil
}

func Execute(args []string) error {
	x, err := newExec()
	if err != nil {
		return err
	}
	if _, err := x.Execute(context.Background(), args); err != nil {
		return err
	}
	return err
}

func New(rootCmd Commander) (*Exec, error) {
	root := &Ancestor{
		Commander: rootCmd,
	}
	root.Root = root
	var addCommands func(cd *Ancestor, cmd Commander)
	addCommands = func(cd *Ancestor, cmd Commander) {
		sub := &Ancestor{
			Root:      root,
			Parent:    cd,
			Commander: cmd,
		}
		cd.ancestors = append(cd.ancestors, sub)
		for _, c := range cmd.Commands() {
			addCommands(sub, c)
		}
	}
	for _, c := range rootCmd.Commands() {
		addCommands(root, c)
	}
	if err := root.run(); err != nil {
		return nil, err
	}
	return &Exec{c: root}, nil
}
