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
	UsageText:          "Parse and transform the backup file into other formats.",
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

		switch c.String("format") {
		case "csv":
			err = CSV(bf, out)
		case "xml":
			err = XML(bf, out)
		case "json":
			// err = formatJSON(bf, out)
			return errors.New("JSON is still TODO")
		default:
			return errors.Errorf("format %s not recognised\nvalid formats are: xml", c.String("format"))
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

// CSV <undefined>
func CSV(bf *types.BackupFile, out io.Writer) error {
	smses := make([]*types.SQLSMS, 0)
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
			if strings.HasPrefix(*stmt.Statement, "INSERT INTO sms") {
				smses = append(smses, types.StatementToSMS(stmt))
			}
		}
	}

	w := csv.NewWriter(out)

	if err := w.Write([]string{
		"ID",
		"THREAD_ID",
		"ADDRESS",
		"ADDRESS_DEVICE_ID",
		"PERSON",
		"DATE_RECEIVED",
		"DATE_SENT",
		"PROTOCOL",
		"READ",
		"STATUS",
		"TYPE",
		"REPLY_PATH_PRESENT",
		"DELIVERY_RECEIPT_COUNT",
		"SUBJECT",
		"BODY",
		"MISMATCHED_IDENTITIES",
		"SERVICE_CENTER",
		"SUBSCRIPTION_ID",
		"EXPIRES_IN",
		"EXPIRE_STARTED",
		"NOTIFIED",
		"READ_RECEIPT_COUNT",
	}); err != nil {
		return errors.Wrap(err, "unable to write CSV headers")
	}

	for _, sms := range smses {
		if err := w.Write(sms.StringArray()); err != nil {
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
	switch t {
	case 3: // draft
		return types.SMSDraft
	case 20: // insecure received
		return types.SMSReceived
	case 23: // insecure sent
		return types.SMSSent
	case 2097156: // GSM? FIXME
		return types.SMSFailed
	case 10485780: // secure received
		return types.SMSReceived
	case 10485783: // secure sent
		return types.SMSSent
	default:
		panic(fmt.Sprintf("undefined SMS type: %v", t))
	}
}
