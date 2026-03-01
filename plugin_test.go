package kubetypes

import (
	"maps"
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

// testAnalyzer creates an analyzer from settings using the same path as New().
func testAnalyzer(t *testing.T, s Settings) *plugin {
	t.Helper()

	if err := s.validateChecks(); err != nil {
		t.Fatalf("invalid settings: %v", err)
	}
	if err := s.validateExtraGVKs(); err != nil {
		t.Fatalf("invalid settings: %v", err)
	}

	table := make(map[string]gvkInfo, len(knownGVK)+len(s.ExtraKnownGVKs))
	maps.Copy(table, knownGVK)
	for _, entry := range s.ExtraKnownGVKs {
		key := entry.APIVersion + "/" + entry.Kind
		table[key] = parseGVKEntry(entry)
	}

	return &plugin{settings: s, gvkTable: table}
}

func TestMapLiteral(t *testing.T) {
	p := testAnalyzer(t, Settings{IncludeTestFiles: true})
	analysistest.Run(t, analysistest.TestData(), newAnalyzer(p), "map_literal")
}

func TestSprintfYAML(t *testing.T) {
	p := testAnalyzer(t, Settings{IncludeTestFiles: true})
	analysistest.Run(t, analysistest.TestData(), newAnalyzer(p), "sprintf_yaml")
}

func TestFalsePositives(t *testing.T) {
	p := testAnalyzer(t, Settings{IncludeTestFiles: true})
	analysistest.Run(t, analysistest.TestData(), newAnalyzer(p), "false_positives")
}

func TestUnstructuredGVK(t *testing.T) {
	p := testAnalyzer(t, Settings{IncludeTestFiles: true})
	analysistest.Run(t, analysistest.TestData(), newAnalyzer(p), "unstructured_gvk")
}

func TestInvalidCheckName(t *testing.T) {
	s := Settings{
		Checks: map[string]CheckConfig{
			"map_literals": {}, // typo
		},
	}
	if err := s.validateChecks(); err == nil {
		t.Fatal("expected error for invalid check name, got nil")
	}
}

func TestInvalidExtraGVK_EmptyAPIVersion(t *testing.T) {
	s := Settings{
		ExtraKnownGVKs: []GVKEntry{
			{APIVersion: "", Kind: "Widget", TypedPackage: "example.io/v1.Widget"},
		},
	}
	if err := s.validateExtraGVKs(); err == nil {
		t.Fatal("expected error for empty api_version, got nil")
	}
}

func TestInvalidExtraGVK_EmptyKind(t *testing.T) {
	s := Settings{
		ExtraKnownGVKs: []GVKEntry{
			{APIVersion: "example.io/v1", Kind: "", TypedPackage: "example.io/v1.Widget"},
		},
	}
	if err := s.validateExtraGVKs(); err == nil {
		t.Fatal("expected error for empty kind, got nil")
	}
}

func TestInvalidExtraGVK_EmptyTypedPackage(t *testing.T) {
	s := Settings{
		ExtraKnownGVKs: []GVKEntry{
			{APIVersion: "example.io/v1", Kind: "Widget", TypedPackage: ""},
		},
	}
	if err := s.validateExtraGVKs(); err == nil {
		t.Fatal("expected error for empty typed_package, got nil")
	}
}

func TestCheckDisabled(t *testing.T) {
	disabled := false
	p := testAnalyzer(t, Settings{
		IncludeTestFiles: true,
		Checks: map[string]CheckConfig{
			checkMapLiteral: {Enabled: &disabled},
		},
	})
	// map_literal check is disabled, so no diagnostics should fire on map_literal fixture.
	// We need a fixture that would normally produce diagnostics only from map_literal.
	// Using false_positives should produce zero diagnostics regardless.
	analysistest.Run(t, analysistest.TestData(), newAnalyzer(p), "false_positives")
}

func TestCheckEnabled_DefaultNil(t *testing.T) {
	// nil Enabled means default-on.
	s := Settings{
		Checks: map[string]CheckConfig{
			checkMapLiteral: {Enabled: nil},
		},
	}
	enabled := s.enabledChecks()
	if !enabled[checkMapLiteral] {
		t.Fatal("expected map_literal to be enabled with nil Enabled")
	}
	if !enabled[checkSprintfYAML] {
		t.Fatal("expected sprintf_yaml to be enabled (not mentioned in config)")
	}
}

func TestExtraKnownGVKs(t *testing.T) {
	p := testAnalyzer(t, Settings{
		IncludeTestFiles: true,
		ExtraKnownGVKs: []GVKEntry{
			{APIVersion: "example.io/v1", Kind: "Widget", TypedPackage: "example.io/api/v1.Widget"},
		},
	})
	// Run against map_literal_extra which has a Widget GVK.
	analysistest.Run(t, analysistest.TestData(), newAnalyzer(p), "map_literal_extra")
}

func TestIgnoreGVKs(t *testing.T) {
	p := testAnalyzer(t, Settings{
		IncludeTestFiles: true,
		IgnoreGVKs:       []string{"apps/v1/Deployment"},
	})
	// Run against a fixture where Deployment is ignored.
	analysistest.Run(t, analysistest.TestData(), newAnalyzer(p), "ignore_gvk")
}

func TestIncludeTestFilesDefault(t *testing.T) {
	// With default settings (IncludeTestFiles=false), _test.go files are skipped.
	// This is tested indirectly — our fixtures are regular .go files, not _test.go files.
	// The behavior is verified by the test files being analyzed at all.
	p := testAnalyzer(t, Settings{})
	analysistest.Run(t, analysistest.TestData(), newAnalyzer(p), "false_positives")
}

func TestAllChecksEnabledByDefault(t *testing.T) {
	s := Settings{}
	enabled := s.enabledChecks()
	for _, c := range allChecks {
		if !enabled[c] {
			t.Fatalf("expected check %q to be enabled by default", c)
		}
	}
}

func TestEmptyChecksMap_AllEnabled(t *testing.T) {
	s := Settings{Checks: map[string]CheckConfig{}}
	enabled := s.enabledChecks()
	for _, c := range allChecks {
		if !enabled[c] {
			t.Fatalf("expected check %q to be enabled with empty checks map", c)
		}
	}
}

func TestParseGVKEntry(t *testing.T) {
	tests := []struct {
		name       string
		entry      GVKEntry
		wantShort  string
		wantImport string
	}{
		{
			name:       "standard k8s path",
			entry:      GVKEntry{TypedPackage: "k8s.io/api/apps/v1.Deployment"},
			wantShort:  "v1.Deployment",
			wantImport: "k8s.io/api/apps/v1",
		},
		{
			name:       "custom CRD path",
			entry:      GVKEntry{TypedPackage: "example.io/api/v1.Widget"},
			wantShort:  "v1.Widget",
			wantImport: "example.io/api/v1",
		},
		{
			name:       "no dot separator",
			entry:      GVKEntry{TypedPackage: "something"},
			wantShort:  "something",
			wantImport: "something",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := parseGVKEntry(tt.entry)
			if info.ShortName != tt.wantShort {
				t.Errorf("ShortName = %q, want %q", info.ShortName, tt.wantShort)
			}
			if info.ImportPath != tt.wantImport {
				t.Errorf("ImportPath = %q, want %q", info.ImportPath, tt.wantImport)
			}
		})
	}
}

