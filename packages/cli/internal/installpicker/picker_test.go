package installpicker

import "testing"

type fakeSelector struct {
	got    []string
	result []string
}

func (f *fakeSelector) Select(skills []string) ([]string, error) {
	f.got = append([]string(nil), skills...)
	return append([]string(nil), f.result...), nil
}

func TestPickerDelegatesToSelector(t *testing.T) {
	selector := &fakeSelector{result: []string{"maestro-snap"}}
	picker := New(selector)

	got, err := picker.Pick([]string{"maestro-snap", "other-skill"})
	if err != nil {
		t.Fatalf("Pick() error = %v", err)
	}

	if len(got) != 1 || got[0] != "maestro-snap" {
		t.Fatalf("Pick() = %#v, want maestro-snap", got)
	}

	if len(selector.got) != 2 || selector.got[0] != "maestro-snap" || selector.got[1] != "other-skill" {
		t.Fatalf("selector got %#v", selector.got)
	}
}

func TestResolveSelectionDefaultsToAllSkillsWhenNothingSelected(t *testing.T) {
	skills := []string{"maestro-snap", "other-skill"}

	for _, tc := range []struct {
		name     string
		selected []string
	}{
		{name: "empty submit", selected: []string{}},
		{name: "install all", selected: []string{installAllOptionValue}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := resolveSelection(tc.selected, skills)
			if len(got) != 2 || got[0] != "maestro-snap" || got[1] != "other-skill" {
				t.Fatalf("resolveSelection(%#v, skills) = %#v, want all skills", tc.selected, got)
			}
		})
	}
}

func TestResolveSelectionPreservesExplicitSkillSelection(t *testing.T) {
	skills := []string{"maestro-snap", "other-skill"}

	got := resolveSelection([]string{"other-skill"}, skills)
	if len(got) != 1 || got[0] != "other-skill" {
		t.Fatalf("resolveSelection(%#v, skills) = %#v, want explicit selection", []string{"other-skill"}, got)
	}
}
