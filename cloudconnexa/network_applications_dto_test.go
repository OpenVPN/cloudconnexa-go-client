package cloudconnexa

import (
	"encoding/json"
	"testing"
)

func TestNetworkApplicationRoute_ExactMatchRoundTrip(t *testing.T) {
	original := NetworkApplicationRoute{
		Value:           "example.com",
		AllowEmbeddedIP: true,
		ExactMatch:      true,
	}

	encoded, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var decoded NetworkApplicationRoute
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if decoded != original {
		t.Errorf("Round-trip mismatch.\nWant: %+v\nGot:  %+v", original, decoded)
	}

	var raw map[string]any
	if err := json.Unmarshal(encoded, &raw); err != nil {
		t.Fatalf("Unmarshal to map failed: %v", err)
	}
	if _, ok := raw["exactMatch"]; !ok {
		t.Error("Expected JSON key \"exactMatch\" to be present when set true")
	}
}

func TestNetworkApplicationRoute_ExactMatchOmittedWhenFalse(t *testing.T) {
	encoded, err := json.Marshal(NetworkApplicationRoute{
		Value:           "example.com",
		AllowEmbeddedIP: false,
		ExactMatch:      false,
	})
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var raw map[string]any
	if err := json.Unmarshal(encoded, &raw); err != nil {
		t.Fatalf("Unmarshal to map failed: %v", err)
	}
	if _, ok := raw["exactMatch"]; ok {
		t.Errorf("Expected JSON key \"exactMatch\" to be omitted when false; got: %s", encoded)
	}
}

func TestApplicationRoute_NoExactMatch(t *testing.T) {
	// Host application routes must not carry exactMatch — the schema does
	// not define it for HostApplicationRouteRequest.
	encoded, err := json.Marshal(ApplicationRoute{
		Value:           "example.com",
		AllowEmbeddedIP: true,
	})
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var raw map[string]any
	if err := json.Unmarshal(encoded, &raw); err != nil {
		t.Fatalf("Unmarshal to map failed: %v", err)
	}
	if _, ok := raw["exactMatch"]; ok {
		t.Errorf("Host ApplicationRoute should not include \"exactMatch\"; got: %s", encoded)
	}
}

func TestNetworkApplicationDomainRoute_RoundTrip(t *testing.T) {
	original := NetworkApplicationDomainRoute{
		ID:              "route-1",
		Type:            "DOMAIN",
		Domain:          "example.com",
		AllowEmbeddedIP: true,
		ExactMatch:      true,
	}

	encoded, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var decoded NetworkApplicationDomainRoute
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if decoded != original {
		t.Errorf("Round-trip mismatch.\nWant: %+v\nGot:  %+v", original, decoded)
	}
}
