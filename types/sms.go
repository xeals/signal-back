package types

import (
	"encoding/xml"
	"log"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"github.com/xeals/signal-back/signal"
)

// Character sets as specified by IANA.
const (
	CharsetASCII = "3"
	CharsetUTF8  = "106"
)

// SMSType is an SMS type as defined by the XML backup spec.
type SMSType uint64

// SMS types
const (
	SMSInvalid  SMSType = iota // 0
	SMSReceived                // 1
	SMSSent                    // 2
	SMSDraft                   // 3
	SMSOutbox                  // 4
	SMSFailed                  // 5
	SMSQueued                  // 6
)

// MMS message types as defined by the MMS Encapsulation Protocol.
// See: http://www.openmobilealliance.org/release/MMS/V1_2-20050429-A/OMA-MMS-ENC-V1_2-20050301-A.pdf
const (
	MMSSendReq           uint64 = iota + 128 // 128
	MMSSendConf                              // 129
	MMSNotificationInd                       // 130
	MMSNotifyResponseInd                     // 131
	MMSRetrieveConf                          // 132
	MMSAckknowledgeInd                       // 133
	MMSDeliveryInd                           // 134
	MMSReadRecInd                            // 135
	MMSReadOrigInd                           // 136
	MMSForwardReq                            // 137
	MMSForwardConf                           // 138
	MMSMBoxStoreReq                          // 139
	MMSMBoxStoreConf                         // 140
	MMSMBoxViewReq                           // 141
	MMSMBoxViewConf                          // 142
	MMSMBoxUploadReq                         // 143
	MMSMBoxUploadConf                        // 144
	MMSMBoxDeleteReq                         // 145
	MMSMBoxDeleteConf                        // 146
	MMSMBoxDescr                             // 147
)

// Recipient represents a recipient record.
type Recipient struct {
	XMLName xml.Name `xml:"recipient"`
	Phone   string   `xml:"phone,attr"` // required
}

// SMSes holds a set of MMS or SMS records.
type SMSes struct {
	XMLName xml.Name `xml:"smses"`
	Count   int      `xml:"count,attr"`
	MMS     []MMS    `xml:"mms"`
	SMS     []SMS    `xml:"sms"`
}

// SMS represents a Short Message Service record.
type SMS struct {
	XMLName       xml.Name `xml:"sms"`
	Protocol      *uint64  `xml:"protocol,attr"`       // optional
	RecipientID   string   `xml:"recipient_id,attr"`   // optional
	Address       string   `xml:"address,attr"`        // required
	Date          string   `xml:"date,attr"`           // required
	Type          SMSType  `xml:"type,attr"`           // required
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
	ContactName   *uint64  `xml:"contact_name,attr"`   // optional
}

// MMS represents a Multimedia Messaging Service record.
type MMS struct {
	XMLName      xml.Name  `xml:"mms"`
	Parts        []MMSPart `xml:"parts"`
	Body         *string   `xml:"-"`
	TextOnly     uint64    `xml:"text_only,attr"`     // optional
	Sub          string    `xml:"sub,attr"`           // optional
	RetrSt       string    `xml:"retr_st,attr"`       // required
	Date         uint64    `xml:"date,attr"`          // required
	CtCls        string    `xml:"ct_cls,attr"`        // required
	SubCs        string    `xml:"sub_cs,attr"`        // required
	Read         uint64    `xml:"read,attr"`          // required
	CtL          string    `xml:"ct_l,attr"`          // required
	TrID         string    `xml:"tr_id,attr"`         // required
	St           string    `xml:"st,attr"`            // required
	MsgBox       uint64    `xml:"msg_box,attr"`       // required
	RecipientID  string    `xml:"recipient_id,attr"`  // required
	Address      string    `xml:"address,attr"`       // required
	MCls         string    `xml:"m_cls,attr"`         // required
	DTm          string    `xml:"d_tm,attr"`          // required
	ReadStatus   string    `xml:"read_status,attr"`   // required
	CtT          string    `xml:"ct_t,attr"`          // required
	RetrTxtCs    string    `xml:"retr_txt_cs,attr"`   // required
	DRpt         uint64    `xml:"d_rpt,attr"`         // required
	MId          string    `xml:"m_id,attr"`          // required
	DateSent     uint64    `xml:"date_sent,attr"`     // required
	Seen         uint64    `xml:"seen,attr"`          // required
	MType        *uint64   `xml:"m_type,attr"`        // required
	V            uint64    `xml:"v,attr"`             // required
	Exp          string    `xml:"exp,attr"`           // required
	Pri          uint64    `xml:"pri,attr"`           // required
	Rr           uint64    `xml:"rr,attr"`            // required
	RespTxt      string    `xml:"resp_txt,attr"`      // required
	RptA         string    `xml:"rpt_a,attr"`         // required
	Locked       uint64    `xml:"locked,attr"`        // required
	RetrTxt      string    `xml:"retr_txt,attr"`      // required
	RespSt       uint64    `xml:"resp_st,attr"`       // required
	MSize        *uint64   `xml:"m_size,attr"`        // required
	ReadableDate string    `xml:"readable_date,attr"` // optional
	ContactName  *string   `xml:"contact_name,attr"`  // optional
}

