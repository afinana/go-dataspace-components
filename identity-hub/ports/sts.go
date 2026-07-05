package ports

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// SecurityTokenService handles generating self-issued ID tokens.
type SecurityTokenService struct {
	participantDID string
	keyID          string
	privateKey     *rsa.PrivateKey
}

// NewSecurityTokenService creates a new STS instance.
func NewSecurityTokenService(participantDID string, keyID string, privateKey *rsa.PrivateKey) *SecurityTokenService {
	return &SecurityTokenService{
		participantDID: participantDID,
		keyID:          keyID,
		privateKey:     privateKey,
	}
}

// TokenClaims represents claims inside the Self-Issued ID Token.
type TokenClaims struct {
	Issuer    string `json:"iss"`
	Subject   string `json:"sub"`
	Audience  string `json:"aud"`
	Expiry    int64  `json:"exp"`
	IssuedAt  int64  `json:"iat"`
	JWTID     string `json:"jti"`
}

// GenerateSelfIssuedToken generates and signs a short-lived Self-Issued JWT Token.
func (s *SecurityTokenService) GenerateSelfIssuedToken(audience string, duration time.Duration) (string, error) {
	// 1. Build JWS Header
	header := map[string]string{
		"alg": "RS256",
		"typ": "JWT",
		"kid": s.keyID,
	}

	headerJSON, err := json.Marshal(header)
	if err != nil {
		return "", fmt.Errorf("failed to marshal header: %w", err)
	}

	// 2. Build Payload Claims
	now := time.Now()
	claims := TokenClaims{
		Issuer:    s.participantDID,
		Subject:   s.participantDID,
		Audience:  audience,
		Expiry:    now.Add(duration).Unix(),
		IssuedAt:  now.Unix(),
		JWTID:     fmt.Sprintf("jti-%d", now.UnixNano()),
	}

	claimsJSON, err := json.Marshal(claims)
	if err != nil {
		return "", fmt.Errorf("failed to marshal claims: %w", err)
	}

	// 3. Base64URL encode header and claims
	encodedHeader := base64.RawURLEncoding.EncodeToString(headerJSON)
	encodedClaims := base64.RawURLEncoding.EncodeToString(claimsJSON)

	// 4. Create signing input
	signingInput := encodedHeader + "." + encodedClaims

	// 5. Sign the input using RSASSA-PKCS1-V1_5 with SHA-256
	hasher := sha256.New()
	hasher.Write([]byte(signingInput))
	hashed := hasher.Sum(nil)

	signature, err := rsa.SignPKCS1v15(rand.Reader, s.privateKey, crypto.SHA256, hashed)
	if err != nil {
		return "", fmt.Errorf("failed to sign token signature: %w", err)
	}

	encodedSignature := base64.RawURLEncoding.EncodeToString(signature)

	// 6. Complete JWS token
	return signingInput + "." + encodedSignature, nil
}

// VerifySelfIssuedToken verifies an incoming self-issued ID token.
func VerifySelfIssuedToken(token string, publicKey *rsa.PublicKey) (*TokenClaims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid token formatting")
	}

	signingInput := parts[0] + "." + parts[1]
	signatureBytes, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return nil, fmt.Errorf("failed to decode signature: %w", err)
	}

	hasher := sha256.New()
	hasher.Write([]byte(signingInput))
	hashed := hasher.Sum(nil)

	// Verify signature
	err = rsa.VerifyPKCS1v15(publicKey, crypto.SHA256, hashed, signatureBytes)
	if err != nil {
		return nil, fmt.Errorf("invalid token signature check: %w", err)
	}

	// Decode claims
	claimsBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("failed to decode claims block: %w", err)
	}

	var claims TokenClaims
	if err := json.Unmarshal(claimsBytes, &claims); err != nil {
		return nil, fmt.Errorf("failed to parse claims payload: %w", err)
	}

	// Verify Expiration
	if time.Now().Unix() > claims.Expiry {
		return nil, fmt.Errorf("token expired")
	}

	// Verify self-issued condition (iss == sub)
	if claims.Issuer != claims.Subject {
		return nil, fmt.Errorf("token is not self-issued (iss != sub)")
	}

	return &claims, nil
}
