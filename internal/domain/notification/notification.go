package notification

import "time"

type Tone string

const (
	ToneTeal   Tone = "teal"
	ToneAmber  Tone = "amber"
	ToneRed    Tone = "red"
	ToneGreen  Tone = "green"
	ToneViolet Tone = "violet"
	ToneGrey   Tone = "grey"
)

// Notification with a nil RecipientUserID is a broadcast to every user;
// otherwise it targets one specific user.
type Notification struct {
	ID              string
	RecipientUserID *int64
	Tone            Tone
	Icon            string
	Title           string
	Message         string
	CreatedAt       time.Time
	Read            bool
}
