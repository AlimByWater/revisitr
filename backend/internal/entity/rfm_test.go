package entity

import (
	"encoding/json"
	"testing"
)

func TestStandardTemplates_AllPresent(t *testing.T) {
	keys := StandardTemplateKeys()
	if len(keys) != 4 {
		t.Fatalf("expected 4 template keys, got %d", len(keys))
	}
	for _, k := range keys {
		tmpl, ok := StandardTemplates[k]
		if !ok {
			t.Errorf("template %q not found in StandardTemplates", k)
			continue
		}
		if tmpl.Name == "" {
			t.Errorf("template %q has empty name", k)
		}
		if tmpl.Description == "" {
			t.Errorf("template %q has empty description", k)
		}
	}
}

func TestStandardTemplates_ThresholdsValid(t *testing.T) {
	for key, tmpl := range StandardTemplates {
		// R thresholds must be strictly ascending
		for i := 1; i < len(tmpl.RThresholds); i++ {
			if tmpl.RThresholds[i] <= tmpl.RThresholds[i-1] {
				t.Errorf("template %q: R thresholds not ascending at index %d: %v", key, i, tmpl.RThresholds)
			}
		}
		// F thresholds must be strictly descending
		for i := 1; i < len(tmpl.FThresholds); i++ {
			if tmpl.FThresholds[i] >= tmpl.FThresholds[i-1] {
				t.Errorf("template %q: F thresholds not descending at index %d: %v", key, i, tmpl.FThresholds)
			}
		}
		// All values must be positive
		for i, v := range tmpl.RThresholds {
			if v < 0 {
				t.Errorf("template %q: R threshold[%d] is negative: %d", key, i, v)
			}
		}
		for i, v := range tmpl.FThresholds {
			if v < 1 {
				t.Errorf("template %q: F threshold[%d] must be >= 1, got %d", key, i, v)
			}
		}
	}
}

func TestRFMConfig_ActiveTemplate_Standard(t *testing.T) {
	cfg := &RFMConfig{
		ActiveTemplateType: TemplateTypeStandard,
		ActiveTemplateKey:  "coffeegng",
	}
	tmpl, ok := cfg.ActiveTemplate()
	if !ok {
		t.Fatal("expected ActiveTemplate to return ok=true for standard template")
	}
	if tmpl.Key != "coffeegng" {
		t.Errorf("expected key=coffeegng, got %q", tmpl.Key)
	}
	if tmpl.RThresholds != [4]int{3, 7, 14, 30} {
		t.Errorf("unexpected R thresholds: %v", tmpl.RThresholds)
	}
}

func TestRFMConfig_ActiveTemplate_StandardInvalid(t *testing.T) {
	cfg := &RFMConfig{
		ActiveTemplateType: TemplateTypeStandard,
		ActiveTemplateKey:  "nonexistent",
	}
	_, ok := cfg.ActiveTemplate()
	if ok {
		t.Error("expected ok=false for nonexistent standard template")
	}
}

func TestRFMConfig_ActiveTemplate_Custom(t *testing.T) {
	rTh, _ := json.Marshal([4]int{5, 15, 30, 60})
	fTh, _ := json.Marshal([4]int{10, 6, 3, 2})
	name := "Мой формат"

	cfg := &RFMConfig{
		ActiveTemplateType: TemplateTypeCustom,
		ActiveTemplateKey:  "custom",
		CustomTemplateName: &name,
		CustomRThresholds:  rTh,
		CustomFThresholds:  fTh,
	}

	tmpl, ok := cfg.ActiveTemplate()
	if !ok {
		t.Fatal("expected ok=true for custom template")
	}
	if tmpl.Name != "Мой формат" {
		t.Errorf("expected custom name, got %q", tmpl.Name)
	}
	if tmpl.RThresholds != [4]int{5, 15, 30, 60} {
		t.Errorf("unexpected R thresholds: %v", tmpl.RThresholds)
	}
	if tmpl.FThresholds != [4]int{10, 6, 3, 2} {
		t.Errorf("unexpected F thresholds: %v", tmpl.FThresholds)
	}
}

func TestRFMConfig_ActiveTemplate_CustomInvalidJSON(t *testing.T) {
	cfg := &RFMConfig{
		ActiveTemplateType: TemplateTypeCustom,
		ActiveTemplateKey:  "custom",
		CustomRThresholds:  json.RawMessage(`invalid`),
		CustomFThresholds:  json.RawMessage(`[1,2,3,4]`),
	}
	_, ok := cfg.ActiveTemplate()
	if ok {
		t.Error("expected ok=false for invalid JSON thresholds")
	}
}

func TestAllRFMSegments(t *testing.T) {
	segments := AllRFMSegments()
	if len(segments) != 7 {
		t.Fatalf("expected 7 segments, got %d", len(segments))
	}
	// Verify priority order: new first, promising last
	if segments[0] != RFMSegmentNew {
		t.Errorf("expected first segment to be %q, got %q", RFMSegmentNew, segments[0])
	}
	if segments[6] != RFMSegmentPromising {
		t.Errorf("expected last segment to be %q, got %q", RFMSegmentPromising, segments[6])
	}
}

func TestSegmentNames_AllCovered(t *testing.T) {
	for _, seg := range AllRFMSegments() {
		if _, ok := SegmentNames[seg]; !ok {
			t.Errorf("segment %q has no entry in SegmentNames", seg)
		}
	}
}
