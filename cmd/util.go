package cmd

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"syscall"

	"github.com/pkg/errors"
	"github.com/urfave/cli"
	"github.com/xeals/signal-back/types"
	"golang.org/x/crypto/ssh/terminal"
)

// AppHelp is the help template.
const AppHelp = `Usage: {{.HelpName}} COMMAND [OPTION...] BACKUPFILE

  {{range .Flags}}{{.}}
  {{end}}{{if .Commands}}
Commands:
{{range .Commands}}  {{index .Names 0}}{{ "\t"}}{{.Usage}}
{{end}}{{end}}
`

// TODO: Work out how to display global flags here
// SubcommandHelp is the subcommand help template.
const SubcommandHelp = `Usage: {{.HelpName}} [OPTION...] BACKUPFILE

{{if .UsageText}}{{.UsageText}}
{{else}}{{.Usage}}
{{end}}{{if .Flags}}
  {{range .Flags}}{{.}}
  {{end}}{{end}}
`

var coreFlags = []cli.Flag{
	cli.StringFlag{
		Name:  "password, p",
		Usage: "use `PASS` as password for backup file",
	},
	cli.StringFlag{
		Name:  "pwdfile, P",
		Usage: "read password from `FILE`",
	},
	cli.BoolFlag{
		Name:  "verbose, v",
		Usage: "enable verbose logging output",
	},
}

func setup(c *cli.Context) (*types.BackupFile, error) {
	// -- Enable logging

	if c.Bool("verbose") {
		log.SetOutput(os.Stderr)
	} else {
		log.SetOutput(ioutil.Discard)
	}

	// -- Verify

	if c.Args().Get(0) == "" {
		return nil, errors.New("must specify a Signal backup file")
	}

	// -- Initialise

	pass, err := readPassword(c)
	if err != nil {
		return nil, errors.Wrap(err, "unable to read password")
	}

	bf, err := types.NewBackupFile(c.Args().Get(0), pass)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open backup file")
	}

	return bf, nil
}

func readPassword(c *cli.Context) (string, error) {
	var pass string

	if c.String("password") != "" {
		pass = c.String("password")
	} else if c.String("pwdfile") != "" {
		bs, err := ioutil.ReadFile(c.String("pwdfile"))
		if err != nil {
			return "", errors.Wrap(err, "unable to read file")
		}
		pass = string(bs)
	} else {
		// Read from stdin
		fmt.Fprint(os.Stderr, "Password: ")
		raw, err := terminal.ReadPassword(int(syscall.Stdin))
		if err != nil {
			return "", errors.Wrap(err, "unable to read from stdin")
		}
		pass = string(raw)
	}
	return pass, nil
}

// E is a wrapper to simply create a cli.ExitError.
func E(err error, msg string, code int) *cli.ExitError {
	if err == nil {
		return cli.NewExitError(errors.New(msg), code)
	}
	return cli.NewExitError(errors.Wrap(err, msg), code)
}
