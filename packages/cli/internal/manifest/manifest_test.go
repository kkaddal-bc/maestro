package manifest

import (
	"strings"
	"testing"
)

func TestParseValidManifest(t *testing.T) {
	got, err := Parse(strings.NewReader(`{
		"version": "v1.2.3",
		"skills": [
			{"name": "maestro-snap", "description": "Capture UI state"}
		],
		"extra": true
	}`))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if got.Version != "v1.2.3" {
		t.Fatalf("Version = %q", got.Version)
	}
	if len(got.Skills) != 1 || got.Skills[0].Name != "maestro-snap" {
		t.Fatalf("Skills = %#v", got.Skills)
	}
}

func TestParseMissingRequiredFields(t *testing.T) {
	tests := []struct {
		name string
		json string
	}{
		{
			name: "missing version",
			json: `{"skills":[{"name":"maestro-snap","description":"x"}]}`,
		},
		{
			name: "missing skills",
			json: `{"version":"v1.2.3"}`,
		},
		{
			name: "missing skill name",
			json: `{"version":"v1.2.3","skills":[{"description":"x"}]}`,
		},
		{
			name: "missing skill description",
			json: `{"version":"v1.2.3","skills":[{"name":"maestro-snap"}]}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := Parse(strings.NewReader(tt.json)); err == nil {
				t.Fatal("Parse() error = nil, want error")
			}
		})
	}
}
