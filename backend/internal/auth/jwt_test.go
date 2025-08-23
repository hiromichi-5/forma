package auth

import (
	"testing"
	"time"
)

func fixed(t time.Time) func() time.Time { return func() time.Time { return t } }

func TestJWT_IssueAndParse_OK(t *testing.T) {
	base := time.Unix(1_700_000_000, 0)
	s := Signer{Secret: []byte("k"), TTL: time.Hour, Now: fixed(base)}

	tok, err := s.Issue("u-1")
	if err != nil {
		t.Fatalf("issue err: %v", err)
	}
	cl, err := s.Parse(tok)
	if err != nil {
		t.Fatalf("parse err: %v", err)
	}
	if cl.Subject != "u-1" {
		t.Fatalf("want sub=u-1, got %s", cl.Subject)
	}
	if cl.IssuedAt.Time != base {
		t.Fatalf("iat mismatch")
	}
	if cl.ExpiresAt.Time != base.Add(time.Hour) {
		t.Fatalf("exp mismatch")
	}
}

func TestJWT_Expired(t *testing.T) {
	base := time.Unix(1_700_000_000, 0)
	s := Signer{Secret: []byte("k"), TTL: time.Hour, Now: fixed(base)}

	tok, err := s.Issue("u-1")
	if err != nil {
		t.Fatalf("issue err: %v", err)
	}

	s.Now = fixed(base.Add(2 * time.Hour))
	if _, err := s.Parse(tok); err == nil {
		t.Fatalf("want error for expired token")
	}
}
