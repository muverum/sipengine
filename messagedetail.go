package sipengine

import "time"

type MessageDetail struct {
	From string
	To string
	Start time.Time
	End time.Time
	Jurisdiction string
	Lata string
	//The number which the provided SIP target resolves to for number portability purposes.
	PortabilityIdentifier string
}
