package token

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var ErrExpiredToken = errors.New("token has expired")

type Payload struct {
	ID        uuid.UUID `json:"id"`
	Username  string    `json:"username"`
	IssuedAt  time.Time `json:"iat"`
	ExpiredAt time.Time `json:"exp"`
	Role      string    `json:"role,omitempty"`
}

func NewPayload(username string, role string, duration time.Duration) (*Payload, error) {
	tokenID, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}

	payload := &Payload{
		ID:        tokenID,
		Username:  username,
		IssuedAt:  time.Now(),
		Role:      role,
		ExpiredAt: time.Now().Add(duration),
	}

	return payload, nil
}

func (p *Payload) Valid() error {
	if time.Now().After(p.ExpiredAt) {
		return ErrExpiredToken
	}
	return nil
}
