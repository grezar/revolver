package stdin

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/goccy/go-yaml"
	fromprovider "github.com/grezar/revolver/provider/from"
	"github.com/grezar/revolver/secrets"
	"github.com/mattn/go-isatty"
)

const (
	name     = "Stdin"
	keyInput = "Input"
)

func init() {
	fromprovider.Register(&Stdin{})
}

// fromprovider.Provider
type Stdin struct{}

func (u *Stdin) Name() string {
	return name
}

func (s *Spec) Summary() string {
	return fmt.Sprintf("")
}

func (u *Stdin) UnmarshalSpec(bytes []byte) (fromprovider.Operator, error) {
	var s Spec
	if err := yaml.Unmarshal(bytes, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

// fromprovider.Operator
type Spec struct{}

func (s *Spec) Do(ctx context.Context) (secrets.Secrets, error) {
	var input string
	if isatty.IsTerminal(os.Stdin.Fd()) {
		fmt.Printf("Input for the Stdin provider: ")
		fmt.Scan(&input)
	} else {
		bytes, err := io.ReadAll(os.Stdin)
		if err != nil {
			return nil, err
		}
		input = string(bytes)
	}
	return secrets.Secrets{
		keyInput: input,
	}, nil
}

func (s *Spec) Cleanup(ctx context.Context) error {
	// No thing to do
	return nil
}