package installpicker

import (
	"errors"
	"io"

	"github.com/charmbracelet/huh"
)

const installAllOptionValue = "__install_all__"

type Selector interface {
	Select(skills []string) ([]string, error)
}

type Picker struct {
	selector Selector
}

type HuhSelector struct {
	in  io.Reader
	out io.Writer
}

func New(selector Selector) *Picker {
	return &Picker{selector: selector}
}

func NewHuhSelector(in io.Reader, out io.Writer) *HuhSelector {
	return &HuhSelector{in: in, out: out}
}

func (p *Picker) Pick(skills []string) ([]string, error) {
	if len(skills) == 0 {
		return nil, errors.New("no skills available")
	}
	if p == nil || p.selector == nil {
		return nil, errors.New("picker selector is not configured")
	}
	return p.selector.Select(skills)
}

func (s *HuhSelector) Select(skills []string) ([]string, error) {
	if len(skills) == 0 {
		return nil, errors.New("no skills available")
	}

	selected := make([]string, 0, len(skills))
	options := make([]huh.Option[string], 0, len(skills)+1)
	options = append(options, huh.NewOption("Install all", installAllOptionValue))
	for _, skill := range skills {
		options = append(options, huh.NewOption(skill, skill))
	}

	field := huh.NewMultiSelect[string]().
		Title("Select skills to install").
		Options(options...).
		Value(&selected)

	form := huh.NewForm(huh.NewGroup(field))
	if s.in != nil {
		form = form.WithInput(s.in)
	}
	if s.out != nil {
		form = form.WithOutput(s.out)
	}

	if err := form.Run(); err != nil {
		return nil, err
	}

	return resolveSelection(selected, skills), nil
}

func containsValue(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func resolveSelection(selected, skills []string) []string {
	if len(selected) == 0 {
		return skills
	}
	if containsValue(selected, installAllOptionValue) {
		return skills
	}
	return selected
}
