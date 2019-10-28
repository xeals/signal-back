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
		} else if p.StringParameter != nil {
			s[i] = *p.StringParameter
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
		"UNIDENTIFIED",
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
		"QUOTE_ID",
		"QUOTE_AUTHOR",
		"QUOTE_BODY",
		"QUOTE_ATTACHMENT",
		"QUOTE_MISSING",
		"SHARED_CONTACTS",
		"UNIDENTIFIED",
		"LINK_PREVIEWS",
		"VIEW_ONCE",
	}

	SMSPartCSVHeaders = []string{
		"ROW_ID",
		"MMS_ID",
		"seq",
		"CONTENT_TYPE",
		"NAME",
		"chset",
		"CONTENT_DISPOSITION",
		"fn",
		"cid",
		"CONTENT_LOCATION",
		"ctt_s",
		"ctt_t",
		"encrypted",
		"TRANSFER_STATE",
		"DATA",
		"SIZE",
		"FILE_NAME",
		"THUMBNAIL",
		"THUMBNAIL_ASPECT_RATIO",
		"UNIQUE_ID",
		"DIGEST",
		"FAST_PREFLIGHT_ID",
		"VOICE_NOTE",
		"DATA_RANDOM",
		"THUMBNAIL_RANDOM",
		"QUOTE",
		"WIDTH",
		"HEIGHT",
		"CAPTION",
		"STICKER_PACK_ID",
		"STICKER_PACK_KEY",
		"STICKER_ID",
		"DATA_HASH",
		"BLUR_HASH",
		"TRANSFORM_PROPERTIES",
	}
)

// SQLRecipient info
//
// https://github.com/signalapp/Signal-Android/blob/master/src/org/thoughtcrime/securesms/database/RecipientDatabase.java#L148
type SQLRecipient struct {
	ID                     uint64
	UUID                   *string
	Phone                  string
	Email                  *string
	GroupID                *string
	Blocked                uint64 // default 0
	MessageRingtone        *string
	MessageVibrate         uint64 // default 0
	CallRingtone           *string
	CallVibrate            uint64 // default 0
	NotificationChannel    *string
	MuteUntil              uint64 // default 0
	Color                  *string
	SeenInviteReminder     uint64 // default 0
	DefaultSubscriptionID  uint64 // default -1
	MessageExpirationTime  uint64 // default 0
	Registered             uint64 // default 0
	SystemDisplayName      *string
	SystemPhotoURI         *string
	SystemPhoneLabel       *string
	SystemPhoneType        uint64 // default -1
	SystemContactURI       *string
	ProfileKey             *string
	SignalProfileName      *string
	SignalProfileAvatar    *string
	ProfileSharing         uint64 // default 0
	UnidentifiedAccessMode uint64 // default 0
	ForceSmsSelection      uint64 // default 0
}

// StatementToRecipient converts a of SQL statement to a single recipîent.
func StatementToRecipient(sql *signal.SqlStatement) *SQLRecipient {
	return ParametersToRecipient(sql.GetParameters())
}

// ParametersToRecipient converts a set of SQL parameters to a single recipient.
func ParametersToRecipient(ps []*signal.SqlStatement_SqlParameter) *SQLRecipient {
	if len(ps) < 28 {
		return nil
	}

	result := &SQLRecipient{
		ID:                     ps[0].GetIntegerParameter(),
		UUID:                   ps[1].StringParameter,
		Phone:                  ps[2].GetStringParameter(),
		Email:                  ps[3].StringParameter,
		GroupID:                ps[4].StringParameter,
		Blocked:                ps[5].GetIntegerParameter(),
		MessageRingtone:        ps[6].StringParameter,
		MessageVibrate:         ps[7].GetIntegerParameter(),
		CallRingtone:           ps[8].StringParameter,
		CallVibrate:            ps[9].GetIntegerParameter(),
		NotificationChannel:    ps[10].StringParameter,
		MuteUntil:              ps[11].GetIntegerParameter(),
		Color:                  ps[12].StringParameter,
		SeenInviteReminder:     ps[13].GetIntegerParameter(),
		DefaultSubscriptionID:  ps[14].GetIntegerParameter(),
		MessageExpirationTime:  ps[15].GetIntegerParameter(),
		Registered:             ps[16].GetIntegerParameter(),
		SystemDisplayName:      ps[17].StringParameter,
		SystemPhotoURI:         ps[18].StringParameter,
		SystemPhoneLabel:       ps[19].StringParameter,
		SystemPhoneType:        ps[20].GetIntegerParameter(),
		SystemContactURI:       ps[21].StringParameter,
		ProfileKey:             ps[22].StringParameter,
		SignalProfileName:      ps[23].StringParameter,
		SignalProfileAvatar:    ps[24].StringParameter,
		ProfileSharing:         ps[25].GetIntegerParameter(),
		UnidentifiedAccessMode: ps[26].GetIntegerParameter(),
		ForceSmsSelection:      ps[27].GetIntegerParameter(),
	}

	return result
}

