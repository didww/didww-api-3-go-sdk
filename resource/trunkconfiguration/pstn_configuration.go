package trunkconfiguration

// PSTNConfiguration represents a PSTN trunk configuration.
type PSTNConfiguration struct {
	Dst string `json:"dst"`
}

func (c *PSTNConfiguration) ConfigurationType() string { return "pstn_configurations" }
