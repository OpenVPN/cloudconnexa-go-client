package cloudconnexa

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestHost_GatewaysIDsRoundTrip(t *testing.T) {
	original := Host{
		ID:          "host-1",
		Name:        "Host 1",
		GatewaysIDs: []string{"gw-1", "gw-2"},
	}

	encoded, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var decoded Host
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if !reflect.DeepEqual(decoded.GatewaysIDs, original.GatewaysIDs) {
		t.Errorf("GatewaysIDs round-trip mismatch.\nWant: %v\nGot:  %v", original.GatewaysIDs, decoded.GatewaysIDs)
	}

	var raw map[string]any
	if err := json.Unmarshal(encoded, &raw); err != nil {
		t.Fatalf("Unmarshal to map failed: %v", err)
	}
	if _, ok := raw["gatewaysIds"]; !ok {
		t.Error("Expected JSON key \"gatewaysIds\" to be present")
	}
}