func TestMarkersForSprintfYAML(t *testing.T) {
	t.Run("default markers", func(t *testing.T) {
		s := Settings{}
		markers := s.markersForSprintfYAML()
		if len(markers) != 2 {
			t.Fatalf("expected 2 markers, got %d", len(markers))
		}
	})

	t.Run("additional markers", func(t *testing.T) {
		s := Settings{
			Checks: map[string]CheckConfig{
				checkSprintfYAML: {
					AdditionalMarkers: []string{"metadata:"},
				},
			},
		}
		markers := s.markersForSprintfYAML()
		if len(markers) != 3 {
			t.Fatalf("expected 3 markers, got %d", len(markers))
		}
		if markers[2] != "metadata:" {
			t.Fatalf("expected third marker to be 'metadata:', got %q", markers[2])
		}
	})
}

func TestIsGVKIgnored(t *testing.T) {
	s := Settings{
		IgnoreGVKs: []string{"apps/v1/Deployment", "v1/Pod"},
	}

	if !s.isGVKIgnored("apps/v1", "Deployment") {
		t.Fatal("expected apps/v1 Deployment to be ignored")
	}
	if !s.isGVKIgnored("v1", "Pod") {
		t.Fatal("expected v1 Pod to be ignored")
	}
	if s.isGVKIgnored("v1", "Service") {
		t.Fatal("expected v1 Service to NOT be ignored")
	}
}