// MMSPart holds a data blob for an MMS.
type MMSPart struct {
	XMLName  xml.Name `xml:"part"`
	UniqueID uint64   `xml:"-"`
	Seq      uint64   `xml:"seq,attr"`   // required
	Ct       string   `xml:"ct,attr"`    // required
	Name     string   `xml:"name,attr"`  // required
	ChSet    string   `xml:"chset,attr"` // required
	Cd       string   `xml:"cd,attr"`    // required
	Fn       string   `xml:"fn,attr"`    // required
	CID      string   `xml:"cid,attr"`   // required
	Cl       string   `xml:"cl,attr"`    // required
	CttS     string   `xml:"ctt_s,attr"` // required
	CttT     string   `xml:"ctt_t,attr"` // required
	Text     string   `xml:"text,attr"`  // required
	Data     *string  `xml:"data,attr"`  // optional
}

// NewRecipientFromStatement constructs an XML recipient struct from a SQL statement.
func NewRecipientFromStatement(stmt *signal.SqlStatement) (uint64, *Recipient, error) {
	recipient := StatementToRecipient(stmt)
	if recipient == nil {
		return 0, nil, errors.Errorf("expected 28 columns for recipient, have %v", len(stmt.GetParameters()))
	}

	xml := Recipient{
		Phone: recipient.Phone,
	}

	return recipient.ID, &xml, nil
}

// NewSMSFromStatement constructs an XML SMS struct from a SQL statement.
func NewSMSFromStatement(stmt *signal.SqlStatement) (*SMS, error) {
	sms := StatementToSMS(stmt)
	if sms == nil {
		return nil, errors.Errorf("expected 22 columns for SMS, have %v", len(stmt.GetParameters()))
	}

	xml := SMS{
		Protocol:      &sms.Protocol,
		Subject:       sms.Subject,
		ServiceCenter: sms.ServiceCenter,
		Read:          sms.Read,
		Status:        int64(sms.Status),
		DateSent:      sms.DateSent,
		ReadableDate:  intToTime(sms.DateReceived),
	}

	if sms.RecipientID != nil {
		xml.RecipientID = *sms.RecipientID
	}
	if sms.Type != nil {
		xml.Type = translateSMSType(*sms.Type)
	}
	if sms.Body != nil {
		xml.Body = *sms.Body
	}
	if sms.DateReceived != nil {
		xml.Date = strconv.FormatUint(*sms.DateReceived, 10)
	}
	if sms.Person != nil {
		xml.ContactName = sms.Person
	}

	return &xml, nil
}

func NewMMSFromStatement(stmt *signal.SqlStatement) (uint64, *MMS, error) {
	mms := StatementToMMS(stmt)
	if mms == nil {
		return 0, nil, errors.Errorf("expected at least 42 columns for MMS, have %v", len(stmt.GetParameters()))
	}

	xml := MMS{
		TextOnly:     0,
		Sub:          "null",
		RetrSt:       "null",
		Date:         *mms.DateReceived,
		CtCls:        "null",
		SubCs:        "null",
		Body:         nil,
		Read:         mms.Read,
		CtL:          "null",
		TrID:         "null",
		St:           "null",
		MCls:         "personal",
		DTm:          "null",
		ReadStatus:   "null",
		CtT:          "application/vnd.wap.multipart.related",
		RetrTxtCs:    "null",
		DateSent:     *mms.DateSent / 1000,
		Seen:         mms.Read,
		Exp:          "null",
		RespTxt:      "null",
		RptA:         "null",
		Locked:       0,
		RetrTxt:      "null",
		MSize:        nil,
		ReadableDate: *intToTime(mms.DateReceived),
	}

	if mms.MessageType != nil {
		if err := SetMMSMessageType(*mms.MessageType, &xml); err != nil {
			log.Fatalf("%v\nplease report this issue, as well as (if possible) details about the MMS\nthe ID of the offending MMS is: %d", err, mms.ID)
		}
	}

	if mms.RetrSt != nil {
		xml.RetrSt = strconv.FormatUint(*mms.RetrSt, 10)
	}
	if mms.CtCls != nil {
		xml.CtCls = strconv.FormatUint(*mms.CtCls, 10)
	}
	if mms.SubCs != nil {
		xml.SubCs = strconv.FormatUint(*mms.SubCs, 10)
	}
	if mms.Body != nil {
		xml.Body = mms.Body
	}
	if mms.ContentLocation != nil {
		xml.CtL = *mms.ContentLocation
	}
	if mms.TransactionID != nil {
		xml.TrID = *mms.TransactionID
	}
	if mms.RecipientID != nil {
		xml.RecipientID = *mms.RecipientID
	}
	if mms.Expiry != nil {
		xml.Exp = strconv.FormatUint(*mms.Expiry, 10)
	}
	if mms.MCls != nil {
		xml.MCls = *mms.MCls
	}
	if mms.DTm != nil {
		xml.DTm = strconv.FormatUint(*mms.DTm, 10)
	}
	if mms.ReadStatus != nil {
		xml.ReadStatus = strconv.FormatUint(*mms.ReadStatus, 10)
	}
	if mms.CtT != nil {
		xml.CtT = *mms.CtT
	}
	if mms.RetrTxtCs != nil {
		xml.RetrTxtCs = strconv.FormatUint(*mms.RetrTxtCs, 10)
	}
	if mms.DRpt != nil {
		xml.DRpt = *mms.DRpt
	}
	if mms.MID != nil {
		xml.MId = *mms.MID
	}
	if mms.Pri != nil {
		xml.Pri = *mms.Pri
	}
	if mms.Rr != nil {
		xml.Rr = *mms.Rr
	}
	if mms.RespTxt != nil {
		xml.RespTxt = *mms.RespTxt
	}
	if mms.RptA != nil {
		xml.RptA = strconv.FormatUint(*mms.RptA, 10)
	}
	if mms.RetrTxt != nil {
		xml.RetrTxt = *mms.RetrTxt
	}
	if mms.RespSt != nil {
		xml.RespSt = *mms.RespSt
	}
	if mms.MessageSize != nil {
		xml.MSize = mms.MessageSize
	}

	return mms.ID, &xml, nil
}

