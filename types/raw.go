package types

import (
	"strconv"

	"github.com/xeals/signal-back/signal"
)

// StatementToStringArray formats a SqlStatement fairly literally as an array.
// Null parameters are left empty.
func StatementToStringArray(sql *signal.SqlStatement) []string {
	s := make([]string, len(sql.GetParameters()))
	for i, p := range sql.GetParameters() {
		if p.IntegerParameter != nil {
			s[i] = strconv.Itoa(int(*p.IntegerParameter))
		} else if p.StringParamter != nil {
			s[i] = *p.StringParamter
		}
	}
	return s
}

// CSV column headers.
var (
	SMSCSVHeaders = []string{
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
	}

	MMSCSVHeaders = []string{
		"ID",
		"THREAD_ID",
		"DATE_SENT",
		"DATE_RECEIVED",
		"MESSAGE_BOX",
		"READ",
		"m_id",
		"sub",
		"sub_cs",
		"BODY",
		"PART_COUNT",
		"ct_t",
		"CONTENT_LOCATION",
		"ADDRESS",
		"ADDRESS_DEVICE_ID",
		"EXPIRY",
		"m_cls",
		"MESSAGE_TYPE",
		"v",
		"MESSAGE_SIZE",
		"pri",
		"rr",
		"rpt_a",
		"resp_st",
		"STATUS",
		"TRANSACTION_ID",
		"retr_st",
		"retr_txt",
		"retr_txt_cs",
		"read_status",
		"ct_cls",
		"resp_txt",
		"d_tm",
		"DELIVERY_RECEIPT_COUNT",
		"MISMATCHED_IDENTITIES",
		"NETWORK_FAILURE",
		"d_rpt",
		"SUBSCRIPTION_ID",
		"EXPIRES_IN",
		"EXPIRE_STARTED",
		"NOTIFIED",
		"READ_RECEIPT_COUNT",
	}
)

// type SQLSMS struct {
// 	ID                   uint64
// 	ThreadID             *uint64
// 	Address              *string
// 	AddressDeviceID      uint64 // default 1
// 	Person               *uint64
// 	DateReceived         *uint64
// 	DateSent             *uint64
// 	Protocol             uint64 // effectively default 0
// 	Read                 uint64 // default 0
// 	Status               uint64 // default -1
// 	Type                 *uint64
// 	ReplyPathPresent     *uint64
// 	DeliveryReceiptCount uint64 // default 0
// 	Subject              *string
// 	Body                 *string
// 	MismatchedIdentities *string // default null
// 	ServiceCenter        *string
// 	SubscriptionID       uint64 // default -1
// 	ExpiresIn            uint64 // default 0
// 	ExpireStarted        uint64 // default 0
// 	Notified             uint64 // default 0
// 	ReadReceiptCount     uint64 // default 0
// }

// // StatementToSMS converts a of SQL statement to a single SMS.
// func StatementToSMS(sql *signal.SqlStatement) *SQLSMS {
// 	return ParametersToSMS(sql.GetParameters())
// }

// // ParametersToSMS converts a set of SQL parameters to a single SMS.
// func ParametersToSMS(ps []*signal.SqlStatement_SqlParameter) *SQLSMS {
// 	if len(ps) != 22 {
// 		return nil
// 	}
// 	return &SQLSMS{
// 		ID:                   ps[0].GetIntegerParameter(),
// 		ThreadID:             ps[1].IntegerParameter,
// 		Address:              ps[2].StringParamter,
// 		AddressDeviceID:      ps[3].GetIntegerParameter(),
// 		Person:               ps[4].IntegerParameter,
// 		DateReceived:         ps[5].IntegerParameter,
// 		DateSent:             ps[6].IntegerParameter,
// 		Protocol:             ps[7].GetIntegerParameter(),
// 		Read:                 ps[8].GetIntegerParameter(),
// 		Status:               ps[9].GetIntegerParameter(),
// 		Type:                 ps[10].IntegerParameter,
// 		ReplyPathPresent:     ps[11].IntegerParameter,
// 		DeliveryReceiptCount: ps[12].GetIntegerParameter(),
// 		Subject:              ps[13].StringParamter,
// 		Body:                 ps[14].StringParamter,
// 		MismatchedIdentities: ps[15].StringParamter,
// 		ServiceCenter:        ps[16].StringParamter,
// 		SubscriptionID:       ps[17].GetIntegerParameter(),
// 		ExpiresIn:            ps[18].GetIntegerParameter(),
// 		ExpireStarted:        ps[19].GetIntegerParameter(),
// 		Notified:             ps[20].GetIntegerParameter(),
// 		ReadReceiptCount:     ps[21].GetIntegerParameter(),
// 	}
// }

