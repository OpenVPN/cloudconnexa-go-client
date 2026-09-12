package cloudconnexa

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestIkeProtocol_AutoInitiateOmitsDeadPeerDetection(t *testing.T) {
	autoInitiate := true
	encoded, err := json.Marshal(IkeProtocol{
		ProtocolVersion: "IKE_V2",
		StartupAction:   "START",
		AutoInitiate:    &autoInitiate,
	})
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// The API rejects a deadPeerDetection object whose fields are unset, so it must be
	// absent entirely when autoInitiate is used.
	if strings.Contains(string(encoded), "deadPeerDetection") {
		t.Errorf("Expected no deadPeerDetection key, got %s", encoded)
	}
	if !strings.Contains(string(encoded), `"autoInitiate":true`) {
		t.Errorf("Expected autoInitiate to be encoded, got %s", encoded)
	}
}

func TestIkeProtocol_DeadPeerDetectionOmitsAutoInitiate(t *testing.T) {
	encoded, err := json.Marshal(IkeProtocol{
		ProtocolVersion:   "IKE_V2",
		StartupAction:     "START",
		DeadPeerDetection: &DeadPeerDetection{TimeoutSec: 30, DeadPeerHandling: "RESTART"},
	})
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	if strings.Contains(string(encoded), "autoInitiate") {
		t.Errorf("Expected no autoInitiate key, got %s", encoded)
	}
	if !strings.Contains(string(encoded), `"deadPeerDetection":{"timeoutSec":30,"deadPeerHandling":"RESTART"}`) {
		t.Errorf("Expected deadPeerDetection to be encoded, got %s", encoded)
	}
}

func TestIkeProtocol_UnmarshalAutoInitiateFalse(t *testing.T) {
	var decoded IkeProtocol
	if err := json.Unmarshal([]byte(`{"autoInitiate":false}`), &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	// A false autoInitiate must stay distinguishable from an absent one.
	if decoded.AutoInitiate == nil {
		t.Fatal("Expected AutoInitiate to be set")
	}
	if *decoded.AutoInitiate {
		t.Error("Expected AutoInitiate to be false")
	}
}
