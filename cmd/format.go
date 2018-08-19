package cmd

import (
	"bytes"
	"encoding/base64"
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
	"github.com/xeals/signal-back/signal"
	"github.com/xeals/signal-back/types"
)

// Format fulfils the `format` subcommand.
var Format = cli.Command{
	Name:               "format",
	Usage:              "Read and format the backup file",
	UsageText:          "Parse and transform the backup file into other formats.\nValid formats include: CSV, XML, RAW.",
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

		switch strings.ToLower(c.String("format")) {
		case "csv":
			err = CSV(bf, strings.ToLower(c.String("message")), out)
		case "xml":
			err = XML(bf, out)
		case "json":
			// err = formatJSON(bf, out)
			return errors.New("JSON is still TODO")
		case "raw":
			err = Raw(bf, out)
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
	type attachmentDetails struct {
		Size uint64
		Body string
	}

	var attachmentBuffer bytes.Buffer
	attachmentEncoder := base64.NewEncoder(base64.StdEncoding, &attachmentBuffer)
	attachments := map[uint64]attachmentDetails{}
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
			err := bf.DecryptAttachment(a, attachmentEncoder)
			attachmentEncoder.Close()
			if err != nil {
				return errors.Wrap(err, "unable to process attachment")
			}
			attachments[*a.AttachmentId] = attachmentDetails{
				Size: uint64(*a.Length),
				Body: attachmentBuffer.String(),
			}
			attachmentBuffer.Reset()
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
				id, mms, err := types.NewMMSFromStatement(stmt)
				if err == nil {
					mmses[id] = *mms
				}
			}

			if strings.HasPrefix(*stmt.Statement, "INSERT INTO part") {
				mmsId, part, err := types.NewPartFromStatement(stmt)
				if err == nil {
					mmsParts[mmsId] = append(mmsParts[mmsId], *part)
				}
			}
		}
	}

	for id, mms := range mmses {
		var messageSize uint64
		parts, ok := mmsParts[id]
		if ok {
			for i := 0; i < len(parts); i++ {
				if attachment, ok := attachments[parts[i].UniqueID]; ok {
					messageSize += attachment.Size
					parts[i].Data = &attachment.Body
				}
			}
		}
		if mms.Body != nil && len(*mms.Body) > 0 {
			parts = append(parts, types.MMSPart{
				Seq:   0,
				Ct:    "text/plain",
				Name:  "null",
				ChSet: types.CharsetUTF8,
				Cd:    "null",
				Fn:    "null",
				CID:   "null",
				Cl:    fmt.Sprintf("txt%06d.txt", id),
				CttS:  "null",
				CttT:  "null",
				Text:  *mms.Body,
			})
			messageSize += uint64(len(*mms.Body))
			if len(parts) == 1 {
				mms.TextOnly = 1
			}
		}
		if len(parts) == 0 {
			continue
		}
		mms.Parts = parts
		mms.MSize = &messageSize
		if mms.MType == nil {
			if types.SetMMSMessageType(types.MMSSendReq, &mms) != nil {
				panic("logic error: this should never happen")
			}
			smses.MMS = append(smses.MMS, mms)
			if types.SetMMSMessageType(types.MMSRetrieveConf, &mms) != nil {
				panic("logic error: this should never happen")
			}
		}
		smses.MMS = append(smses.MMS, mms)
	}

	smses.Count = len(smses.SMS)
	x, err := xml.MarshalIndent(smses, "", "  ")
	if err != nil {
		return errors.Wrap(err, "unable to format XML")
	}

	w := types.NewMultiWriter(out)
	w.W([]byte("<?xml version='1.0' encoding='UTF-8' standalone='yes' ?>\n"))
	w.W([]byte("<?xml-stylesheet type=\"text/xsl\" href=\"sms.xsl\" ?>\n"))
	w.W(x)
	return errors.WithMessage(w.Error(), "failed to write out XML")
}

// Raw performs an ever plainer dump than CSV, and is largely unusable for any purpose outside
// debugging.
func Raw(bf *types.BackupFile, out io.Writer) error {
	var (
		err error
		f   *signal.BackupFrame
	)

	for {
		f, err = bf.Frame()
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
			out.Write([]byte(stmt.String()))
			out.Write([]byte{'\n'})
		}
	}
	return nil
}
