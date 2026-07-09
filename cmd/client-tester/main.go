package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	cp "github.com/afinana/go-dataspace-components/control-plane/domain"
	dp "github.com/afinana/go-dataspace-components/data-plane/domain"
)

func main() {
	fmt.Println("=================================================================")
	fmt.Println("   Sovereign Dataspace Connector - Client Integration Tester    ")
	fmt.Println("=================================================================")
	fmt.Println()

	// Wait briefly for containers to be up if running in docker network
	time.Sleep(1 * time.Second)

	client := &http.Client{Timeout: 10 * time.Second}

	// 1. Query Provider Catalog
	fmt.Println(">>> [1/4] Querying Catalog from Provider Connector...")
	catalogURL := "http://localhost:8081/catalog?requester=did:web:consumer-domain"
	resp, err := client.Get(catalogURL)
	if err != nil {
		fmt.Printf("❌ Failed to query catalog: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("❌ Catalog request returned status: %d\n", resp.StatusCode)
		os.Exit(1)
	}
	
	var catalog map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&catalog); err != nil {
		fmt.Printf("❌ Failed to decode catalog response: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✔ Catalog successfully fetched. Title: %q, Publisher: %q\n", catalog["title"], catalog["publisher"])
	fmt.Println()

	// 2. Perform Contract Negotiation Handshake (Control Plane)
	fmt.Println(">>> [2/4] Triggering Contract Negotiation (DSP Handshake)...")
	negotiationURL := "http://localhost:8081/protocol/negotiation/request"
	reqPayload := map[string]any{
		"id":              "negotiation-request-01",
		"callbackAddress": "http://localhost:8080/callback",
		"offer": map[string]any{
			"id":      "offer-01",
			"assetId": "dataset-asset-01",
			"policy":  nil,
		},
	}
	payloadBytes, _ := json.Marshal(reqPayload)
	resp, err = client.Post(negotiationURL, "application/json", bytes.NewBuffer(payloadBytes))
	if err != nil {
		fmt.Printf("❌ Failed to contact Control Plane: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	var negResponse map[string]any
	json.NewDecoder(resp.Body).Decode(&negResponse)
	fmt.Printf("✔ Contract Negotiation Handshake Complete. ID: %q, Status: %q\n", negResponse["id"], negResponse["status"])
	fmt.Println()

	// 3. Initiate Transfer Signaling with Data Plane (CP -> DP Call simulation)
	fmt.Println(">>> [3/4] Registering Data Flow via Data Plane Signaling Listener...")
	signalingURL := "http://localhost:8082/signaling/start"
	
	// Create signaling request containing HTTP reverse proxy mapping
	flowReq := dp.DataFlowRequest{
		ID:                  "transfer-process-01",
		ContractAgreementID: "agreement-01",
		SourceDataAddress: cp.DataAddress{
			Type: "HttpData",
			Properties: map[string]string{
				// Endpoint we wish to query securely through the proxy
				"endpoint":      "http://control-plane:8081/mock-backend",
				"authType":      "bearer",
				"authHeaderKey": "X-BACKEND-API-KEY",
				"authSecret":    "provider-secret-backend-key-abcde12345",
			},
		},
		DestinationDataAddress: cp.DataAddress{
			Type: "HttpProxy",
		},
		Properties: map[string]string{
			// The token our consumer will use to authenticate with the proxy port
			"auth_token": "consumer-edr-token-12345",
		},
	}
	flowBytes, _ := json.Marshal(flowReq)
	
	resp, err = client.Post(signalingURL, "application/json", bytes.NewBuffer(flowBytes))
	if err != nil {
		fmt.Printf("❌ Failed to trigger signaling: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	var flowResp dp.DataFlowResponse
	json.NewDecoder(resp.Body).Decode(&flowResp)
	if !flowResp.Success {
		fmt.Printf("❌ Signaling rejected by Data Plane: %s\n", flowResp.ErrorDetail)
		os.Exit(1)
	}
	fmt.Printf("✔ Data Flow signaling successfully established. Data Plane ID: %q\n", flowResp.DataPlaneID)
	fmt.Println()

	// 4. Query data via Data Plane API Reverse Proxy
	fmt.Println(">>> [4/4] Fetching Data Assets via Data Plane API Proxy...")
	proxyURL := "http://localhost:8082/public/get?consumer_query=valid"
	
	req, err := http.NewRequest(http.MethodGet, proxyURL, nil)
	if err != nil {
		fmt.Printf("❌ Failed to create request: %v\n", err)
		os.Exit(1)
	}
	// Inject our negotiated EDR authentication token
	req.Header.Set("Authorization", "Bearer consumer-edr-token-12345")

	resp, err = client.Do(req)
	if err != nil {
		fmt.Printf("❌ Proxy fetch failed: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		fmt.Printf("❌ Proxy fetch returned status %d: %s\n", resp.StatusCode, string(bodyBytes))
		os.Exit(1)
	}

	// Decode target payload proxied from httpbin.org containing our dynamic headers
	var backendResult map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&backendResult); err != nil {
		fmt.Printf("❌ Failed to parse backend result: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✔ Egress Data successfully pulled through Proxy!")
	
	// Print a portion of the returned HTTP payload showing the injected headers
	if headers, ok := backendResult["headers"].(map[string]any); ok {
		fmt.Println("Backend Injected Request Headers detected:")
		fmt.Printf("  - X-Backend-Api-Key: %q\n", headers["X-Backend-Api-Key"])
		fmt.Printf("  - Host:              %q\n", headers["Host"])
	}
	fmt.Println()
	fmt.Println("=================================================================")
	fmt.Println("   Integration Test Completed Successfully!                      ")
	fmt.Println("=================================================================")
}
