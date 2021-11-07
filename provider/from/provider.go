//go:generate mockgen -source=$GOFILE -destination=mocks/$GOFILE
package fromprovider

import (
	"context"

	"github.com/grezar/revolver/secrets"
)

var (
	registry = map[string]Provider{}
)

func Register(p Provider) {
	registry[p.Name()] = p
}

func Get(providerType string) Provider {
	return registry[providerType]
}

type Provider interface {
	Name() string
	UnmarshalSpec(bytes []byte) (Operator, error)
}

type Operator interface {
	Do(ctx context.Context) (secrets.Secrets, error)
	Cleanup(ctx context.Context) error
}