// SQLSMS info
//
// https://github.com/signalapp/Signal-Android/blob/master/src/org/thoughtcrime/securesms/database/SmsDatabase.java#L77
type SQLSMS struct {
	ID                   uint64
	ThreadID             *uint64
	RecipientID          *string
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
	Unidentified         uint64 // default 0
}

// StatementToSMS converts a of SQL statement to a single SMS.
func StatementToSMS(sql *signal.SqlStatement) *SQLSMS {
	return ParametersToSMS(sql.GetParameters())
}

// ParametersToSMS converts a set of SQL parameters to a single SMS.
func ParametersToSMS(ps []*signal.SqlStatement_SqlParameter) *SQLSMS {
	// TODO: update to 23 ?
	if len(ps) < 22 {
		return nil
	}

	result := &SQLSMS{
		ID:                   ps[0].GetIntegerParameter(),
		ThreadID:             ps[1].IntegerParameter,
		RecipientID:          ps[2].StringParameter,
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
		Subject:              ps[13].StringParameter,
		Body:                 ps[14].StringParameter,
		MismatchedIdentities: ps[15].StringParameter,
		ServiceCenter:        ps[16].StringParameter,
		SubscriptionID:       ps[17].GetIntegerParameter(),
		ExpiresIn:            ps[18].GetIntegerParameter(),
		ExpireStarted:        ps[19].GetIntegerParameter(),
		Notified:             ps[20].GetIntegerParameter(),
		ReadReceiptCount:     ps[21].GetIntegerParameter(),
		//Unidentified:         ps[22].GetIntegerParameter(),
	}

	return result
}

// SQLMMS info
//
// https://github.com/signalapp/Signal-Android/blob/master/src/org/thoughtcrime/securesms/database/MmsDatabase.java#L110
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
	RecipientID          *string
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
	QuoteID              uint64 // default 0
	QuoteAuthor          *string
	QuoteBody            *string
	QuoteAttachment      uint64 // default -1
	QuoteMissing         uint64 // default 0
	SharedContacts       *string
	Unidentified         uint64 // default 0
}

// StatementToMMS converts a of SQL statement to a single MMS.
func StatementToMMS(sql *signal.SqlStatement) *SQLMMS {
	return ParametersToMMS(sql.GetParameters())
}

// ParametersToMMS converts a set of SQL parameters to a single MMS.
func ParametersToMMS(ps []*signal.SqlStatement_SqlParameter) *SQLMMS {
	// TODO: update to 49 ?
	if len(ps) < 42 {
		return nil
	}

	result := &SQLMMS{
		ID:                   ps[0].GetIntegerParameter(),
		ThreadID:             ps[1].IntegerParameter,
		DateSent:             ps[2].IntegerParameter,
		DateReceived:         ps[3].IntegerParameter,
		MessageBox:           ps[4].IntegerParameter,
		Read:                 ps[5].GetIntegerParameter(),
		MID:                  ps[6].StringParameter,
		Sub:                  ps[7].StringParameter,
		SubCs:                ps[8].IntegerParameter,
		Body:                 ps[9].StringParameter,
		PartCount:            ps[10].IntegerParameter,
		CtT:                  ps[11].StringParameter,
		ContentLocation:      ps[12].StringParameter,
		RecipientID:          ps[13].StringParameter,
		AddressDeviceID:      ps[14].IntegerParameter,
		Expiry:               ps[15].IntegerParameter,
		MCls:                 ps[16].StringParameter,
		MessageType:          ps[17].IntegerParameter,
		V:                    ps[18].IntegerParameter,
		MessageSize:          ps[19].IntegerParameter,
		Pri:                  ps[20].IntegerParameter,
		Rr:                   ps[21].IntegerParameter,
		RptA:                 ps[22].IntegerParameter,
		RespSt:               ps[23].IntegerParameter,
		Status:               ps[24].IntegerParameter,
		TransactionID:        ps[25].StringParameter,
		RetrSt:               ps[26].IntegerParameter,
		RetrTxt:              ps[27].StringParameter,
		RetrTxtCs:            ps[28].IntegerParameter,
		ReadStatus:           ps[29].IntegerParameter,
		CtCls:                ps[30].IntegerParameter,
		RespTxt:              ps[31].StringParameter,
		DTm:                  ps[32].IntegerParameter,
		DeliveryReceiptCount: ps[33].GetIntegerParameter(),
		MismatchedIdentities: ps[34].StringParameter,
		NetworkFailure:       ps[35].StringParameter,
		DRpt:                 ps[36].IntegerParameter,
		SubscriptionID:       ps[37].GetIntegerParameter(),
		ExpiresIn:            ps[38].GetIntegerParameter(),
		ExpireStarted:        ps[39].GetIntegerParameter(),
		Notified:             ps[40].GetIntegerParameter(),
		ReadReceiptCount:     ps[41].GetIntegerParameter(),
		//QuoteID:              ps[42].GetIntegerParameter(),
		//QuoteAuthor:          ps[43].StringParameter,
		//QuoteBody:            ps[44].StringParameter,
		//QuoteAttachment:      ps[45].GetIntegerParameter(),
		//QuoteMissing:         ps[46].GetIntegerParameter(),
		//SharedContacts:       ps[47].StringParameter,
		//Unidentified:         ps[48].GetIntegerParameter(),
	}

	return result
}

