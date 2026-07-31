package jsonld_test

import (
	"testing"

	"github.com/afinana/go-dataspace-components/internal/pkg/jsonld"
)

func TestContextArrays(t *testing.T) {
	dspCtx := jsonld.DSPContextArray()
	if len(dspCtx) != 3 {
		t.Fatalf("expected 3 DSP contexts, got %d", len(dspCtx))
	}
	if dspCtx[0] != jsonld.DSP2025Context {
		t.Errorf("expected %s, got %s", jsonld.DSP2025Context, dspCtx[0])
	}
	if dspCtx[1] != jsonld.DCATContext {
		t.Errorf("expected %s, got %s", jsonld.DCATContext, dspCtx[1])
	}
	if dspCtx[2] != jsonld.ODRLContext {
		t.Errorf("expected %s, got %s", jsonld.ODRLContext, dspCtx[2])
	}

	mgmtCtx := jsonld.MgmtContextArray()
	if len(mgmtCtx) != 1 {
		t.Fatalf("expected 1 Mgmt context, got %d", len(mgmtCtx))
	}
	if mgmtCtx[0] != jsonld.EDCMgmtContext {
		t.Errorf("expected %s, got %s", jsonld.EDCMgmtContext, mgmtCtx[0])
	}
}

func TestConstants(t *testing.T) {
	if jsonld.TypeCatalogRequestMessage != "dspace:CatalogRequestMessage" {
		t.Errorf("unexpected TypeCatalogRequestMessage: %s", jsonld.TypeCatalogRequestMessage)
	}
	if jsonld.TypeContractRequestMessage != "dspace:ContractRequestMessage" {
		t.Errorf("unexpected TypeContractRequestMessage: %s", jsonld.TypeContractRequestMessage)
	}
	if jsonld.TypeTransferRequestMessage != "dspace:TransferRequestMessage" {
		t.Errorf("unexpected TypeTransferRequestMessage: %s", jsonld.TypeTransferRequestMessage)
	}
}
