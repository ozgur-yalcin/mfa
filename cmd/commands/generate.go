package commands

import (
	"context"
	"fmt"

	"github.com/ozgur-yalcin/mfa/otp"
	"github.com/spf13/cobra"
)

type generateCommand struct {
	r           *rootCommand
	name        string
	use         string
	commands    []Commander
	mode        string
	base32      bool
	hash        string
	valueLength int
	counter     int64
	epoch       int64
	interval    int64
}

func (c *generateCommand) Name() string {
	return c.name
}

func (c *generateCommand) Use() string {
	return c.use
}

func (c *generateCommand) Init(cd *Commandeer) error {
	cmd := cd.Cmd
	cmd.Short = "Generate one-time password from secret key"
	cmd.Long = "Generate one-time password from secret key"
	cmd.Flags().StringVarP(&c.mode, "mode", "m", "totp", "use time-variant TOTP mode or use event-based HOTP mode")
	cmd.Flags().BoolVarP(&c.base32, "base32", "b", true, "use base32 encoding of KEY instead of hex")
	cmd.Flags().StringVarP(&c.hash, "hash", "H", "SHA1", "A cryptographic hash method H (SHA1, SHA256, SHA512)")
	cmd.Flags().IntVarP(&c.valueLength, "length", "l", 6, "A HOTP value length d")
	cmd.Flags().Int64VarP(&c.counter, "counter", "c", 0, "used for HOTP, A counter C, which counts the number of iterations")
	cmd.Flags().Int64VarP(&c.epoch, "epoch", "e", 0, "used for TOTP, epoch (T0) which is the Unix time from which to start counting time steps")
	cmd.Flags().Int64VarP(&c.interval, "interval", "i", 30, "used for TOTP, an interval (Tx) which will be used to calculate the value of the counter CT")
	return nil
}

func (c *generateCommand) Args(ctx context.Context, cd *Commandeer, args []string) error {
	if err := cobra.ExactArgs(1)(cd.Cmd, args); err != nil {
		return err
	}
	if c.mode != "hotp" && c.mode != "totp" {
		return fmt.Errorf("mode should be hotp or totp")
	}
	return nil
}

func (c *generateCommand) PreRun(cd, runner *Commandeer) error {
	c.r = cd.Root.Command.(*rootCommand)
	return nil
}

func (c *generateCommand) Run(ctx context.Context, cd *Commandeer, args []string) (err error) {
	if err := cobra.ExactArgs(1)(cd.Cmd, args); err != nil {
		return err
	}
	secretKey := args[0]
	code := ""
	if c.mode == "hotp" {
		hotp := otp.NewHOTP(c.base32, c.hash, c.counter, c.valueLength)
		code, err = hotp.GeneratePassCode(secretKey)
	} else if c.mode == "totp" {
		totp := otp.NewTOTP(c.base32, c.hash, c.valueLength, c.epoch, c.interval)
		code, err = totp.GeneratePassCode(secretKey)
	} else {
		err = fmt.Errorf("mode should be hotp or totp")
	}
	if err != nil {
		fmt.Printf("%s\n", err)
	} else {
		fmt.Println("Code:", code)
	}
	return
}

func (c *generateCommand) Commands() []Commander {
	return c.commands
}

func newGenerateCommand() *generateCommand {
	generateCmd := &generateCommand{
		name: "generate",
		use:  "generate [flags] <secret key>",
	}
	return generateCmd
}
