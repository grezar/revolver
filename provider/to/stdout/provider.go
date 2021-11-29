package stdout

import (
	"context"
	"fmt"

	"github.com/goccy/go-yaml"
	toprovider "github.com/grezar/revolver/provider/to"
	"github.com/grezar/revolver/secrets"
)

const (
	name = "Stdout"
)

func init() {
	toprovider.Register(&Stdout{})
}

func (t *Stdout) Name() string {
	return name
}

// toprovider.Provider
type Stdout struct {
	Token string
}

func (t *Stdout) UnmarshalSpec(bytes []byte) (toprovider.Operator, error) {
	var s Spec
	if err := yaml.Unmarshal(bytes, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

// toprovider.Operator
type Spec struct {
	Output string `yaml:"output"`
}

func (s *Spec) Summary() string {
	return "output to stdout"
}

// Do implements toprovider.Operator interface
func (s *Spec) Do(ctx context.Context) error {
	output, err := secrets.ExecuteTemplate(ctx, s.Output)
	if err != nil {
		return err
	}
	fmt.Print(output)
	return nil
}
