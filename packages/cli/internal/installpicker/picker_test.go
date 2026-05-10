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
