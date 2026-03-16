package orderitem

// DidOrderItem represents a DID order item.
type DidOrderItem struct {
	SkuID              string  `json:"sku_id,omitempty"`
	Qty                int     `json:"qty,omitempty"`
	Nrc                string  `json:"nrc,omitempty"`
	Mrc                string  `json:"mrc,omitempty"`
	BilledFrom         *string `json:"billed_from,omitempty"`
	BilledTo           *string `json:"billed_to,omitempty"`
	ProratedMrc        bool    `json:"prorated_mrc,omitempty"`
	NanpaPrefixID      string  `json:"nanpa_prefix_id,omitempty"`
	BillingCyclesCount *int    `json:"billing_cycles_count,omitempty"`
	DIDGroupID         string  `json:"did_group_id,omitempty"`
}

func (i *DidOrderItem) orderItemType() string { return "did_order_items" }
