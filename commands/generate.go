package commands

import (
	"context"
	"errors"
	"flag"
	"log"

	"github.com/ozgur-yalcin/mfa/otp"
)

type generateCommand struct {
	r           *rootCommand
	name        string
	use         string
	commands    []Commander
	fs          *flag.FlagSet
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

func (c *generateCommand) Init(cd *Ancestor) error {
	c.fs = flag.NewFlagSet(c.name, flag.ExitOnError)
	c.fs.StringVar(&c.mode, "mode", "totp", "use time-variant TOTP mode or use event-based HOTP mode")
	c.fs.StringVar(&c.mode, "m", "totp", "use time-variant TOTP mode or use event-based HOTP mode (shorthand)")
	c.fs.BoolVar(&c.base32, "base32", true, "use base32 encoding of KEY instead of hex")
	c.fs.BoolVar(&c.base32, "b", true, "use base32 encoding of KEY instead of hex (shorthand)")
	c.fs.StringVar(&c.hash, "hash", "SHA1", "A cryptographic hash method H (SHA1, SHA256, SHA512)")
	c.fs.StringVar(&c.hash, "H", "SHA1", "A cryptographic hash method H (SHA1, SHA256, SHA512) (shorthand)")
	c.fs.IntVar(&c.valueLength, "length", 6, "A HOTP value length d")
	c.fs.IntVar(&c.valueLength, "l", 6, "A HOTP value length d (shorthand)")
	c.fs.Int64Var(&c.counter, "counter", 0, "used for HOTP, A counter C, which counts the number of iterations")
	c.fs.Int64Var(&c.counter, "c", 0, "used for HOTP, A counter C, which counts the number of iterations (shorthand)")
	c.fs.Int64Var(&c.epoch, "epoch", 0, "used for TOTP, epoch (T0) which is the Unix time from which to start counting time steps")
	c.fs.Int64Var(&c.epoch, "e", 0, "used for TOTP, epoch (T0) which is the Unix time from which to start counting time steps (shorthand)")
	c.fs.Int64Var(&c.interval, "interval", 30, "used for TOTP, an interval (Tx) which will be used to calculate the value of the counter CT")
	c.fs.Int64Var(&c.interval, "i", 30, "used for TOTP, an interval (Tx) which will be used to calculate the value of the counter CT (shorthand)")
	return nil
}

func (c *generateCommand) Run(ctx context.Context, cd *Ancestor, args []string) (err error) {
	if err := c.fs.Parse(args); err != nil {
		return err
	}
	secretKey := c.fs.Arg(0)
	code, err := c.generateCode(secretKey)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Code:", code)
	return
}

func (c *generateCommand) generateCode(secretKey string) (code string, err error) {
	if c.mode == "hotp" {
		hotp := otp.NewHOTP(c.base32, c.hash, c.counter, c.valueLength)
		code, err = hotp.GeneratePassCode(secretKey)
	} else if c.mode == "totp" {
		totp := otp.NewTOTP(c.base32, c.hash, c.valueLength, c.epoch, c.interval)
		code, err = totp.GeneratePassCode(secretKey)
	} else {
		return code, errors.New("mode should be hotp or totp")
	}
	if err != nil {
		log.Fatal(err)
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
