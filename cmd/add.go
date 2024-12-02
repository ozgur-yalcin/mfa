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
	"github.com/ozgur-yalcin/mfa/src/models"
)

type addCommand struct {
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

func newAddCommand() *addCommand {
	return &addCommand{name: "add"}
}

func (c *addCommand) Name() string {
	return c.name
}

func (c *addCommand) Commands() []Commander {
	return c.commands
}

func (c *addCommand) Init(cd *Ancestor) error {
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

func (c *addCommand) Run(ctx context.Context, cd *Ancestor, args []string) (err error) {
	initialize.Init()
	if err := c.fs.Parse(args); err != nil {
		log.Fatal(err)
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
	if _, err := c.generateCode(secret); err != nil {
		log.Fatal(err)
	}
	if err := c.setAccount(issuer, user, secret); err != nil {
		log.Fatal(err)
	}
	log.Println("account added successfully")
	return nil
}

func (c *addCommand) generateCode(secret string) (code string, err error) {
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

func (c *addCommand) setAccount(issuer string, user string, secret string) error {
	db, err := database.LoadDatabase()
	if err != nil {
		log.Fatal(err)
	}
	if err := db.Open(); err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	accounts, err := db.ListAccounts(issuer, user)
	if err != nil {
		log.Fatal(err)
	}
	if len(accounts) > 0 {
		log.Fatal("account already exists")
	} else if len(accounts) == 0 {
		account := &models.Account{
			Issuer:  issuer,
			User:    user,
			Secret:  secret,
			Mode:    c.mode,
			Hash:    c.hash,
			Digits:  c.digits,
			Period:  c.period,
			Counter: c.counter,
		}
		return db.AddAccount(account)
	}
	return nil
}
