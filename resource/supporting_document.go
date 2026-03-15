package resource

import "time"

// SupportingDocumentTemplate represents a supporting document template.
type SupportingDocumentTemplate struct {
	ID        string `json:"-" jsonapi:"supporting_document_templates"`
	Name      string `json:"name"`
	Permanent bool   `json:"permanent"`
	URL       string `json:"url"`
}

// PermanentSupportingDocument represents a permanent supporting document.
type PermanentSupportingDocument struct {
	ID        string    `json:"-" jsonapi:"permanent_supporting_documents"`
	CreatedAt time.Time `json:"created_at" api:"readonly"`
	// Relationship IDs for create/update
	TemplateID string   `json:"-" rel:"template,supporting_document_templates"`
	IdentityID string   `json:"-" rel:"identity,identities"`
	FileIDs    []string `json:"-" rel:"files,encrypted_files"`
	// Resolved relationships
	Template *SupportingDocumentTemplate `json:"-" rel:"template"`
}