// // StringArray converts an SMS to a CSV-esque array.
// func (sms *SQLSMS) StringArray() []string {
// 	s := []string{strconv.Itoa(int(sms.ID))}

// 	if sms.ThreadID != nil {
// 		s = append(s, strconv.Itoa(int(*sms.ThreadID)))
// 	} else {
// 		s = append(s, "")
// 	}
// 	if sms.Address != nil {
// 		s = append(s, *sms.Address)
// 	} else {
// 		s = append(s, "")
// 	}
// 	s = append(s, strconv.Itoa(int(sms.AddressDeviceID)))
// 	if sms.Person != nil {
// 		s = append(s, strconv.Itoa(int(*sms.Person)))
// 	} else {
// 		s = append(s, "")
// 	}
// 	if sms.DateReceived != nil {
// 		s = append(s, strconv.Itoa(int(*sms.DateReceived)))
// 	} else {
// 		s = append(s, "")
// 	}
// 	if sms.DateSent != nil {
// 		s = append(s, strconv.Itoa(int(*sms.DateSent)))
// 	} else {
// 		s = append(s, "")
// 	}
// 	s = append(s, strconv.Itoa(int(sms.Protocol)))
// 	s = append(s, strconv.Itoa(int(sms.Read)))
// 	s = append(s, strconv.Itoa(int(sms.Status)))
// 	if sms.Type != nil {
// 		s = append(s, strconv.Itoa(int(*sms.Type)))
// 	} else {
// 		s = append(s, "")
// 	}
// 	if sms.ReplyPathPresent != nil {
// 		s = append(s, strconv.Itoa(int(*sms.ReplyPathPresent)))
// 	} else {
// 		s = append(s, "")
// 	}
// 	s = append(s, strconv.Itoa(int(sms.DeliveryReceiptCount)))
// 	if sms.Subject != nil {
// 		s = append(s, *sms.Subject)
// 	} else {
// 		s = append(s, "")
// 	}
// 	if sms.Body != nil {
// 		s = append(s, *sms.Body)
// 	} else {
// 		s = append(s, "")
// 	}
// 	if sms.MismatchedIdentities != nil {
// 		s = append(s, *sms.MismatchedIdentities)
// 	} else {
// 		s = append(s, "")
// 	}
// 	if sms.ServiceCenter != nil {
// 		s = append(s, *sms.ServiceCenter)
// 	} else {
// 		s = append(s, "")
// 	}
// 	s = append(s, strconv.Itoa(int(sms.SubscriptionID)))
// 	s = append(s, strconv.Itoa(int(sms.ExpiresIn)))
// 	s = append(s, strconv.Itoa(int(sms.ExpireStarted)))
// 	s = append(s, strconv.Itoa(int(sms.Notified)))
// 	s = append(s, strconv.Itoa(int(sms.ReadReceiptCount)))

// 	return s
// }

