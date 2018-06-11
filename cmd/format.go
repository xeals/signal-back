package cmd

import (
	"encoding/csv"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/pkg/errors"
	"github.com/urfave/cli"
	"github.com/xeals/signal-back/types"
)

const csvHeaders string = `<?xml version='1.0' encoding='UTF-8' standalone='yes' ?>
<?xml-stylesheet type="text/xsl" href="sms.xsl"?>
`

// Format fulfils the `format` subcommand.
var Format = cli.Command{
	Name:               "format",
	Usage:              "Read and format the backup file",
	UsageText:          "Parse and transform the backup file into other formats.\nValid formats include: CSV, XML.",
	CustomHelpTemplate: SubcommandHelp,
	Flags: append([]cli.Flag{
		cli.StringFlag{
			Name:  "format, f",
			Usage: "output the backup as `FORMAT`",
			Value: "xml",
		},
		cli.StringFlag{
			Name:  "message, m",
			Usage: "format `TYPE` messages",
			Value: "sms",
		},
		cli.StringFlag{
			Name:  "output, o",
			Usage: "write decrypted format to `FILE`",
		},
	}, coreFlags...),
	Action: func(c *cli.Context) error {
		bf, err := setup(c)
		if err != nil {
			return err
		}

		var out io.Writer
		if c.String("output") != "" {
			var file *os.File
			file, err = os.OpenFile(c.String("output"), os.O_CREATE|os.O_WRONLY, 0644)
			out = io.Writer(file)
			if err != nil {
				return errors.Wrap(err, "unable to open output file")
			}
			defer func() {
				if file.Close() != nil {
					log.Fatalf("unable to close output file: %s", err.Error())
				}
			}()
		} else {
			out = os.Stdout
		}

		defer func() {
			if r := recover(); r != nil {
				fmt.Fprintln(os.Stderr, "Panicked during formatting:", r)
			}
		}()
		switch strings.ToLower(c.String("format")) {
		case "csv":
			err = CSV(bf, strings.ToLower(c.String("message")), out)
		case "xml":
			err = XML(bf, out)
		case "json":
			// err = formatJSON(bf, out)
			return errors.New("JSON is still TODO")
		default:
			return errors.Errorf("format %s not recognised", c.String("format"))
		}
		if err != nil {
			return errors.Wrap(err, "failed to format output")
		}

		return nil
	},
}

// JSON <undefined>
func JSON(bf *types.BackupFile, out io.Writer) error {
	return nil
}

// CSV dumps the raw backup data into a comma-separated value format.
func CSV(bf *types.BackupFile, message string, out io.Writer) error {
	ss := make([][]string, 0)
	for {
		f, err := bf.Frame()
		if err != nil {
			break
		}

		// Attachment needs removing
		if a := f.GetAttachment(); a != nil {
			err := bf.DecryptAttachment(a, ioutil.Discard)
			if err != nil {
				return errors.Wrap(err, "unable to chew through attachment")
			}
		}

		if stmt := f.GetStatement(); stmt != nil {
			if (*stmt.Statement)[:15] == "INSERT INTO "+message {
				ss = append(ss, types.StatementToStringArray(stmt))
			}
		}
	}

	w := csv.NewWriter(out)
	var headers []string
	if message == "mms" {
		headers = types.MMSCSVHeaders
	} else {
		headers = types.SMSCSVHeaders
	}

	if err := w.Write(headers); err != nil {
		return errors.Wrap(err, "unable to write CSV headers")
	}

	for _, sms := range ss {
		if err := w.Write(sms); err != nil {
			return errors.Wrap(err, "unable to format CSV")
		}
	}

	w.Flush()

	return errors.WithMessage(w.Error(), "unable to end CSV writer or something")
}

// XML formats the backup into the same XML format as SMS Backup & Restore
// uses. Layout described at their website
// http://synctech.com.au/fields-in-xml-backup-files/
func XML(bf *types.BackupFile, out io.Writer) error {
	smses := &types.SMSes{}
	mmses := map[uint64]types.MMS{}
	mmsParts := map[uint64][]types.MMSPart{}
	for {
		f, err := bf.Frame()
		if err != nil {
			break
		}

		// Attachment needs removing
		if a := f.GetAttachment(); a != nil {
			err := bf.DecryptAttachment(a, ioutil.Discard)
			if err != nil {
				return errors.Wrap(err, "unable to chew through attachment")
			}
		}

		if stmt := f.GetStatement(); stmt != nil {
			// Only use SMS/MMS statements
			if strings.HasPrefix(*stmt.Statement, "INSERT INTO sms") {
				sms, err := types.NewSMSFromStatement(stmt)
				if err == nil {
					smses.SMS = append(smses.SMS, *sms)
				}
			}

			if strings.HasPrefix(*stmt.Statement, "INSERT INTO mms") {
				// TODO this
				log.Println("MMS export not yet supported")
			}

			if strings.HasPrefix(*stmt.Statement, "INSERT INTO part") {
				// TODO also this
			}
		}
	}

	for id, p := range mmsParts {
		if mms, ok := mmses[id]; ok {
			mms.Parts = p
			smses.MMS = append(smses.MMS, mms)
		}
	}

	// Headers
	if _, err := out.Write([]byte(csvHeaders)); err != nil {
		return errors.Wrap(err, "failed to write out XML")
	}

	smses.Count = len(smses.SMS)

	// x, err := xml.Marshal(smses)
	enc := xml.NewEncoder(out)
	enc.Indent("", "  ")
	return errors.WithMessage(enc.Encode(smses), "unable to format XML")
}
