package document

import "time"

type Type string

const (
	TypeSOP    Type = "sop"
	TypeManual Type = "manual"
	TypePolicy Type = "policy"
	TypeForm   Type = "form"
	TypeRecord Type = "record"
	// TypeWarranty tags a warranty document, typically linked to an
	// Equipment via Document.EquipmentID (Phase 5).
	TypeWarranty Type = "warranty"
	// TypeCertificate tags a calibration certificate, typically linked to a
	// CalibrationEvent via Document.CalibrationEventID (Phase 6).
	TypeCertificate Type = "certificate"
)

func (t Type) Valid() bool {
	switch t {
	case TypeSOP, TypeManual, TypePolicy, TypeForm, TypeRecord, TypeWarranty, TypeCertificate:
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
	// EquipmentID optionally links this Document to one Equipment (Phase 5,
	// e.g. a warranty). nil = not linked.
	EquipmentID *string
	// CalibrationEventID optionally links this Document to one
	// CalibrationEvent (Phase 6, e.g. a calibration certificate). nil = not
	// linked.
	CalibrationEventID *int64
}

type DocHistory struct {
	ID         int64
	DocumentID string
	Version    string
	Change     string
	Date       time.Time
	Who        string
}
