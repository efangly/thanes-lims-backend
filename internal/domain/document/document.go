package document

import "time"

type Type string

const (
	TypeSOP    Type = "sop"
	TypeManual Type = "manual"
	TypePolicy Type = "policy"
	TypeForm   Type = "form"
	TypeRecord Type = "record"
)

func (t Type) Valid() bool {
	switch t {
	case TypeSOP, TypeManual, TypePolicy, TypeForm, TypeRecord:
		return true
	default:
		return false
	}
}

type Document struct {
	ID          string
	Name        string
	Type        Type
	Version     string
	CreatedBy   string
	IssuedAt    time.Time
	AccessLevel string
	Locked      bool
	StorageKey  string
}

type DocHistory struct {
	ID         int64
	DocumentID string
	Version    string
	Change     string
	Date       time.Time
	Who        string
}
