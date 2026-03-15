package resource

// PublicKey represents a DIDWW public key for encryption.
type PublicKey struct {
	ID  string `json:"-" jsonapi:"public_keys"`
	Key string `json:"key"`
}
