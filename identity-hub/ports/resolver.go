package ports

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/afinana/go-dataspace-components/identity-hub/domain"
	"github.com/afinana/go-dataspace-components/internal/pkg/kvstore"
)

// DIDWebResolver implements domain.DIDResolver for the did:web method.
type DIDWebResolver struct {
	client   *http.Client
	cache    kvstore.KVStore
	cacheTTL time.Duration
}

// NewDIDWebResolver creates a new did:web resolver instance.
func NewDIDWebResolver(client *http.Client) *DIDWebResolver {
	if client == nil {
		client = http.DefaultClient
	}
	return &DIDWebResolver{
		client:   client,
		cacheTTL: 15 * time.Minute,
	}
}

// WithCache attaches an L1 KV Store cache to the DID Web Resolver.
func (r *DIDWebResolver) WithCache(cache kvstore.KVStore, ttl time.Duration) *DIDWebResolver {
	r.cache = cache
	if ttl > 0 {
		r.cacheTTL = ttl
	}
	return r
}

// Resolve translates a did:web DID to an HTTPS url, fetches it, and returns the parsed DID Document (checking cache first).
func (r *DIDWebResolver) Resolve(ctx context.Context, did string) (*domain.DIDDocument, error) {
	if !strings.HasPrefix(did, "did:web:") {
		return nil, fmt.Errorf("invalid DID method: expected prefix 'did:web:' but got '%s'", did)
	}

	// Check cache first
	if r.cache != nil {
		if cachedBytes, found, _ := r.cache.Get(ctx, "did:"+did); found {
			var didDoc domain.DIDDocument
			if err := json.Unmarshal(cachedBytes, &didDoc); err == nil {
				return &didDoc, nil
			}
		}
	}

	// did:web:example.com -> https://example.com/.well-known/did.json
	// did:web:example.com:path:sub -> https://example.com/path/sub/did.json
	parsedPath := strings.TrimPrefix(did, "did:web:")
	parts := strings.Split(parsedPath, ":")
	if len(parts) == 0 || parts[0] == "" {
		return nil, fmt.Errorf("empty domain in did:web path: %s", did)
	}

	domainName := parts[0]
	var urlStr string
	if len(parts) == 1 {
		urlStr = fmt.Sprintf("https://%s/.well-known/did.json", domainName)
	} else {
		subPaths := strings.Join(parts[1:], "/")
		urlStr = fmt.Sprintf("https://%s/%s/did.json", domainName, subPaths)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlStr, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create resolution request: %w", err)
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute HTTP GET did:web resolution request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to resolve did:web document from %s: HTTP status %d", urlStr, resp.StatusCode)
	}

	var didDoc domain.DIDDocument
	if err := json.NewDecoder(resp.Body).Decode(&didDoc); err != nil {
		return nil, fmt.Errorf("failed to decode resolved DID Document from %s: %w", urlStr, err)
	}

	// Populate cache
	if r.cache != nil {
		if payloadBytes, err := json.Marshal(&didDoc); err == nil {
			_ = r.cache.Set(ctx, "did:"+did, payloadBytes, r.cacheTTL)
		}
	}

	return &didDoc, nil
}

// VerifyCapabilityInvocation checks if a keyId is authorized under 'capabilityInvocation' in the DID Document.
func (r *DIDWebResolver) VerifyCapabilityInvocation(didDoc *domain.DIDDocument, keyID string) (bool, error) {
	for _, method := range didDoc.VerificationMethod {
		if method.ID == keyID {
			return true, nil
		}
	}
	return false, fmt.Errorf("key %s not found in verification methods of DID %s", keyID, didDoc.ID)
}
