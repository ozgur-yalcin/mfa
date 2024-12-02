package cmd

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"sync"
	"text/tabwriter"

	"github.com/ozgur-yalcin/mfa/src/database"
	"github.com/ozgur-yalcin/mfa/src/initialize"
	"github.com/ozgur-yalcin/mfa/src/models"
)

type listCommand struct {
	r        *rootCommand
	fs       *flag.FlagSet
	commands []Commander
	name     string
}

func newListCommand() *listCommand {
	return &listCommand{name: "list"}
}

func (c *listCommand) Name() string {
	return c.name
}

func (c *listCommand) Commands() []Commander {
	return c.commands
}

func (c *listCommand) Init(cd *Ancestor) {
	c.fs = flag.NewFlagSet(c.name, flag.ExitOnError)
}

func (c *listCommand) Run(ctx context.Context, cd *Ancestor, args []string) (err error) {
	initialize.Init()
	if err := c.fs.Parse(args); err != nil {
		return err
	}
	var issuer, user string
	if pairs := strings.SplitN(c.fs.Arg(0), ":", 2); len(pairs) == 2 {
		issuer = pairs[0]
		user = pairs[1]
	} else {
		issuer = c.fs.Arg(0)
	}
	if err := c.listAccounts(issuer, user); err != nil {
		return err
	}
	return
}

func (c *listCommand) listAccounts(issuer string, user string) (err error) {
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
	type otp struct {
		issuer string
		user   string
		code   string
	}
	var otps []otp
	if len(accounts) == 0 {
		log.Println("no accounts found!")
	} else {
		var wg sync.WaitGroup
		for _, account := range accounts {
			wg.Add(1)
			go func(account models.Account) {
				defer wg.Done()
				code, err := account.OTP()
				if err != nil {
					log.Printf("%s %s generate code error%s\n", account.Issuer, account.User, err)
				} else {
					otps = append(otps, otp{
						issuer: account.Issuer,
						user:   account.User,
						code:   code,
					})
				}
			}(account)
		}
		wg.Wait()
		sort.Slice(otps, func(i, j int) bool {
			return otps[i].issuer < otps[j].issuer
		})
		writer := tabwriter.NewWriter(os.Stdout, 8, 8, 1, '\t', 0)
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\n", "#", "Issuer", "User", "Code")
		for i, item := range otps {
			_, err := fmt.Fprintf(writer, "%d\t%s\t%s\t%s\n", i+1, item.issuer, item.user, item.code)
			if err != nil {
				log.Printf(err.Error())
			}
		}
		writer.Flush()
	}
	return
}
