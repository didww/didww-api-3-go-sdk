package enums

// AddressVerificationStatus defines the verification status of an address.
type AddressVerificationStatus string

const (
	AddressVerificationStatusPending  AddressVerificationStatus = "Pending"
	AddressVerificationStatusApproved AddressVerificationStatus = "Approved"
	AddressVerificationStatusRejected AddressVerificationStatus = "Rejected"
)

// AreaLevel defines the geographic area level for requirements and DID groups.
type AreaLevel string

const (
	AreaLevelWorldWide AreaLevel = "WorldWide"
	AreaLevelCountry   AreaLevel = "Country"
	AreaLevelArea      AreaLevel = "Area"
	AreaLevelCity      AreaLevel = "City"
)
