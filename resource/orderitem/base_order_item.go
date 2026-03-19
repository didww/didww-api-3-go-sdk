package orderitem

// BaseOrderItem contains common read-only fields returned by the API for all order item types.
type BaseOrderItem struct {
	Nrc          string  `json:"nrc,omitempty" api:"readonly"`
	Mrc          string  `json:"mrc,omitempty" api:"readonly"`
	ProratedMrc  bool    `json:"prorated_mrc,omitempty" api:"readonly"`
	BilledFrom   *string `json:"billed_from,omitempty" api:"readonly"`
	BilledTo     *string `json:"billed_to,omitempty" api:"readonly"`
	SetupPrice   string  `json:"setup_price,omitempty" api:"readonly"`
	MonthlyPrice string  `json:"monthly_price,omitempty" api:"readonly"`
}
