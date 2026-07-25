package domain

import (
	"testing"
	"time"
)

func TestIdentity_DomainModels(t *testing.T) {
	now := time.Now()
	doc := DIDDocument{
		ID:      "did:web:example.com",
		Context: []string{"https://www.w3.org/ns/did/v1"},
		VerificationMethod: []VerificationMethod{
			{
				ID:         "did:web:example.com#key-1",
				Type:       "JsonWebKey2020",
				Controller: "did:web:example.com",
				PublicKeyJwk: map[string]any{
					"kty": "OKP",
					"crv": "Ed25519",
				},
			},
		},
		Authentication: []string{"did:web:example.com#key-1"},
		Service: []DIDService{
			{
				ID:              "did:web:example.com#credential-service",
				Type:            "CredentialService",
				ServiceEndpoint: "http://127.0.0.1:8080/credentials",
				Properties: Metadata{
					"protocol": "dcp",
				},
			},
		},
	}

	vc := VerifiableCredential{
		Context:      []string{"https://www.w3.org/2018/credentials/v1"},
		ID:           "vc-01",
		Type:         []string{"VerifiableCredential", "MembershipCredential"},
		Issuer:       "did:web:issuer.com",
		IssuanceDate: now,
		CredentialSubject: Metadata{
			"holder": "did:web:example.com",
			"status": "ACTIVE",
		},
		Proof: &Proof{
			Type:               "Ed25519Signature2020",
			Created:            now,
			VerificationMethod: "did:web:issuer.com#key-1",
			ProofPurpose:       "assertionMethod",
			ProofValue:         "signature-proof-value",
		},
	}

	vp := VerifiablePresentation{
		Context:              []string{"https://www.w3.org/2018/credentials/v1"},
		ID:                   "vp-01",
		Type:                 []string{"VerifiablePresentation"},
		VerifiableCredential: []VerifiableCredential{vc},
		Proof: &Proof{
			Type:               "Ed25519Signature2020",
			Created:            now,
			VerificationMethod: "did:web:example.com#key-1",
			ProofPurpose:       "authentication",
			ProofValue:         "vp-signature-proof-value",
		},
	}

	pd := PresentationDefinition{
		ID: "pd-01",
		Input: []InputDescriptor{
			{
				ID:      "id-01",
				Purpose: "Verify membership",
				Constraints: &Constraints{
					Fields: []FieldConstraint{
						{
							Path:   []string{"$.credentialSubject.status"},
							Filter: "ACTIVE",
						},
					},
				},
			},
		},
	}

	if doc.ID != "did:web:example.com" {
		t.Errorf("expected DID doc ID 'did:web:example.com', got %s", doc.ID)
	}
	if len(vp.VerifiableCredential) != 1 {
		t.Errorf("expected 1 VC in VP, got %d", len(vp.VerifiableCredential))
	}
	if pd.Input[0].Constraints.Fields[0].Path[0] != "$.credentialSubject.status" {
		t.Errorf("expected constraint path '$.credentialSubject.status'")
	}
}
