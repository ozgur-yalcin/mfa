package commands

import (
	"context"
	"errors"
	"flag"
	"log"
	"strings"

	"github.com/ozgur-yalcin/mfa/internal/database"
	"github.com/ozgur-yalcin/mfa/internal/initialize"
	"github.com/ozgur-yalcin/mfa/internal/models"
	"github.com/ozgur-yalcin/mfa/otp"
)

type addCommand struct {
	r           *rootCommand
	fs          *flag.FlagSet
	commands    []Commander
	name        string
	mode        string
	base32      bool
	hash        string
	valueLength int
	counter     int64
	epoch       int64
	interval    int64
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
	c.fs.BoolVar(&c.base32, "base32", true, "use base32 encoding of KEY instead of hex")
	c.fs.BoolVar(&c.base32, "b", true, "use base32 encoding of KEY instead of hex (shorthand)")
	c.fs.StringVar(&c.hash, "hash", "SHA1", "A cryptographic hash method H")
	c.fs.StringVar(&c.hash, "H", "SHA1", "A cryptographic hash method H (shorthand)")
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

func (c *addCommand) Run(ctx context.Context, cd *Ancestor, args []string) (err error) {
	initialize.Init()
	if err := c.fs.Parse(args); err != nil {
		log.Fatal(err)
	}
	var accountName, userName, secretKey string
	if pairs := strings.SplitN(c.fs.Arg(0), ":", 2); len(pairs) == 2 {
		accountName = pairs[0]
		userName = pairs[1]
		secretKey = c.fs.Arg(1)
	} else {
		accountName = c.fs.Arg(0)
		secretKey = c.fs.Arg(1)
	}
	if _, err := c.generateCode(secretKey); err != nil {
		log.Fatal(err)
	}
	if err := c.saveAccount(accountName, userName, secretKey); err != nil {
		log.Fatal(err)
	}
	log.Println("account added successfully")
	return nil
}

func (c *addCommand) generateCode(secretKey string) (code string, err error) {
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

func (c *addCommand) saveAccount(accountName string, userName string, secretKey string) error {
	db, err := database.LoadDatabase()
	if err != nil {
		log.Fatal(err)
	}
	if err := db.Open(); err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	accounts, err := db.ListAccounts(accountName, userName)
	if err != nil {
		log.Fatal(err)
	}
	if len(accounts) > 0 {
		log.Fatal("account already exists")
	} else if len(accounts) == 0 {
		account := &models.Account{
			AccountName: accountName,
			Username:    userName,
			SecretKey:   secretKey,
			Mode:        c.mode,
			Base32:      c.base32,
			Hash:        c.hash,
			ValueLength: c.valueLength,
			Counter:     c.counter,
			Epoch:       c.epoch,
			Interval:    c.interval,
		}
		return db.CreateAccount(account)
	}
	return nil
}
