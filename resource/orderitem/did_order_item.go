package orderitem

// DidOrderItem represents a DID order item.
type DidOrderItem struct {
	BaseOrderItem
	SkuID              string `json:"sku_id,omitempty"`
	Qty                int    `json:"qty,omitempty"`
	NanpaPrefixID      string `json:"nanpa_prefix_id,omitempty"`
	BillingCyclesCount *int   `json:"billing_cycles_count,omitempty"`
	DIDGroupID         string `json:"did_group_id,omitempty"`
}

func (i *DidOrderItem) orderItemType() string { return typeDidOrderItems }
