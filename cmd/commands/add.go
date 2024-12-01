package commands

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/ozgur-yalcin/mfa/internal/database"
	"github.com/ozgur-yalcin/mfa/internal/initialize"
	"github.com/ozgur-yalcin/mfa/internal/models"
	"github.com/ozgur-yalcin/mfa/otp"
	"github.com/spf13/cobra"
)

type addCommand struct {
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

func (c *addCommand) Name() string {
	return c.name
}

func (c *addCommand) Use() string {
	return c.use
}

func (c *addCommand) Init(cd *Commandeer) error {
	cmd := cd.Cmd
	cmd.Short = "Add account and its secret key"
	cmd.Long = "Add account and its secret key"
	cmd.Flags().StringVarP(&c.mode, "mode", "m", "totp", "use time-variant TOTP mode or use event-based HOTP mode")
	cmd.Flags().BoolVarP(&c.base32, "base32", "b", true, "use base32 encoding of KEY instead of hex")
	cmd.Flags().StringVarP(&c.hash, "hash", "H", "SHA1", "A cryptographic hash method H")
	cmd.Flags().IntVarP(&c.valueLength, "length", "l", 6, "A HOTP value length d")
	cmd.Flags().Int64VarP(&c.counter, "counter", "c", 0, "used for HOTP, A counter C, which counts the number of iterations")
	cmd.Flags().Int64VarP(&c.epoch, "epoch", "e", 0, "used for TOTP, epoch (T0) which is the Unix time from which to start counting time steps")
	cmd.Flags().Int64VarP(&c.interval, "interval", "i", 30, "used for TOTP, an interval (Tx) which will be used to calculate the value of the counter CT")
	return nil
}

func (c *addCommand) Args(ctx context.Context, cd *Commandeer, args []string) error {
	if err := cobra.ExactArgs(2)(cd.Cmd, args); err != nil {
		return err
	}
	if c.mode != "hotp" && c.mode != "totp" {
		return fmt.Errorf("mode should be hotp or totp")
	}
	return nil
}

func (c *addCommand) PreRun(cd, runner *Commandeer) error {
	c.r = cd.Root.Command.(*rootCommand)
	return nil
}

func (c *addCommand) Run(ctx context.Context, cd *Commandeer, args []string) (err error) {
	initialize.Init()
	if err = cobra.ExactArgs(2)(cd.Cmd, args); err != nil {
		return
	}
	accountName := args[0]
	secretKey := args[1]
	var userName string
	if pairs := strings.SplitN(accountName, ":", 2); len(pairs) == 2 {
		accountName = pairs[0]
		userName = pairs[1]
	}
	if _, err = c.generateCode(secretKey); err != nil {
		log.Fatal(err)
	}
	if err = c.saveAccount(accountName, userName, secretKey); err != nil {
		log.Fatal(err)
	}
	if err == nil {
		fmt.Println("account added successfully")
	}
	return err
}

func (c *addCommand) Commands() []Commander {
	return c.commands
}

func newAddCommand() *addCommand {
	addCmd := &addCommand{
		name: "add",
		use:  "add [flags] <account name> <secret key>",
	}
	return addCmd
}

func (c *addCommand) generateCode(secretKey string) (code string, err error) {
	if c.mode == "hotp" {
		hotp := otp.NewHOTP(c.base32, c.hash, c.counter, c.valueLength)
		code, err = hotp.GeneratePassCode(secretKey)
	} else if c.mode == "totp" {
		totp := otp.NewTOTP(c.base32, c.hash, c.valueLength, c.epoch, c.interval)
		code, err = totp.GeneratePassCode(secretKey)
	} else {
		return code, fmt.Errorf("mode should be hotp or totp")
	}
	if err == nil {
		fmt.Println("Code:", code)
	}
	return
}

func (c *addCommand) saveAccount(accountName string, userName string, secretKey string) error {
	db, err := database.LoadDatabase()
	if err != nil {
		return fmt.Errorf("failed to load database: %w", err)
	}
	if err := db.Open(); err != nil {
		log.Fatalf("failed to connect database:%s", err.Error())
	}
	defer db.Close()
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
