package utils

import "testing"

func TestHashPasswordAndCheck(t *testing.T) {
	hash, err := HashPassword("supersecret123")
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}
	if hash == "supersecret123" {
		t.Fatal("password was not hashed")
	}
	if !CheckPassword(hash, "supersecret123") {
		t.Fatal("CheckPassword should succeed with the correct password")
	}
	if CheckPassword(hash, "wrongpassword") {
		t.Fatal("CheckPassword should fail with an incorrect password")
	}
}

func TestGenerateAndParseToken(t *testing.T) {
	secret := "test-secret"
	token, err := GenerateToken(secret, "user-123", "user@example.com")
	if err != nil {
		t.Fatalf("GenerateToken returned error: %v", err)
	}

	claims, err := ParseToken(secret, token)
	if err != nil {
		t.Fatalf("ParseToken returned error: %v", err)
	}
	if claims.UserID != "user-123" {
		t.Errorf("expected UserID user-123, got %s", claims.UserID)
	}
	if claims.Email != "user@example.com" {
		t.Errorf("expected Email user@example.com, got %s", claims.Email)
	}

	if _, err := ParseToken("wrong-secret", token); err == nil {
		t.Fatal("ParseToken should fail when verified with the wrong secret")
	}
	if _, err := ParseToken(secret, "not-a-real-token"); err == nil {
		t.Fatal("ParseToken should fail on a malformed token")
	}
}

func TestHashIdentifierIsDeterministicAndOneWay(t *testing.T) {
	a := HashIdentifier("user:123")
	b := HashIdentifier("user:123")
	c := HashIdentifier("user:456")

	if a != b {
		t.Fatal("HashIdentifier should be deterministic for the same input")
	}
	if a == c {
		t.Fatal("HashIdentifier should differ for different inputs")
	}
	if a == "user:123" {
		t.Fatal("HashIdentifier should not return the raw input")
	}
}
