package installpicker

import (
	"errors"
	"io"

	"github.com/charmbracelet/huh"
)

const installAllValue = "__install_all__"

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

	selected := []string{}
	options := make([]huh.Option[string], 0, len(skills)+1)
	options = append(options, huh.NewOption("Install all", installAllValue))
	for _, skill := range skills {
		options = append(options, huh.NewOption(skill, skill))
	}

	field := huh.NewMultiSelect[string]().
		Title("Select skills to install").
		Options(options...).
		Value(&selected).
		Validate(func(values []string) error {
			if len(values) == 0 {
				return errors.New("select at least one skill")
			}
			return nil
		})

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

	if contains(selected, installAllValue) {
		return skills, nil
	}

	return selected, nil
}

func contains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
