package resource

// NanpaPrefix represents an NANPA prefix.
type NanpaPrefix struct {
	ID  string `json:"-" jsonapi:"nanpa_prefixes"`
	NPA string `json:"npa"`
	NXX string `json:"nxx"`
	// Resolved relationships
	Country *Country `json:"-" rel:"country"`
	Region  *Region  `json:"-" rel:"region"`
}
