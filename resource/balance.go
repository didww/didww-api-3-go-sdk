package resource

// Balance represents a DIDWW account balance.
type Balance struct {
	ID           string `json:"-" jsonapi:"balance"`
	TotalBalance string `json:"total_balance"`
	Credit       string `json:"credit"`
	Balance      string `json:"balance"`
}
