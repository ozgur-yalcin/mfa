package commands

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

type Commander interface {
	Name() string
	Use() string
	Init(*Ancestor) error
	Args(ctx context.Context, cd *Ancestor, args []string) error
	PreRun(this, runner *Ancestor) error
	Run(ctx context.Context, cd *Ancestor, args []string) error
	Commands() []Commander
}

type Ancestor struct {
	Command   Commander
	Cmd       *cobra.Command
	Root      *Ancestor
	Parent    *Ancestor
	ancestors []*Ancestor
}

type Exec struct {
	c *Ancestor
}

func checkArgs(cmd *cobra.Command, args []string) error {
	if !cmd.HasSubCommands() {
		return nil
	}
	var commandName string
	for _, arg := range args {
		if strings.HasPrefix(arg, "-") {
			break
		}
		commandName = arg
	}
	if commandName == "" || cmd.Name() == commandName {
		return nil
	}
	if cmd.HasAlias(commandName) {
		return nil
	}
	return fmt.Errorf("unknown command %q for %q%s", args[0], cmd.CommandPath(), findSuggestions(cmd, commandName))
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
	for i := len(ancestors) - 1; i >= 0; i-- {
		cd := ancestors[i]
		if err := cd.Command.PreRun(cd, c); err != nil {
			return err
		}
	}
	return nil
}

func (c *Ancestor) compile() error {
	c.Cmd = &cobra.Command{
		Use: c.Command.Use(),
		Args: func(cmd *cobra.Command, args []string) error {
			if err := c.Command.Args(cmd.Context(), c, args); err != nil {
				return err
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := c.Command.Run(cmd.Context(), c, args); err != nil {
				return err
			}
			return nil
		},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return c.init()
		},
		DisableFlagsInUseLine:      true,
		SilenceErrors:              false,
		SilenceUsage:               false,
		SuggestionsMinimumDistance: 2,
	}
	if err := c.Command.Init(c); err != nil {
		return err
	}
	for _, cc := range c.ancestors {
		if err := cc.compile(); err != nil {
			return err
		}
		c.Cmd.AddCommand(cc.Cmd)
	}
	return nil
}

func findSuggestions(cmd *cobra.Command, arg string) string {
	if cmd.DisableSuggestions {
		return ""
	}
	suggestionsString := ""
	if suggestions := cmd.SuggestionsFor(arg); len(suggestions) > 0 {
		suggestionsString += "\n\nDid you mean this?\n"
		for _, s := range suggestions {
			suggestionsString += fmt.Sprintf("\t%v\n", s)
		}
	}
	return suggestionsString
}

func newExec() (*Exec, error) {
	rootCmd := &rootCommand{
		name: "mfa",
		use:  "mfa <subcommand> [flags] [args]",
		commands: []Commander{
			newVersionCommand(),
			newGenerateCommand(),
			newAddCommand(),
			newRemoveCommand(),
			newUpdateCommand(),
			newListCommand(),
		},
	}
	return New(rootCmd)
}
