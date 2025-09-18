package tokens

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"kabancount/internal/config"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const (
	ScopeAuth = "authentication"
)

type Token struct {
	Plaintext string    `json:"token"`
	Hash      []byte    `json:"-"`
	UserID    uuid.UUID `json:"-"`
	Expiry    time.Time `json:"expiry"`
	Scope     string    `json:"-"`
}

func GenerateToken(userID uuid.UUID, orgID uuid.UUID, ttl time.Duration, scope string) (*Token, error) {
	cfg := config.Get()
	jti, err := generateJTI()
	if err != nil {
		return nil, fmt.Errorf("failed to generate JTI: %w", err)
	}

	now := time.Now()

	claims := jwt.MapClaims{
		"user_id":         userID.String(),
		"organization_id": orgID.String(),
		"scope":           scope,
		"jti":             jti,
		"iat":             now.Unix(),
		"exp":             now.Add(ttl).Unix(),
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := jwtToken.SignedString([]byte(cfg.JWT.Secret))
	if err != nil {
		return nil, fmt.Errorf("failed to sign token: %w", err)
	}

	var token = &Token{
		Plaintext: tokenString,
		Hash:      []byte(tokenString),
		UserID:    userID,
		Expiry:    now.Add(ttl),
		Scope:     scope,
	}

	return token, nil
}

func generateJTI() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
