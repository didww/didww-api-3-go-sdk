package resource

import (
	"time"

	"github.com/didww/didww-api-3-go-sdk/v3/resource/enums"
)

// Price lookup rows. Prices and fees are decimal strings in USD. When a
// Reject* field is true the price fields are nil: the destination is refused,
// which is still an answer.
//
// The row ID of an SMS OUT or SMS IN price is derived from its content: it
// changes when the row or its price changes, so it does not track a
// destination over time. A Voice OUT ID stays the same when a re-import
// changes only the price; a Voice IN ID names the answering row of the plan.

// VoiceInPrice is a row of a Voice IN price lookup.
type VoiceInPrice struct {
	ID              string  `json:"-" jsonapi:"prices"`
	Destination     string  `json:"destination"`
	Country         *string `json:"country"`
	SrcCountryName  *string `json:"src_country_name"`
	SourceType      string  `json:"source_type"`
	Rate            *string `json:"rate"`
	ConnectionFee   *string `json:"connection_fee"`
	InitialInterval int     `json:"initial_interval"`
	NextInterval    int     `json:"next_interval"`
	RejectCalls     bool    `json:"reject_calls"`
}

// VoiceOutPrice is a row of a Voice OUT price lookup. The Next* fields and
// EffectiveAt announce a scheduled price change.
type VoiceOutPrice struct {
	ID                     string                      `json:"-" jsonapi:"prices"`
	Category               enums.VoiceOutPriceCategory `json:"category"`
	Country                *string                     `json:"country"`
	Network                *string                     `json:"network"`
	Prefix                 string                      `json:"prefix"`
	Rate                   *string                     `json:"rate"`
	ConnectionFee          *string                     `json:"connection_fee"`
	InitialIncrement       int                         `json:"initial_increment"`
	RecurringIncrement     int                         `json:"recurring_increment"`
	NextRate               *string                     `json:"next_rate"`
	NextConnectionFee      *string                     `json:"next_connection_fee"`
	NextInitialIncrement   *int                        `json:"next_initial_increment"`
	NextRecurringIncrement *int                        `json:"next_recurring_increment"`
	EffectiveAt            *time.Time                  `json:"effective_at"`
	Comment                *string                     `json:"comment"`
	RejectCalls            bool                        `json:"reject_calls"`
}

// SMSOutPrice is a row of an SMS OUT (A2P, P2P or OTP) price lookup.
type SMSOutPrice struct {
	ID                string  `json:"-" jsonapi:"prices"`
	SourcePrefix      *string `json:"source_prefix"`
	DestinationPrefix *string `json:"destination_prefix"`
	Country           *string `json:"country"`
	MCC               *int    `json:"mcc"`
	MCCName           *string `json:"mcc_name"`
	MNC               *int    `json:"mnc"`
	MNCName           *string `json:"mnc_name"`
	RejectSMS         bool    `json:"reject_sms"`
	Price             *string `json:"price"`
}

// SMSInPrice is a row of an SMS IN price lookup.
type SMSInPrice struct {
	ID                         string  `json:"-" jsonapi:"prices"`
	SourcePrefix               *string `json:"source_prefix"`
	SourceNumberLengthMin      int     `json:"source_number_length_min"`
	SourceNumberLengthMax      int     `json:"source_number_length_max"`
	DestinationPrefix          *string `json:"destination_prefix"`
	DestinationNumberLengthMin int     `json:"destination_number_length_min"`
	DestinationNumberLengthMax int     `json:"destination_number_length_max"`
	Country                    *string `json:"country"`
	Price                      string  `json:"price"`
}
