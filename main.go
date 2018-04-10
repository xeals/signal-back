package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

const appHelp = `Usage: {{.HelpName}} [OPTION...] BACKUPFILE

  {{range .VisibleFlags}}{{.}}
  {{end}}
`

var (
	version     = "0.0.0"
	buildCommit string

	pass   string
	logger io.Writer
)

func main() {
	app := cli.NewApp()
	app.HideHelp = true
	app.CustomAppHelpTemplate = appHelp
	app.Version = version
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "password, p",
			Usage: "use `PASS` as password for backup file",
		},
		cli.StringFlag{
			Name:  "pwdfile, P",
			Usage: "read password from `FILE`",
		},
		cli.StringFlag{
			Name:  "format, f",
			Usage: "output the backup as `FORMAT`",
		},
		cli.StringFlag{
			Name:  "output, o",
			Usage: "write decrypted format to `FILE`",
		},
		cli.StringFlag{
			Name:  "log, l",
			Usage: "write logging output to `FILE`",
		},
		cli.BoolFlag{
			Name:  "attachments, a",
			Usage: "extract attachments from the backup",
		},
		cli.StringFlag{
			Name:  "outdir, d",
			Usage: "output attachments to `DIRECTORY`",
		},
		cli.BoolFlag{
			Name:  "help, h",
			Usage: "show help",
		},
	}
	app.Action = cli.ActionFunc(func(c *cli.Context) error {
		// Initialise stuff
		if c.Bool("help") {
			err := cli.ShowAppHelp(c)
			return errors.WithMessage(err, "unable to print help")
		}

		// -- Verify

		file := c.Args().Get(0)
		if file == "" {
			return E(nil, "must specify a Signal backup file", 255)
		}

		if !c.Bool("attachments") && c.String("format") == "" {
			return E(nil, "you must specify either attachments or output format", 255)
		}

		// -- Logging

		if c.String("log") != "" {
			f, err := os.OpenFile(c.String("log"), os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				return E(err, "unable to create logging file", 1)
			}
			logger = f
		} else {
			logger = os.Stderr
		}

		// -- Password

		if c.String("password") != "" {
			pass = c.String("password")
		} else if c.String("pwdfile") != "" {
			bs, err := ioutil.ReadFile(c.String("pwdfile"))
			if err != nil {
				return E(err, "unable to read file", 1)
			}
			pass = string(bs)
		} else {
			r := bufio.NewReader(os.Stdin)
			fmt.Print("Password: ")
			t, err := r.ReadString('\n')
			if err != nil {
				return E(err, "unable to read from stdin", 1)
			}
			pass = t
		}

		bf, err := newBackupFile(file, pass)
		if err != nil {
			return E(err, "failed to open backup file", 1)
		}

		// -- Get to work

		if c.Bool("attachments") {
			if path := c.String("outdir"); path != "" {
				err := os.MkdirAll(path, 0755)
				if err != nil {
					return E(err, "unable to create output directory", 1)
				}
				err = os.Chdir(path)
				if err != nil {
					return E(err, "unable to change working directory", 1)
				}
			}
			if err = extractAttachments(bf); err != nil {
				return E(err, "failed to extract attachment", 1)
			}
		}

		if f := c.String("format"); f != "" {
			var out io.Writer
			if c.String("output") != "" {
				out, err = os.OpenFile(c.String("output"), os.O_CREATE|os.O_WRONLY, 0644)
				if err != nil {
					return E(err, "unable to open output file", 1)
				}
			} else {
				out = os.Stdout
			}
			switch f {
			case "xml":
				err = formatXML(bf, out)
			case "json":
				// err = formatJSON(bf, out)
				return E(nil, "JSON is still TODO", 2)
			default:
				return E(nil, fmt.Sprintf("format %s not recognised\nvalid formats are: xml", f), 1)
			}
			if err != nil {
				return E(err, "failed to format "+f, 1)
			}
		}

		return nil
	})

	_ = app.Run(os.Args)
}

// E is a wrapper to simply create a cli.ExitError.
func E(err error, msg string, code int) *cli.ExitError {
	if err == nil {
		return cli.NewExitError(errors.New(msg), code)
	}
	return cli.NewExitError(errors.Wrap(err, msg), code)
}

func log(s string) {
	if logger == nil {
		return
	}
	if _, err := logger.Write([]byte(s + "\n")); err != nil {
		fmt.Println(errors.Wrap(err, "unable to write to log file").Error())
	}
}
