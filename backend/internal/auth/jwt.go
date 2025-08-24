package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Signer struct {
	Secret []byte
	TTL    time.Duration
	Now    func() time.Time // テスト差し替え用。nilならtime.Now
}

func (s Signer) now() time.Time {
	if s.Now != nil {
		return s.Now()
	}
	return time.Now()
}

func (s Signer) Issue(userID string) (string, error) {
	now := s.now()
	claims := jwt.RegisteredClaims{
		Subject:   userID,
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(s.TTL)),
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return tok.SignedString(s.Secret)
}

func (s Signer) Parse(tokenStr string) (*jwt.RegisteredClaims, error) {
	opts := []jwt.ParserOption{
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithTimeFunc(s.now),
	}
	keyfunc := func(t *jwt.Token) (any, error) { return s.Secret, nil }
	tok, err := jwt.ParseWithClaims(tokenStr, &jwt.RegisteredClaims{}, keyfunc, opts...)
	if err != nil {
		return nil, err
	}
	c, ok := tok.Claims.(*jwt.RegisteredClaims)
	if !ok || !tok.Valid {
		return nil, jwt.ErrTokenInvalidClaims
	}
	return c, nil
}
