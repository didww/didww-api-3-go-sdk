package resource

// Country represents a country resource.
type Country struct {
	ID     string `json:"-" jsonapi:"countries"`
	Name   string `json:"name"`
	Prefix string `json:"prefix"`
	ISO    string `json:"iso"`
	// Resolved relationships
	Regions []*Region `json:"-" rel:"regions"`
}

// Region represents a geographic region.
type Region struct {
	ID   string  `json:"-" jsonapi:"regions"`
	Name string  `json:"name"`
	ISO  *string `json:"iso"`
	// Resolved relationships
	Country *Country `json:"-" rel:"country"`
}

// City represents a city.
type City struct {
	ID   string `json:"-" jsonapi:"cities"`
	Name string `json:"name"`
	// Resolved relationships
	Country *Country `json:"-" rel:"country"`
	Region  *Region  `json:"-" rel:"region"`
	Area    *Area    `json:"-" rel:"area"`
}

// Area represents a geographic area.
type Area struct {
	ID   string `json:"-" jsonapi:"areas"`
	Name string `json:"name"`
	// Resolved relationships
	Country *Country `json:"-" rel:"country"`
}

// Pop represents a Point of Presence.
type Pop struct {
	ID   string `json:"-" jsonapi:"pops"`
	Name string `json:"name"`
}
