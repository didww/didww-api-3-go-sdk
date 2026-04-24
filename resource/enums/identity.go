package enums

// IdentityType defines the type of identity (personal, business, or any).
type IdentityType string

const (
	IdentityTypePersonal IdentityType = "personal"
	IdentityTypeBusiness IdentityType = "business"
	IdentityTypeAny      IdentityType = "any"
)
