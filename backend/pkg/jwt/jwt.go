// Package jwt issues and verifies access tokens.
package jwt

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	gojwt "github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	Role string `json:"role"`
	Name string `json:"name"`
	gojwt.RegisteredClaims
}

func (c Claims) UserID() (int64, error) {
	return strconv.ParseInt(c.Subject, 10, 64)
}

type Manager struct {
	secret []byte
	ttl    time.Duration
	issuer string
}

func NewManager(secret string, ttl time.Duration) *Manager {
	return &Manager{secret: []byte(secret), ttl: ttl, issuer: "pos-api"}
}

// Generate returns a signed access token and its expiry time.
func (m *Manager) Generate(userID int64, role, name string, now time.Time) (string, time.Time, error) {
	exp := now.Add(m.ttl)
	claims := Claims{
		Role: role,
		Name: name,
		RegisteredClaims: gojwt.RegisteredClaims{
			Subject:   strconv.FormatInt(userID, 10),
			Issuer:    m.issuer,
			IssuedAt:  gojwt.NewNumericDate(now),
			ExpiresAt: gojwt.NewNumericDate(exp),
		},
	}
	tok, err := gojwt.NewWithClaims(gojwt.SigningMethodHS256, claims).SignedString(m.secret)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("sign token: %w", err)
	}
	return tok, exp, nil
}

var ErrInvalidToken = errors.New("invalid token")

func (m *Manager) Parse(token string) (*Claims, error) {
	claims := &Claims{}
	_, err := gojwt.ParseWithClaims(token, claims, func(t *gojwt.Token) (any, error) {
		return m.secret, nil
	}, gojwt.WithValidMethods([]string{gojwt.SigningMethodHS256.Alg()}), gojwt.WithIssuer(m.issuer))
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}
	return claims, nil
}
