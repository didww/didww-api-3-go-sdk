package enums

// NumberType defines the kind of number a DID group offers, or the kind of
// number a requirement or an emergency calling service applies to.
type NumberType string

const (
	NumberTypeLocal      NumberType = "local"
	NumberTypeNational   NumberType = "national"
	NumberTypeMobile     NumberType = "mobile"
	NumberTypeTollFree   NumberType = "toll_free"
	NumberTypeSharedCost NumberType = "shared_cost"
	NumberTypeGlobal     NumberType = "global"
)