func SetMMSMessageType(messageType uint64, mms *MMS) error {
	switch messageType {
	case MMSSendReq:
		mms.MsgBox = 2
		mms.V = 18
		break
	case MMSNotificationInd:
		// We can safely ignore this case.
		break
	case MMSRetrieveConf:
		mms.MsgBox = 1
		mms.V = 16
		break
	default:
		return errors.Errorf("unsupported message type %v encountered", messageType)
	}

	mms.MType = &messageType
	return nil
}

func NewPartFromStatement(stmt *signal.SqlStatement) (uint64, *MMSPart, error) {
	part := StatementToPart(stmt)
	if part == nil {
		return 0, nil, errors.Errorf("expected at least 25 columns for part, have %v", len(stmt.GetParameters()))
	}

	xml := MMSPart{
		UniqueID: part.UniqueID,
		Seq:      part.Seq,
		Ct:       *part.ContentType,
		Name:     "null",
		ChSet:    CharsetUTF8,
		Cd:       "null",
		Fn:       "null",
		CID:      "null",
		Cl:       "null",
		CttS:     "null",
		CttT:     "null",
	}

	if part.Name != nil {
		xml.Name = *part.Name
	}
	if part.Chset != nil {
		xml.ChSet = strconv.FormatUint(*part.Chset, 10)
	}
	if part.ContentDisposition != nil {
		xml.Cd = *part.ContentDisposition
	}
	if part.Fn != nil {
		xml.Fn = *part.Fn
	}
	if part.Cid != nil {
		xml.CID = *part.Cid
	}
	if part.ContentLocation != nil {
		xml.Cl = *part.ContentLocation
	}
	if part.CttS != nil {
		xml.CttS = strconv.FormatUint(*part.CttS, 10)
	}
	if part.CttT != nil {
		xml.CttT = *part.CttT
	}

	return *part.MmsID, &xml, nil
}

func intToTime(n *uint64) *string {
	if n == nil {
		return nil
	}
	unix := time.Unix(int64(*n)/1000, 0)
	t := unix.Format("Jan 02, 2006 3:04:05 PM")
	return &t
}

func translateSMSType(t uint64) SMSType {
	// Just get the lowest 5 bits, because everything else is masking.
	// https://github.com/signalapp/Signal-Android/blob/master/src/org/thoughtcrime/securesms/database/MmsSmsColumns.java
	v := uint8(t) & 0x1F

	switch v {
	// STANDARD
	case 1: // standard standard
		return SMSReceived
	case 2: // standard sent
		return SMSSent
	case 3: // standard draft
		return SMSDraft
	case 4: // standard outbox
		return SMSOutbox
	case 5: // standard failed
		return SMSFailed
	case 6: // standard queued
		return SMSQueued

		// SIGNAL
	case 20: // signal received
		return SMSReceived
	case 21: // signal outbox
		return SMSOutbox
	case 22: // signal sending
		return SMSQueued
	case 23: // signal sent
		return SMSSent
	case 24: // signal failed
		return SMSFailed
	case 25: // pending secure SMS fallback
		return SMSQueued
	case 26: // pending insecure SMS fallback
		return SMSQueued
	case 27: // signal draft
		return SMSDraft

	default:
		log.Fatalf("undefined SMS type: %#v\nplease report this issue, as well as (if possible) details about the SMS,\nsuch as whether it was sent, received, drafted, etc.\n", t)
		log.Fatalf("note that the output XML may not properly import to Signal\n")
		return SMSInvalid
	}
}
