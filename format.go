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

type MMS struct {
	XMLName      xml.Name  `xml:"mms"`
	Parts        []MMSPart `xml:"parts,attr"`
	TextOnly     *uint64   `xml:"text_only,attr"`     // optional
	Sub          *string   `xml:"sub,attr"`           // optional
	RetrSt       string    `xml:"retr_st,attr"`       // required
	Date         uint64    `xml:"date,attr"`          // required
	CtCls        string    `xml:"ct_cls,attr"`        // required
	SubCs        string    `xml:"sub_cs,attr"`        // required
	Read         uint64    `xml:"read,attr"`          // required
	CtL          string    `xml:"ct_l,attr"`          // required
	TrID         string    `xml:"tr_id,attr"`         // required
	St           string    `xml:"st,attr"`            // required
	MsgBox       uint64    `xml:"msg_box,attr"`       // required
	Address      uint64    `xml:"address,attr"`       // required
	MCls         string    `xml:"m_cls,attr"`         // required
	DTm          string    `xml:"d_tm,attr"`          // required
	ReadStatus   string    `xml:"read_status,attr"`   // required
	CtT          string    `xml:"ct_t,attr"`          // required
	RetrTxtCs    string    `xml:"retr_txt_cs,attr"`   // required
	DRpt         uint64    `xml:"d_rpt,attr"`         // required
	MId          string    `xml:"m_id,attr"`          // required
	DateSent     uint64    `xml:"date_sent,attr"`     // required
	Seen         uint64    `xml:"seen,attr"`          // required
	MType        uint64    `xml:"m_type,attr"`        // required
	V            uint64    `xml:"v,attr"`             // required
	Exp          string    `xml:"exp,attr"`           // required
	Pri          uint64    `xml:"pri,attr"`           // required
	Rr           uint64    `xml:"rr,attr"`            // required
	RespTxt      string    `xml:"resp_txt,attr"`      // required
	RptA         string    `xml:"rpt_a,attr"`         // required
	Locked       uint64    `xml:"locked,attr"`        // required
	RetrTxt      string    `xml:"retr_txt,attr"`      // required
	RespSt       string    `xml:"resp_st,attr"`       // required
	MSize        string    `xml:"m_size,attr"`        // required
	ReadableDate *string   `xml:"readable_date,attr"` // optional
	ContactName  *string   `xml:"contact_name,attr"`  // optional
}

type MMSPart struct {
	XMLName xml.Name `xml:"part"`
	Seq     uint64   `xml:"seq,attr"`   // required
	Ct      uint64   `xml:"ct,attr"`    // required
	Name    string   `xml:"name,attr"`  // required
	ChSet   string   `xml:"chset,attr"` // required
	Cd      string   `xml:"cd,attr"`    // required
	Fn      string   `xml:"fn,attr"`    // required
	CID     string   `xml:"cid,attr"`   // required
	Cl      string   `xml:"cl,attr"`    // required
	CttS    string   `xml:"ctt_s,attr"` // required
	CttT    string   `xml:"ctt_t,attr"` // required
	Text    string   `xml:"text,attr"`  // required
	Data    *string  `xml:"data,attr"`  // optional
}

type SMSes struct {
	XMLName xml.Name `xml:"smses"`
	Count   int      `xml:"count,attr"`
	SMS     []SMS    `xml:"sms"`
	MMS     []MMS    `xml:"mms"`
}
