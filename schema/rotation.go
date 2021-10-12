package schema

import (
	"fmt"
	"io"

	fromprovider "github.com/grezar/revolver/provider/from"
	toprovider "github.com/grezar/revolver/provider/to"

	"github.com/goccy/go-yaml"
	"github.com/pkg/errors"
)

// LoadRotations loads YAML and extracts rotations
func LoadRotations(r io.Reader) ([]*Rotation, error) {
	var rotations []*Rotation
	d := yaml.NewDecoder(r, yaml.UseOrderedMap(), yaml.Strict())
	if err := d.Decode(&rotations); err != nil {
		return nil, errors.Wrap(err, "failed to decode YAML")
	}
	return rotations, nil
}

type Rotation struct {
	Name string `yaml:"name"`
	From From   `yaml:"from"`
	To   []*To  `yaml:"to"`
}

type FromUnmarshaler From

type From struct {
	Provider string           `yaml:"provider"`
	Spec     FromProviderSpec `yaml:"spec"`
}

type ToUnmarshaler To

type To struct {
	Provider string         `yaml:"provider"`
	Spec     ToProviderSpec `yaml:"spec"`
}

type FromProviderSpec struct {
	fromprovider.Operator
	bytes []byte
}

type ToProviderSpec struct {
	toprovider.Operator
	bytes []byte
}

// UnmarshalYAML implements yaml.Unmarshaler interface
func (s *FromProviderSpec) UnmarshalYAML(bytes []byte) error {
	s.bytes = bytes
	return nil
}

// UnmarshalYAML implements yaml.Unmarshaler interface
func (s *ToProviderSpec) UnmarshalYAML(bytes []byte) error {
	s.bytes = bytes
	return nil
}

// UnmarshalYAML implements yaml.Unmarshaler interface
func (f *From) UnmarshalYAML(unmarshal func(interface{}) error) error {
	if err := unmarshal((*FromUnmarshaler)(f)); err != nil {
		return err
	}
	p := fromprovider.Get(f.Provider)
	if p == nil {
		return errors.New(fmt.Sprintf("unknown provider passed: %s", f.Provider))
	}
	if f.Spec.bytes != nil {
		operator, err := p.UnmarshalSpec(f.Spec.bytes)
		if err != nil {
			return err
		}
		f.Spec.Operator = operator
	}
	return nil
}

// UnmarshalYAML implements yaml.Unmarshaler interface
func (t *To) UnmarshalYAML(unmarshal func(interface{}) error) error {
	if err := unmarshal((*ToUnmarshaler)(t)); err != nil {
		return err
	}
	p := toprovider.Get(t.Provider)
	if p == nil {
		return errors.New(fmt.Sprintf("unknown provider passed: %s", t.Provider))
	}
	if t.Spec.bytes != nil {
		operator, err := p.UnmarshalSpec(t.Spec.bytes)
		if err != nil {
			return err
		}
		t.Spec.Operator = operator
	}
	return nil
}
