package cmd

import (
	"context"
	"errors"
	"flag"
	"log"
	"strings"

	"github.com/ozgur-yalcin/mfa/otp"
	"github.com/ozgur-yalcin/mfa/src/database"
	"github.com/ozgur-yalcin/mfa/src/initialize"
)

type setCommand struct {
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

func newSetCommand() *setCommand {
	return &setCommand{name: "set"}
}

func (c *setCommand) Name() string {
	return c.name
}

func (c *setCommand) Commands() []Commander {
	return c.commands
}

func (c *setCommand) Init(cd *Ancestor) error {
	c.fs = flag.NewFlagSet(c.name, flag.ExitOnError)
	c.fs.StringVar(&c.mode, "mode", "totp", "use time-variant TOTP mode or use event-based HOTP mode")
	c.fs.StringVar(&c.mode, "m", "totp", "use time-variant TOTP mode or use event-based HOTP mode (shorthand)")
	c.fs.StringVar(&c.hash, "hash", "SHA1", "A cryptographic hash method H")
	c.fs.StringVar(&c.hash, "H", "SHA1", "A cryptographic hash method H (shorthand)")
	c.fs.IntVar(&c.digits, "digits", 6, "A HOTP value digits d")
	c.fs.IntVar(&c.digits, "l", 6, "A HOTP value digits d (shorthand)")
	c.fs.Int64Var(&c.counter, "counter", 0, "used for HOTP, A counter C, which counts the number of iterations")
	c.fs.Int64Var(&c.counter, "c", 0, "used for HOTP, A counter C, which counts the number of iterations (shorthand)")
	c.fs.Int64Var(&c.period, "period", 30, "used for TOTP, an period (Tx) which will be used to calculate the value of the counter CT")
	c.fs.Int64Var(&c.period, "i", 30, "used for TOTP, an period (Tx) which will be used to calculate the value of the counter CT (shorthand)")
	return nil
}

func (c *setCommand) Run(ctx context.Context, cd *Ancestor, args []string) error {
	initialize.Init()
	if err := c.fs.Parse(args); err != nil {
		return err
	}
	var issuer, user, secret string
	if pairs := strings.SplitN(c.fs.Arg(0), ":", 2); len(pairs) == 2 {
		issuer = pairs[0]
		user = pairs[1]
		secret = c.fs.Arg(1)
	} else {
		issuer = c.fs.Arg(0)
		secret = c.fs.Arg(1)
	}
	if issuer == "" {
		return errors.New("issuer cannot be empty")
	}
	if secret == "" {
		return errors.New("secret cannot be empty")
	}
	if _, err := c.generateCode(secret); err != nil {
		return err
	}
	if err := c.setAccount(issuer, user, secret); err != nil {
		return err
	}
	log.Println("account updated successfully")
	return nil
}

func (c *setCommand) generateCode(secret string) (code string, err error) {
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
		return code, err
	}
	return
}

func (c *setCommand) setAccount(issuer string, user string, secret string) error {
	db, err := database.LoadDatabase()
	if err != nil {
		return err
	}
	if err := db.Open(); err != nil {
		return err
	}
	defer db.Close()
	accounts, err := db.ListAccounts(issuer, user)
	if err != nil {
		return err
	}
	if len(accounts) == 0 {
		return errors.New("account not found")
	} else if len(accounts) > 1 {
		return errors.New("multiple accounts found")
	} else if len(accounts) == 1 {
		account := db.GetAccount(issuer, user)
		account.Issuer = issuer
		account.User = user
		account.Secret = secret
		account.Mode = c.mode
		account.Hash = c.hash
		account.Digits = c.digits
		account.Counter = c.counter
		account.Period = c.period
		return db.SetAccount(account)
	}
	return nil
}
