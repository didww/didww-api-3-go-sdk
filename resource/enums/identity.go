package enums

// IdentityType defines the type of identity (personal, business, or any).
type IdentityType string

const (
	IdentityTypePersonal IdentityType = "Personal"
	IdentityTypeBusiness IdentityType = "Business"
	IdentityTypeAny      IdentityType = "Any"
)
