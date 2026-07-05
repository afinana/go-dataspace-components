package ports

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"
	"time"
)

func TestSecurityTokenService_GenerateAndVerify(t *testing.T) {
	// 1. Generate standard RSA keypair for testing
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate test private key: %v", err)
	}

	did := "did:web:consumer.example.com"
	keyID := "did:web:consumer.example.com#key-1"
	sts := NewSecurityTokenService(did, keyID, privateKey)

	// 2. Generate a token
	audience := "did:web:provider.example.com"
	token, err := sts.GenerateSelfIssuedToken(audience, 5*time.Minute)
	if err != nil {
		t.Fatalf("failed to generate self-issued token: %v", err)
	}

	// 3. Verify standard token with valid public key
	claims, err := VerifySelfIssuedToken(token, &privateKey.PublicKey)
	if err != nil {
		t.Fatalf("failed to verify valid token: %v", err)
	}

	if claims.Issuer != did {
		t.Errorf("expected issuer %s, got %s", did, claims.Issuer)
	}
	if claims.Subject != did {
		t.Errorf("expected subject %s, got %s", did, claims.Subject)
	}
	if claims.Audience != audience {
		t.Errorf("expected audience %s, got %s", audience, claims.Audience)
	}

	// 4. Verify invalid signature fails
	wrongPrivKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate wrong private key: %v", err)
	}

	_, err = VerifySelfIssuedToken(token, &wrongPrivKey.PublicKey)
	if err == nil {
		t.Error("expected verification to fail for incorrect public key, but succeeded")
	}
}