// SQLPart info
//
// https://github.com/signalapp/Signal-Android/blob/master/src/org/thoughtcrime/securesms/database/AttachmentDatabase.java#L120
type SQLPart struct {
	RowID                uint64 // primary
	MmsID                *uint64
	Seq                  uint64 // default 0
	ContentType          *string
	Name                 *string
	Chset                *uint64
	ContentDisposition   *string
	Fn                   *string
	Cid                  *string
	ContentLocation      *string
	CttS                 *uint64
	CttT                 *string
	encrypted            *uint64
	TransferState        *uint64
	Data                 *string
	Size                 *uint64
	FileName             *string
	Thumbnail            *string
	ThumbnailAspectRatio *float64
	UniqueID             uint64 // not null
	Digest               []byte
	FastPreflightID      *string
	VoiceNote            uint64 // default 0
	DataRandom           []byte
	ThumbnailRandom      []byte
	Quote                uint64  // default 0
	Width                uint64  // default 0
	Height               uint64  // default 0
	Caption              *string //default null
	StickerPackID        *string //default null
	StickerPackKey       *string //default null
	StickerID            *uint64 //default -1
	DataHash             *string //default null
	BlurHash             *string //default null
	TransformProperties  *string //default null
}

// StatementToPart converts a of SQL statement to a single part.
func StatementToPart(sql *signal.SqlStatement) *SQLPart {
	return ParametersToPart(sql.GetParameters())
}

// ParametersToPart converts a set of SQL parameters to a single part.
func ParametersToPart(ps []*signal.SqlStatement_SqlParameter) *SQLPart {
	// TODO: update to 35 ? 28 ?
	if len(ps) < 25 {
		return nil
	}
	return &SQLPart{
		RowID:                ps[0].GetIntegerParameter(),
		MmsID:                ps[1].IntegerParameter,
		Seq:                  ps[2].GetIntegerParameter(),
		ContentType:          ps[3].StringParameter,
		Name:                 ps[4].StringParameter,
		Chset:                ps[5].IntegerParameter,
		ContentDisposition:   ps[6].StringParameter,
		Fn:                   ps[7].StringParameter,
		Cid:                  ps[8].StringParameter,
		ContentLocation:      ps[9].StringParameter,
		CttS:                 ps[10].IntegerParameter,
		CttT:                 ps[11].StringParameter,
		encrypted:            ps[12].IntegerParameter,
		TransferState:        ps[13].IntegerParameter,
		Data:                 ps[14].StringParameter,
		Size:                 ps[15].IntegerParameter,
		FileName:             ps[16].StringParameter,
		Thumbnail:            ps[17].StringParameter,
		ThumbnailAspectRatio: ps[18].DoubleParameter,
		UniqueID:             ps[19].GetIntegerParameter(),
		Digest:               ps[20].GetBlobParameter(),
		FastPreflightID:      ps[21].StringParameter,
		VoiceNote:            ps[22].GetIntegerParameter(),
		DataRandom:           ps[23].GetBlobParameter(),
		ThumbnailRandom:      ps[24].GetBlobParameter(),
		Quote:                ps[25].GetIntegerParameter(),
		Width:                ps[26].GetIntegerParameter(),
		Height:               ps[27].GetIntegerParameter(),
		//Caption:              ps[28].StringParameter,
		//StickerPackID			ps[29].StringParameter,
		//StickerPackKey        ps[30].StringParameter,
		//StickerID				ps[31].StringParameter,
		//DataHash				ps[32].StringParameter,
		//BlurHash				ps[33].StringParameter,
		//TransformProperties	ps[34].StringParameter,
	}
}