// type SQLMMS struct {
// 	ID                   uint64
// 	ThreadID             *uint64
// 	DateSent             *uint64
// 	DateReceived         *uint64
// 	MessageBox           *uint64
// 	Read                 uint64 // default 0
// 	MID                  *string
// 	Sub                  *string
// 	SubCs                *uint64
// 	Body                 *string
// 	PartCount            *uint64
// 	CtT                  *string
// 	ContentLocation      *string
// 	Address              *string
// 	AddressDeviceID      *uint64
// 	Expiry               *uint64
// 	MCls                 *string
// 	MessageType          *uint64
// 	V                    *uint64
// 	MessageSize          *uint64
// 	Pri                  *uint64
// 	Rr                   *uint64
// 	RptA                 *uint64
// 	RespSt               *uint64
// 	Status               *uint64
// 	TransactionID        *string
// 	RetrSt               *uint64
// 	RetrTxt              *string
// 	RetrTxtCs            *uint64
// 	ReadStatus           *uint64
// 	CtCls                *uint64
// 	RespTxt              *string
// 	DTm                  *uint64
// 	DeliveryReceiptCount uint64  // default 0
// 	MismatchedIdentities *string // default null
// 	NetworkFailure       *string // default null
// 	DRpt                 *uint64
// 	SubscriptionID       uint64 // default -1
// 	ExpiresIn            uint64 // default 0
// 	ExpireStarted        uint64 // default 0
// 	Notified             uint64 // default 0
// 	ReadReceiptCount     uint64 // default 0
// }

// // StatementToMMS converts a of SQL statement to a single MMS.
// func StatementToMMS(sql *signal.SqlStatement) *SQLMMS {
// 	return ParametersToMMS(sql.GetParameters())
// }

// // ParametersToMMS converts a set of SQL parameters to a single MMS.
// func ParametersToMMS(ps []*signal.SqlStatement_SqlParameter) *SQLMMS {
// 	if len(ps) != 42 {
// 		return nil
// 	}
// 	return &SQLMMS{
// 		ID:                   ps[0].GetIntegerParameter(),
// 		ThreadID:             ps[1].IntegerParameter,
// 		DateSent:             ps[2].IntegerParameter,
// 		DateReceived:         ps[3].IntegerParameter,
// 		MessageBox:           ps[4].IntegerParameter,
// 		Read:                 ps[5].GetIntegerParameter(),
// 		MID:                  ps[6].StringParamter,
// 		Sub:                  ps[7].StringParamter,
// 		SubCs:                ps[8].IntegerParameter,
// 		Body:                 ps[9].StringParamter,
// 		PartCount:            ps[10].IntegerParameter,
// 		CtT:                  ps[11].StringParamter,
// 		ContentLocation:      ps[12].StringParamter,
// 		Address:              ps[13].StringParamter,
// 		AddressDeviceID:      ps[14].IntegerParameter,
// 		Expiry:               ps[15].IntegerParameter,
// 		MCls:                 ps[16].StringParamter,
// 		MessageType:          ps[17].IntegerParameter,
// 		V:                    ps[18].IntegerParameter,
// 		MessageSize:          ps[19].IntegerParameter,
// 		Pri:                  ps[20].IntegerParameter,
// 		Rr:                   ps[21].IntegerParameter,
// 		RptA:                 ps[22].IntegerParameter,
// 		RespSt:               ps[23].IntegerParameter,
// 		Status:               ps[24].IntegerParameter,
// 		TransactionID:        ps[25].StringParamter,
// 		RetrSt:               ps[26].IntegerParameter,
// 		RetrTxt:              ps[27].StringParamter,
// 		RetrTxtCs:            ps[28].IntegerParameter,
// 		ReadStatus:           ps[29].IntegerParameter,
// 		CtCls:                ps[30].IntegerParameter,
// 		RespTxt:              ps[31].StringParamter,
// 		DTm:                  ps[32].IntegerParameter,
// 		DeliveryReceiptCount: ps[33].GetIntegerParameter(),
// 		MismatchedIdentities: ps[34].StringParamter,
// 		NetworkFailure:       ps[35].StringParamter,
// 		DRpt:                 ps[36].IntegerParameter,
// 		SubscriptionID:       ps[37].GetIntegerParameter(),
// 		ExpiresIn:            ps[38].GetIntegerParameter(),
// 		ExpireStarted:        ps[39].GetIntegerParameter(),
// 		Notified:             ps[40].GetIntegerParameter(),
// 		ReadReceiptCount:     ps[41].GetIntegerParameter(),
// 	}
// }

// // StringArray converts an SMS to a CSV-esque array.
// func (sms *SQLMMS) StringArray() []string {
// 	s := []string{strconv.Itoa(int(sms.ID))}

// 	return s
// }
