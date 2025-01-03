package cmd

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"html"
	"image"
	"log"
	"net/url"
	"os"
	"strings"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/tiff"
	_ "golang.org/x/image/webp"

	"github.com/ozgur-yalcin/mfa/lib"
	"github.com/ozgur-yalcin/mfa/lib/qrcode"
	"github.com/ozgur-yalcin/mfa/src/database"
	"github.com/ozgur-yalcin/mfa/src/initialize"
	"github.com/ozgur-yalcin/mfa/src/models"
)

type qrCommand struct {
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

func newQrCommand() *qrCommand {
	return &qrCommand{name: "qr"}
}

func (c *qrCommand) Name() string {
	return c.name
}

func (c *qrCommand) Commands() []Commander {
	return c.commands
}

func (c *qrCommand) Init(cd *Ancestor) {
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
}

func (c *qrCommand) Run(ctx context.Context, cd *Ancestor, args []string) (err error) {
	initialize.Init()
	if err := c.fs.Parse(args); err != nil {
		return err
	}
	qr, err := c.readQRCode(c.fs.Arg(0))
	if err != nil {
		return err
	}
	u, err := url.Parse(html.UnescapeString(qr.String()))
	if err != nil {
		return err
	}
	if u.Scheme != "otpauth" {
		return errors.New("invalid scheme")
	}
	account := &models.Account{
		Mode:    c.mode,
		Hash:    c.hash,
		Digits:  c.digits,
		Period:  c.period,
		Counter: c.counter,
	}
	if host := u.Hostname(); host != "" {
		account.Mode = host
	}
	if path := strings.TrimPrefix(u.Path, "/"); path != "" {
		if pairs := strings.SplitN(path, ":", 2); len(pairs) == 2 {
			account.Issuer = pairs[0]
			account.User = pairs[1]
		} else {
			account.Issuer = path
		}
	}
	if issuer := u.Query().Get("issuer"); issuer != "" {
		account.Issuer = issuer
	}
	if secret := u.Query().Get("secret"); secret != "" {
		account.Secret = secret
	}
	if hash := u.Query().Get("algorithm"); hash != "" {
		account.Hash = hash
	}
	if digits := u.Query().Get("digits"); digits != "" {
		fmt.Sscanf(digits, "%d", &account.Digits)
	}
	if period := u.Query().Get("period"); period != "" && account.Mode == "totp" {
		fmt.Sscanf(period, "%d", &account.Period)
	}
	if counter := u.Query().Get("counter"); counter != "" && account.Mode == "hotp" {
		fmt.Sscanf(counter, "%d", &account.Counter)
	}
	if err := c.addAccount(account); err != nil {
		return err
	}
	log.Println("account added successfully")
	return
}

func (c *qrCommand) readQRCode(path string) (*lib.Result, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}
	source := lib.NewLuminanceSourceFromImage(img)
	binary := lib.NewHybridBinarizer(source)
	bitmap, err := lib.NewBinaryBitmap(binary)
	if err != nil {
		return nil, err
	}
	reader := qrcode.NewQRCodeReader()
	return reader.Decode(bitmap, nil)
}

func (c *qrCommand) addAccount(account *models.Account) (err error) {
	db, err := database.LoadDatabase()
	if err != nil {
		return err
	}
	if err := db.Open(); err != nil {
		return err
	}
	defer db.Close()
	accounts, err := db.ListAccounts(account.Issuer, account.User)
	if err != nil {
		return err
	}
	if len(accounts) > 0 {
		return errors.New("account already exists")
	} else if len(accounts) == 0 {
		return db.AddAccount(account)
	}
	return
}
