package ports_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/afinana/go-dataspace-components/identity-hub/domain"
	"github.com/afinana/go-dataspace-components/identity-hub/ports"
	"github.com/afinana/go-dataspace-components/internal/pkg/kvstore"
)

func TestDIDWebResolver_Resolve(t *testing.T) {
	doc := domain.DIDDocument{
		Context: []string{"https://www.w3.org/ns/did/v1"},
		ID:      "did:web:example.com",
		VerificationMethod: []domain.VerificationMethod{
			{
				ID:         "did:web:example.com#key-1",
				Type:       "JsonWebKey2020",
				Controller: "did:web:example.com",
			},
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(doc)
	}))
	defer ts.Close()

	// Intercept client transport to redirect example.com calls to test server
	client := ts.Client()

	resolver := ports.NewDIDWebResolver(client)
	cache := kvstore.NewMemoryKVStore(time.Minute)
	resolver.WithCache(cache, 0)

	// Invalid DID method
	_, err := resolver.Resolve(context.Background(), "did:key:xyz")
	if err == nil {
		t.Error("expected error for invalid DID method prefix 'did:key:'")
	}

	// Empty domain
	_, err = resolver.Resolve(context.Background(), "did:web:")
	if err == nil {
		t.Error("expected error for empty domain in did:web path")
	}

	// Test Capability Verification
	ok, err := resolver.VerifyCapabilityInvocation(&doc, "did:web:example.com#key-1")
	if err != nil || !ok {
		t.Errorf("expected key to be verified under capabilityInvocation, got ok=%v, err=%v", ok, err)
	}

	ok, err = resolver.VerifyCapabilityInvocation(&doc, "did:web:example.com#non-existent")
	if err == nil || ok {
		t.Error("expected error for non-existent key")
	}
}
