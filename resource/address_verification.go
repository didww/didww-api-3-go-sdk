package resource

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/didww/didww-api-3-go-sdk/resource/enums"
)

// AddressVerification represents an address verification request.
type AddressVerification struct {
	ID                 string                          `json:"-" jsonapi:"address_verifications"`
	ServiceDescription *string                         `json:"service_description,omitempty"`
	CallbackURL        *string                         `json:"callback_url,omitempty"`
	CallbackMethod     *string                         `json:"callback_method,omitempty"`
	Status             enums.AddressVerificationStatus `json:"status" api:"readonly"`
	RejectReasons      []string                        `json:"reject_reasons" api:"readonly"`
	CreatedAt          time.Time                       `json:"created_at" api:"readonly"`
	Reference          string                          `json:"reference" api:"readonly"`
	// Relationship IDs for create/update
	AddressID string   `json:"-" rel:"address,addresses"`
	DIDIDs    []string `json:"-" rel:"dids,dids"`
	// Resolved relationships
	AddressRel *Address `json:"-" rel:"address"`
}

// UnmarshalJSON splits the semicolon-separated reject_reasons string into a slice.
func (a *AddressVerification) UnmarshalJSON(data []byte) error {
	type Alias AddressVerification
	aux := &struct {
		RejectReasons *string `json:"reject_reasons"`
		*Alias
	}{
		Alias: (*Alias)(a),
	}
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}
	if aux.RejectReasons != nil {
		rawItems := strings.Split(*aux.RejectReasons, "; ")
		a.RejectReasons = make([]string, 0, len(rawItems))
		for _, item := range rawItems {
			if item != "" {
				a.RejectReasons = append(a.RejectReasons, item)
			}
		}
	} else {
		a.RejectReasons = nil
	}
	return nil
}
