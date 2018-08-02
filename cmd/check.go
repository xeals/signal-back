package cmd

import (
	"io/ioutil"
	"log"
	"os"

	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

// Check fulfils the `format` subcommand.
var Check = cli.Command{
	Name:               "check",
	Usage:              "Verify that a backup is readable",
	UsageText:          "Attempts to decrypt the provided backup and do nothing with it except verify that it's readable\n from start to finish. Enables verbose logging by default.",
	CustomHelpTemplate: SubcommandHelp,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "password, p",
			Usage: "use `PASS` as password for backup file",
		},
		cli.StringFlag{
			Name:  "pwdfile, P",
			Usage: "read password from `FILE`",
		},
	},
	Action: func(c *cli.Context) error {
		bf, err := setup(c)
		if err != nil {
			return err
		}

		log.SetOutput(os.Stderr)

		if err := Raw(bf, ioutil.Discard); err != nil {
			return errors.Wrap(err, "Encountered error while checking")
		}

		log.Println("Backup looks okay from here.")
		return nil
	},
}
