package cmd

import (
	"encoding/csv"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/urfave/cli"
	"github.com/xeals/signal-back/types"
)

// Format fulfils the `format` subcommand.
var Format = cli.Command{
	Name:               "format",
	Usage:              "Read and format the backup file",
	UsageText:          "Parse and transform the backup file into other formats.\nValid formats include: CSV, XML.",
	CustomHelpTemplate: SubcommandHelp,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "format, f",
			Usage: "output the backup as `FORMAT`",
			Value: "xml",
		},
		cli.StringFlag{
			Name:  "log, l",
			Usage: "write logging output to `FILE`",
		},
		cli.StringFlag{
			Name:  "message, m",
			Usage: "format `TYPE` messages",
			Value: "sms",
		},
		cli.StringFlag{
			Name:  "password, p",
			Usage: "use `PASS` as password for backup file",
		},
		cli.StringFlag{
			Name:  "pwdfile, P",
			Usage: "read password from `FILE`",
		},
		cli.StringFlag{
			Name:  "output, o",
			Usage: "write decrypted format to `FILE`",
		},
	},
	Action: func(c *cli.Context) error {
		bf, err := setup(c)
		if err != nil {
			return err
		}

		var out io.Writer
		if c.String("output") != "" {
			out, err = os.OpenFile(c.String("output"), os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				return errors.Wrap(err, "unable to open output file")
			}
		} else {
			out = os.Stdout
		}

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

				ps := stmt.Parameters
				unix := time.Unix(int64(*ps[5].IntegerParameter)/1000, 0)
				readable := unix.Format("Jan 02, 2006 3:04:05 PM")

				sms := types.SMS{
					Protocol:      ps[7].IntegerParameter,
					Address:       ps[2].GetStringParamter(),
					Date:          strconv.FormatUint(ps[5].GetIntegerParameter(), 10),
					Type:          translateSMSType(ps[10].GetIntegerParameter()),
					Subject:       ps[13].StringParamter,
					Body:          ps[14].GetStringParamter(),
					ServiceCenter: ps[16].StringParamter,
					Read:          ps[8].GetIntegerParameter(),
					Status:        int64(*ps[9].IntegerParameter),
					DateSent:      ps[6].IntegerParameter,
					ReadableDate:  &readable,
					ContactName:   ps[4].StringParamter,
				}
				smses.SMS = append(smses.SMS, sms)
			}

			if strings.HasPrefix(*stmt.Statement, "INSERT INTO mms") {
				// TODO this
				log.Println("MMS export not yet supported")
			}
		}
	}

	smses.Count = len(smses.SMS)
	x, err := xml.Marshal(smses)
	if err != nil {
		return errors.Wrap(err, "unable to format XML")
	}

	w := types.NewMultiWriter(out)
	w.W([]byte(`<?xml version='1.0' encoding='UTF-8' standalone='yes' ?>`))
	w.W([]byte(`<?xml-stylesheet type="text/xsl" href="sms.xsl"?>`))
	w.W(x)
	return errors.WithMessage(w.Error(), "failed to write out XML")
}

func translateSMSType(t uint64) types.SMSType {
	// Just get the lower 8 bits, because everything else is masking.
	// https://github.com/signalapp/Signal-Android/blob/master/src/org/thoughtcrime/securesms/database/MmsSmsColumns.java
	v := uint8(t)

	switch v {
	// STANDARD
	case 1: // standard standard
		return types.SMSReceived
	case 2: // standard sent
		return types.SMSSent
	case 3: // standard draft
		return types.SMSDraft
	case 4: // standard outbox
		return types.SMSOutbox
	case 5: // standard failed
		return types.SMSFailed
	case 6: // standard queued
		return types.SMSQueued

		// SIGNAL
	case 20: // signal received
		return types.SMSReceived
	case 21: // signal outbox
		return types.SMSOutbox
	case 22: // signal sending
		return types.SMSQueued
	case 23: // signal sent
		return types.SMSSent
	case 24: // signal failed
		return types.SMSFailed
	case 25: // pending secure SMS fallback
		return types.SMSQueued
	case 26: // pending insecure SMS fallback
		return types.SMSQueued
	case 27: // signal draft
		return types.SMSDraft

	default:
		panic(fmt.Sprintf("undefined SMS type: %#v\nplease report this issue, as well as (if possible) details about the SMS,\nsuch as whether it was sent, received, drafted, etc.", t))
	}
}
