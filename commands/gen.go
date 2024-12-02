package commands

import (
	"context"
	"errors"
	"flag"
	"log"

	"github.com/ozgur-yalcin/mfa/otp"
)

type genCommand struct {
	r        *rootCommand
	fs       *flag.FlagSet
	commands []Commander
	name     string
	mode     string
	hash     string
	digits   int
	period   int64
	counter  int64
}

func newGenCommand() *genCommand {
	return &genCommand{name: "gen"}
}

func (c *genCommand) Name() string {
	return c.name
}

func (c *genCommand) Commands() []Commander {
	return c.commands
}

func (c *genCommand) Init(cd *Ancestor) error {
	c.fs = flag.NewFlagSet(c.name, flag.ExitOnError)
	c.fs.StringVar(&c.mode, "mode", "totp", "use time-variant TOTP mode or use event-based HOTP mode")
	c.fs.StringVar(&c.mode, "m", "totp", "use time-variant TOTP mode or use event-based HOTP mode (shorthand)")
	c.fs.StringVar(&c.hash, "hash", "SHA1", "A cryptographic hash method H (SHA1, SHA256, SHA512)")
	c.fs.StringVar(&c.hash, "H", "SHA1", "A cryptographic hash method H (SHA1, SHA256, SHA512) (shorthand)")
	c.fs.IntVar(&c.digits, "digits", 6, "A HOTP value digits d")
	c.fs.IntVar(&c.digits, "l", 6, "A HOTP value digits d (shorthand)")
	c.fs.Int64Var(&c.counter, "counter", 0, "used for HOTP, A counter C, which counts the number of iterations")
	c.fs.Int64Var(&c.counter, "c", 0, "used for HOTP, A counter C, which counts the number of iterations (shorthand)")
	c.fs.Int64Var(&c.period, "period", 30, "used for TOTP, an period (Tx) which will be used to calculate the value of the counter CT")
	c.fs.Int64Var(&c.period, "i", 30, "used for TOTP, an period (Tx) which will be used to calculate the value of the counter CT (shorthand)")
	return nil
}

func (c *genCommand) Run(ctx context.Context, cd *Ancestor, args []string) (err error) {
	if err := c.fs.Parse(args); err != nil {
		return err
	}
	secret := c.fs.Arg(0)
	code, err := c.generateCode(secret)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Code:", code)
	return
}

func (c *genCommand) generateCode(secret string) (code string, err error) {
	if c.mode == "hotp" {
		hotp := otp.NewHOTP(c.hash, c.digits, c.counter)
		code, err = hotp.GeneratePassCode(secret)
	} else if c.mode == "totp" {
		totp := otp.NewTOTP(c.hash, c.digits, c.period)
		code, err = totp.GeneratePassCode(secret)
	} else {
		return code, errors.New("mode should be hotp or totp")
	}
	if err != nil {
		log.Fatal(err)
	}
	return
}
