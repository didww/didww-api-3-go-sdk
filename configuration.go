package didww

// Environment represents a DIDWW API environment.
type Environment string

const (
	Sandbox    Environment = "https://sandbox-api.didww.com/v3"
	Production Environment = "https://api.didww.com/v3"
)
