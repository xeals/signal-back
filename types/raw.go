package types

import (
	"strconv"

	"github.com/xeals/signal-back/signal"
)

type SQLSMS struct {
	ID                   uint64
	ThreadID             *uint64
	Address              *string
	AddressDeviceID      uint64 // default 1
	Person               *uint64
	DateReceived         *uint64
	DateSent             *uint64
	Protocol             uint64 // effectively default 0
	Read                 uint64 // default 0
	Status               uint64 // default -1
	Type                 *uint64
	ReplyPathPresent     *uint64
	DeliveryReceiptCount uint64 // default 0
	Subject              *string
	Body                 *string
	MismatchedIdentities *string // default null
	ServiceCenter        *string
	SubscriptionID       uint64 // default -1
	ExpiresIn            uint64 // default 0
	ExpireStarted        uint64 // default 0
	Notified             uint64 // default 0
	ReadReceiptCount     uint64 // default 0
}

func StatementToStringArray(sql *signal.SqlStatement) []string {
	s := make([]string, 0)
	for _, p := range sql.GetParameters() {
		if p.IntegerParameter != nil {
			s = append(s, strconv.Itoa(int(*p.IntegerParameter)))
		} else if p.StringParamter != nil {
			s = append(s, *p.StringParamter)
		}
	}
	return s
}

// StatementToSMS converts a of SQL statement to a single SMS.
func StatementToSMS(sql *signal.SqlStatement) *SQLSMS {
	return ParametersToSMS(sql.GetParameters())
}

// ParametersToSMS converts a set of SQL parameters to a single SMS.
func ParametersToSMS(ps []*signal.SqlStatement_SqlParameter) *SQLSMS {
	return &SQLSMS{
		ID:                   ps[0].GetIntegerParameter(),
		ThreadID:             ps[1].IntegerParameter,
		Address:              ps[2].StringParamter,
		AddressDeviceID:      ps[3].GetIntegerParameter(),
		Person:               ps[4].IntegerParameter,
		DateReceived:         ps[5].IntegerParameter,
		DateSent:             ps[6].IntegerParameter,
		Protocol:             ps[7].GetIntegerParameter(),
		Read:                 ps[8].GetIntegerParameter(),
		Status:               ps[9].GetIntegerParameter(),
		Type:                 ps[10].IntegerParameter,
		ReplyPathPresent:     ps[11].IntegerParameter,
		DeliveryReceiptCount: ps[12].GetIntegerParameter(),
		Subject:              ps[13].StringParamter,
		Body:                 ps[14].StringParamter,
		MismatchedIdentities: ps[15].StringParamter,
		ServiceCenter:        ps[16].StringParamter,
		SubscriptionID:       ps[17].GetIntegerParameter(),
		ExpiresIn:            ps[18].GetIntegerParameter(),
		ExpireStarted:        ps[19].GetIntegerParameter(),
		Notified:             ps[20].GetIntegerParameter(),
		ReadReceiptCount:     ps[21].GetIntegerParameter(),
	}
}

// StringArray converts an SMS to a CSV-esque array.
func (sms *SQLSMS) StringArray() []string {
	s := []string{strconv.Itoa(int(sms.ID))}

	if sms.ThreadID != nil {
		s = append(s, strconv.Itoa(int(*sms.ThreadID)))
	} else {
		s = append(s, "")
	}
	if sms.Address != nil {
		s = append(s, *sms.Address)
	} else {
		s = append(s, "")
	}
	s = append(s, strconv.Itoa(int(sms.AddressDeviceID)))
	if sms.Person != nil {
		s = append(s, strconv.Itoa(int(*sms.Person)))
	} else {
		s = append(s, "")
	}
	if sms.DateReceived != nil {
		s = append(s, strconv.Itoa(int(*sms.DateReceived)))
	} else {
		s = append(s, "")
	}
	if sms.DateSent != nil {
		s = append(s, strconv.Itoa(int(*sms.DateSent)))
	} else {
		s = append(s, "")
	}
	s = append(s, strconv.Itoa(int(sms.Protocol)))
	s = append(s, strconv.Itoa(int(sms.Read)))
	s = append(s, strconv.Itoa(int(sms.Status)))
	if sms.Type != nil {
		s = append(s, strconv.Itoa(int(*sms.Type)))
	} else {
		s = append(s, "")
	}
	if sms.ReplyPathPresent != nil {
		s = append(s, strconv.Itoa(int(*sms.ReplyPathPresent)))
	} else {
		s = append(s, "")
	}
	s = append(s, strconv.Itoa(int(sms.DeliveryReceiptCount)))
	if sms.Subject != nil {
		s = append(s, *sms.Subject)
	} else {
		s = append(s, "")
	}
	if sms.Body != nil {
		s = append(s, *sms.Body)
	} else {
		s = append(s, "")
	}
	if sms.MismatchedIdentities != nil {
		s = append(s, *sms.MismatchedIdentities)
	} else {
		s = append(s, "")
	}
	if sms.ServiceCenter != nil {
		s = append(s, *sms.ServiceCenter)
	} else {
		s = append(s, "")
	}
	s = append(s, strconv.Itoa(int(sms.SubscriptionID)))
	s = append(s, strconv.Itoa(int(sms.ExpiresIn)))
	s = append(s, strconv.Itoa(int(sms.ExpireStarted)))
	s = append(s, strconv.Itoa(int(sms.Notified)))
	s = append(s, strconv.Itoa(int(sms.ReadReceiptCount)))

	return s
}

type SQLMMS struct {
	ID                   uint64
	ThreadID             *uint64
	DateSent             *uint64
	DateReceived         *uint64
	MessageBox           *uint64
	Read                 uint64 // default 0
	MID                  *string
	Sub                  *string
	SubCs                *uint64
	Body                 *string
	PartCount            *uint64
	CtT                  *string
	ContentLocation      *string
	Address              *string
	AddressDeviceID      *uint64
	Expiry               *uint64
	MCls                 *string
	MessageType          *uint64
	V                    *uint64
	MessageSize          *uint64
	Pri                  *uint64
	Rr                   *uint64
	RptA                 *uint64
	RespSt               *uint64
	Status               *uint64
	TransactionID        *string
	RetrSt               *uint64
	RetrTxt              *string
	RetrTxtCs            *uint64
	ReadStatus           *uint64
	CtCls                *uint64
	RespTxt              *string
	DTm                  *uint64
	DeliveryReceiptCount uint64  // default 0
	MismatchedIdentities *string // default null
	NetworkFailure       *string // default null
	DRpt                 *uint64
	SubscriptionID       uint64 // default -1
	ExpiresIn            uint64 // default 0
	ExpireStarted        uint64 // default 0
	Notified             uint64 // default 0
	ReadReceiptCount     uint64 // default 0
}
