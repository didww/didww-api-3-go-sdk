package enums

// AddressVerificationStatus defines the verification status of an address.
type AddressVerificationStatus string

const (
	AddressVerificationStatusPending  AddressVerificationStatus = "pending"
	AddressVerificationStatusApproved AddressVerificationStatus = "approved"
	AddressVerificationStatusRejected AddressVerificationStatus = "rejected"
)

// AreaLevel defines the geographic area level for requirements and DID groups.
type AreaLevel string

const (
	AreaLevelWorldWide AreaLevel = "world_wide"
	AreaLevelCountry   AreaLevel = "country"
	AreaLevelArea      AreaLevel = "area"
	AreaLevelCity      AreaLevel = "city"
)
