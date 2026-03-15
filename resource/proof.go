package resource

import (
	"encoding/json"
	"time"

	"github.com/didww/didww-api-3-go-sdk/jsonapi"
)

// Proof represents a proof document.
type Proof struct {
	ID        string     `json:"-" jsonapi:"proofs"`
	CreatedAt time.Time  `json:"created_at" api:"readonly"`
	ExpiresAt *time.Time `json:"expires_at" api:"readonly"`
	// Polymorphic entity relationship (type: "identities" or "addresses")
	EntityID   string `json:"-"`
	EntityType string `json:"-"`
	// Other relationship IDs
	ProofTypeID string   `json:"-" rel:"proof_type,proof_types"`
	FileIDs     []string `json:"-" rel:"files,encrypted_files"`
	// Resolved relationships
	ProofType *ProofType `json:"-" rel:"proof_type"`
}

// MarshalRelationships implements RelationshipMarshaler for Proof (polymorphic entity only).
func (p *Proof) MarshalRelationships() (map[string]any, error) {
	rels := make(map[string]any)
	if p.EntityID != "" && p.EntityType != "" {
		rels["entity"] = jsonapi.ToOneRelationship(jsonapi.RelationshipRef{Type: p.EntityType, ID: p.EntityID})
	}
	return rels, nil
}

// UnmarshalRelationships implements RelationshipUnmarshaler for Proof.
// Handles polymorphic entity and proof_type ID extraction from response.
func (p *Proof) UnmarshalRelationships(rels map[string]json.RawMessage) error {
	if raw, ok := rels["entity"]; ok {
		ref, err := jsonapi.ParseToOneRelationship(raw)
		if err != nil {
			return err
		}
		if ref != nil {
			p.EntityID = ref.ID
			p.EntityType = ref.Type
		}
	}
	if raw, ok := rels["proof_type"]; ok {
		ref, err := jsonapi.ParseToOneRelationship(raw)
		if err != nil {
			return err
		}
		if ref != nil {
			p.ProofTypeID = ref.ID
		}
	}
	return nil
}

// ProofType represents a type of proof document.
type ProofType struct {
	ID         string `json:"-" jsonapi:"proof_types"`
	Name       string `json:"name"`
	EntityType string `json:"entity_type"`
}
