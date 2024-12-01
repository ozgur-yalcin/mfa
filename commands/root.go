package commands

import (
	"context"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"

	"github.com/ozgur-yalcin/mfa/internal/initialize"
	"github.com/spf13/cobra"
)

type rootCommand struct {
	Printf   func(format string, v ...interface{})
	Println  func(a ...interface{})
	Out      io.Writer
	logger   log.Logger
	name     string
	use      string
	commands []Commander
}

func (r *rootCommand) Name() string {
	return r.name
}

func (r *rootCommand) Use() string {
	return r.use
}

func (r *rootCommand) Init(cd *Ancestor) error {
	cmd := cd.Command
	cmd.Use = "mfa [flags]"
	cmd.Short = "mfa generate OTP"
	cmd.Long = "mfa is a command line tool for generating and validating one-time password."
	return nil
}

func (r *rootCommand) Args(ctx context.Context, cd *Ancestor, args []string) error {
	return nil
}

func (r *rootCommand) PreRun(cd, runner *Ancestor) error {
	return nil
}

func (r *rootCommand) Run(ctx context.Context, cd *Ancestor, args []string) error {
	slog.Debug(fmt.Sprintf("mfa version %q finishing with parameters %q", initialize.Version, os.Args))
	return cd.Command.Usage()
}

func (r *rootCommand) Commands() []Commander {
	return r.commands
}

func (r *Exec) Execute(ctx context.Context, args []string) (*Ancestor, error) {
	if args == nil {
		args = []string{}
	}
	r.c.Command.SetArgs(args)
	cobraCommand, err := r.c.Command.ExecuteContextC(ctx)
	var cd *Ancestor
	if cobraCommand != nil {
		if err == nil {
			err = checkArgs(cobraCommand, args)
		}

		// Find the ancestor that was executed.
		var find func(*cobra.Command, *Ancestor) *Ancestor
		find = func(what *cobra.Command, in *Ancestor) *Ancestor {
			if in.Command == what {
				return in
			}
			for _, in2 := range in.ancestors {
				if found := find(what, in2); found != nil {
					return found
				}
			}
			return nil
		}
		cd = find(cobraCommand, r.c)
	}

	return cd, err
}

func Execute(args []string) error {
	x, err := newExec()
	if err != nil {
		return err
	}
	cd, err := x.Execute(context.Background(), args)
	if err != nil {
		if err == errHelp {
			cd.Command.Help()
			fmt.Println()
			return nil
		}
		if IsCommandError(err) {
			cd.Command.Help()
			fmt.Println()
		}
	}
	return err
}

func New(rootCmd Commander) (*Exec, error) {
	rootCd := &Ancestor{
		Commander: rootCmd,
	}
	rootCd.Root = rootCd
	var addCommands func(cd *Ancestor, cmd Commander)
	addCommands = func(cd *Ancestor, cmd Commander) {
		cd2 := &Ancestor{
			Root:      rootCd,
			Parent:    cd,
			Commander: cmd,
		}
		cd.ancestors = append(cd.ancestors, cd2)
		for _, c := range cmd.Commands() {
			addCommands(cd2, c)
		}

	}
	for _, cmd := range rootCmd.Commands() {
		addCommands(rootCd, cmd)
	}
	if err := rootCd.compile(); err != nil {
		return nil, err
	}
	return &Exec{c: rootCd}, nil
}