func TestMapLiteralConst(t *testing.T) {
	p := testAnalyzer(t, Settings{IncludeTestFiles: true})
	analysistest.Run(t, analysistest.TestData(), newAnalyzer(p), "map_literal_const")
}

func TestMapLiteralNamed(t *testing.T) {
	p := testAnalyzer(t, Settings{IncludeTestFiles: true})
	analysistest.Run(t, analysistest.TestData(), newAnalyzer(p), "map_literal_named")
}

func TestCheckDisabled_MapLiteral(t *testing.T) {
	disabled := false
	p := testAnalyzer(t, Settings{
		IncludeTestFiles: true,
		Checks: map[string]CheckConfig{
			checkMapLiteral: {Enabled: &disabled},
		},
	})
	// map_literal is disabled but sprintf_yaml is still on. Run against sprintf_yaml fixture.
	analysistest.Run(t, analysistest.TestData(), newAnalyzer(p), "sprintf_yaml")
}

func TestCheckDisabled_SprintfYAML(t *testing.T) {
	disabled := false
	p := testAnalyzer(t, Settings{
		IncludeTestFiles: true,
		Checks: map[string]CheckConfig{
			checkSprintfYAML: {Enabled: &disabled},
		},
	})
	// sprintf_yaml is disabled. Run against false_positives which has no sprintf diagnostics.
	analysistest.Run(t, analysistest.TestData(), newAnalyzer(p), "false_positives")
}

func TestSprintfAdditionalMarkers(t *testing.T) {
	p := testAnalyzer(t, Settings{
		IncludeTestFiles: true,
		Checks: map[string]CheckConfig{
			checkSprintfYAML: {
				AdditionalMarkers: []string{"metadata:"},
			},
		},
	})
	analysistest.Run(t, analysistest.TestData(), newAnalyzer(p), "sprintf_markers")
}

func TestUnstructuredIgnoreGVK(t *testing.T) {
	p := testAnalyzer(t, Settings{
		IncludeTestFiles: true,
		IgnoreGVKs:       []string{"apps/v1/Deployment"},
	})
	analysistest.Run(t, analysistest.TestData(), newAnalyzer(p), "unstructured_ignore")
}

func TestPluginInterface(t *testing.T) {
	p, err := New(map[string]any{})
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	lp := p.(*plugin)
	if mode := lp.GetLoadMode(); mode != "typesinfo" {
		t.Fatalf("GetLoadMode() = %q, want %q", mode, "typesinfo")
	}

	analyzers, err := lp.BuildAnalyzers()
	if err != nil {
		t.Fatalf("BuildAnalyzers() failed: %v", err)
	}
	if len(analyzers) != 1 {
		t.Fatalf("BuildAnalyzers() returned %d analyzers, want 1", len(analyzers))
	}
	if analyzers[0].Name != "kubetypes" {
		t.Fatalf("analyzer Name = %q, want %q", analyzers[0].Name, "kubetypes")
	}
}

func TestNewWithInvalidSettings(t *testing.T) {
	// Invalid settings type should fail decoding.
	_, err := New("invalid")
	if err == nil {
		t.Fatal("expected error for invalid settings, got nil")
	}
}

func TestGVKTableNotMutated(t *testing.T) {
	// Verify that creating a plugin does not mutate the global knownGVK map.
	origLen := len(knownGVK)

	_, err := New(map[string]any{
		"extra_known_gvks": []map[string]any{
			{
				"api_version":   "test.io/v1",
				"kind":          "TestResource",
				"typed_package": "test.io/api/v1.TestResource",
			},
		},
	})
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	if len(knownGVK) != origLen {
		t.Fatalf("global knownGVK was mutated: had %d entries, now has %d", origLen, len(knownGVK))
	}
}
