package tokens

import (
	"time"
)

// Maker is an interface for managing tokens
type Maker interface {
	// CreateToken creates a new tokens for a specific username and duration
	CreateToken(userId int32, duration time.Duration) (string, *Payload, error)

	// VerifyToken checks if the tokens is valid or not
	VerifyToken(token string) (*Payload, error)
}
