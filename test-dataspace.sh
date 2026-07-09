#!/bin/bash
# Sovereign Dataspace Connector - E2E Integration Test Suite
set -e

# Color definitions
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}=================================================================${NC}"
echo -e "${BLUE}        Sovereign Dataspace Connector E2E Integration Suite      ${NC}"
echo -e "${BLUE}=================================================================${NC}"
echo ""

# Helper to check commands
check_status() {
  if [ $? -eq 0 ]; then
    echo -e "${GREEN}✔ Success${NC}"
  else
    echo -e "${RED}❌ Failed${NC}"
    exit 1
  fi
}

# 1. Test Identity Hub VC Issuance and Querying
echo -e "${BLUE}>>> [1/5] Testing Identity Hub (did:web / Decentralized Claims)...${NC}"
echo "Registering a new Verifiable Credential..."
curl -s -X POST -H "Content-Type: application/json" -d '{
  "id": "vc-test-membership",
  "issuer": "did:web:issuer-authority",
  "type": ["VerifiableCredential", "XDataShareMembershipCredential"],
  "issuanceDate": "2026-07-06T20:00:00Z"
}' http://localhost:8080/credentials
check_status

echo "Querying Verifiable Presentations for XDataShareMembershipCredential scope..."
curl -s -X POST -H "Content-Type: application/json" -d '{
  "scopes": ["XDataShareMembershipCredential"]
}' http://localhost:8080/presentations/query | grep -q "VerifiablePresentation"
check_status
echo ""

# 2. Test Control Plane DCAT Catalog API
echo -e "${BLUE}>>> [2/5] Querying W3C DCAT-AP Catalog from Control Plane (port 8081)...${NC}"
curl -s http://localhost:8081/catalog | grep -q "catalog-main"
check_status
echo ""

# 3. Test Control Plane DSP Contract Negotiation
echo -e "${BLUE}>>> [3/5] Triggering DSP Contract Negotiation Handshake...${NC}"
curl -s -X POST -H "Content-Type: application/json" -d '{
  "counterPartyAddress": "http://control-plane:8081",
  "counterPartyId": "did:web:provider",
  "policy": {
    "@type": "Offer",
    "@id": "policy-01"
  }
}' http://localhost:8081/protocol/negotiation/request | grep -q "REQUESTED"
check_status
echo ""

# 4. Test Data Plane Control Signaling
echo -e "${BLUE}>>> [4/5] Initiating Data Flow Signaling on Data Plane...${NC}"
curl -s -X POST -H "Content-Type: application/json" -d '{
  "id": "flow-process-test",
  "contractAgreementId": "agreement-test-99",
  "sourceDataAddress": {
    "type": "HttpData",
    "properties": {
      "endpoint": "http://control-plane:8081/mock-backend",
      "authType": "bearer",
      "authHeaderKey": "X-BACKEND-API-KEY",
      "authSecret": "my-secret-key-12345"
    }
  },
  "destinationDataAddress": {
    "type": "HttpProxy"
  },
  "properties": {
    "auth_token": "consumer-test-token"
  }
}' http://localhost:8082/signaling/start | grep -q "success\":true"
check_status
echo ""

# 5. Run Go Integration client-tester suite
echo -e "${BLUE}>>> [5/5] Executing Compiled Client Integration Tester...${NC}"
go run cmd/client-tester/main.go
check_status

echo ""
echo -e "${GREEN}=================================================================${NC}"
echo -e "${GREEN}        Sovereign Dataspace Integration Tests Passed!            ${NC}"
echo -e "${GREEN}=================================================================${NC}"
