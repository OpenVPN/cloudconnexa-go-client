package cloudconnexa

import (
	"encoding/json"
	"testing"
)

func TestAccessGroup_DefaultGroupRoundTrip(t *testing.T) {
	original := AccessGroup{
		ID:           "ag-1",
		Name:         "AG 1",
		DefaultGroup: "default",
	}

	encoded, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var decoded AccessGroup
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if decoded.DefaultGroup != original.DefaultGroup {
		t.Errorf("DefaultGroup round-trip mismatch.\nWant: %q\nGot:  %q", original.DefaultGroup, decoded.DefaultGroup)
	}

	var raw map[string]any
	if err := json.Unmarshal(encoded, &raw); err != nil {
		t.Fatalf("Unmarshal to map failed: %v", err)
	}
	if _, ok := raw["defaultGroup"]; !ok {
		t.Error("Expected JSON key \"defaultGroup\" to be present")
	}
}
