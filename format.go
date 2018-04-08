package main

import (
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/xeals/signal-back/signal"
)

func slurp(bf *backupFile) ([]*signal.BackupFrame, error) {
	frames := []*signal.BackupFrame{}
	for {
		f, err := bf.frame()
		if err != nil {
			return frames, nil // TODO error matching
		}

		frames = append(frames, f)

		// Attachment needs removing
		if a := f.GetAttachment(); a != nil {
			_, err := bf.decryptAttachment(a, ioutil.Discard)
			if err != nil {
				return nil, errors.Wrap(err, "unable to chew through attachment")
			}
		}
	}
}

func analyseTables(bf *backupFile) (map[string]int, error) {
	counts := make(map[string]int)

	frames, err := slurp(bf)
	if err != nil {
		return nil, errors.Wrap(err, "failed to slurp frames")
	}
	for _, f := range frames {
		if f.GetHeader() != nil {
			counts["header"]++
			continue
		}
		if f.GetVersion() != nil {
			counts["version"]++
			continue
		}
		if f.GetAttachment() != nil {
			counts["attachment"]++
			continue
		}
		if f.GetAvatar() != nil {
			counts["avatar"]++
			continue
		}
		if f.GetPreference() != nil {
			counts["pref"]++
			continue
		}
		if stmt := f.GetStatement(); stmt != nil {
			if strings.HasPrefix(*stmt.Statement, "DROP TABLE") {
				if counts["drop_table"] == 0 {
					fmt.Println(stmt)
				}
				counts["drop_table"]++
				continue
			}
			if strings.HasPrefix(*stmt.Statement, "CREATE TABLE") {
				if counts["create_table"] == 0 {
					fmt.Println(stmt)
				}
				counts["create_table"]++
				continue
			}
			if strings.HasPrefix(*stmt.Statement, "DROP INDEX") {
				if counts["drop_index"] == 0 {
					fmt.Println(stmt)
				}
				counts["drop_index"]++
				continue
			}
			if strings.HasPrefix(*stmt.Statement, "CREATE INDEX") ||
				strings.HasPrefix(*stmt.Statement, "CREATE UNIQUE INDEX") {
				if counts["create_index"] == 0 {
					fmt.Println(stmt)
				}
				counts["create_index"]++
				continue
			}
			if strings.HasPrefix(*stmt.Statement, "INSERT INTO") {
				table := strings.Split(*stmt.Statement, " ")[2]
				if counts["insert_into_"+table] == 0 {
					fmt.Println(stmt)
				}
				counts["insert_into_"+table]++
				continue
			}

			counts["other_stmt"]++
		}
	}

	return counts, nil
}

func formatJSON(bf *backupFile, out io.Writer) error {
	tables, err := analyseTables(bf)
	if err != nil {
		return errors.Wrap(err, "failed to analyse tables")
	}
	out.Write([]byte(fmt.Sprintf("%v\n", tables)))
	return nil
}

// Formats the backup into the same XML format as SMS Backup & Restore
// uses. Layout described at their website
// http://synctech.com.au/fields-in-xml-backup-files/
func formatXML(bf *backupFile, out io.Writer) error {
	smses := &SMSes{}
	for {
		f, err := bf.frame()
		if err != nil {
			break
		}

		// Attachment needs removing
		if a := f.GetAttachment(); a != nil {
			_, err := bf.decryptAttachment(a, ioutil.Discard)
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

				sms := SMS{
					Protocol:      ps[7].IntegerParameter,
					Address:       ps[2].GetStringParamter(),
					Date:          ps[5].GetStringParamter(),
					Type:          ps[10].GetIntegerParameter(),
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
			}
		}
	}

	smses.Count = len(smses.SMS)
	x, err := xml.Marshal(smses)
	if err != nil {
		return errors.Wrap(err, "unable to format XML")
	}

	w := multiWriter{out, nil}
	w.write([]byte(`<?xml version='1.0' encoding='UTF-8' standalone='yes' ?>'\n`))
	w.write([]byte(`<?xml-stylesheet type="text/xsl" href="sms.xsl"?>'\n`))
	w.write(x)
	return errors.WithMessage(w.err, "failed to write out XML")
}

type multiWriter struct {
	io.Writer
	err error
}

func (w *multiWriter) write(p []byte) {
	if w.err != nil {
		return
	}
	_, w.err = w.Write(p)
}

type SMS struct {
	XMLName       xml.Name `xml:"sms"`
	Protocol      *uint64  `xml:"protocol,attr"`       // optional
	Address       string   `xml:"address,attr"`        // required
	Date          string   `xml:"date,attr"`           // required
	Type          uint64   `xml:"type,attr"`           // required
	Subject       *string  `xml:"subject,attr"`        // optional
	Body          string   `xml:"body,attr"`           // required
	TOA           *string  `xml:"toa,attr"`            // optional
	SCTOA         *string  `xml:"sc_toa,attr"`         // optional
	ServiceCenter *string  `xml:"service_center,attr"` // optional
	Read          uint64   `xml:"read,attr"`           // required
	Status        int64    `xml:"status,attr"`         // required
	Locked        *uint64  `xml:"locked,attr"`         // optional
	DateSent      *uint64  `xml:"date_sent,attr"`      // optional
	ReadableDate  *string  `xml:"readable_date,attr"`  // optional
	ContactName   *string  `xml:"contact_name,attr"`   // optional
}

type SMSes struct {
	XMLName xml.Name `xml:"smses"`
	Count   int      `xml:"count,attr"`
	SMS     []SMS    `xml:"sms"`
}
