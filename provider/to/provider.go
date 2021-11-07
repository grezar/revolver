//go:generate mockgen -source=$GOFILE -destination=mocks/$GOFILE
package toprovider

import "context"

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
	Do(ctx context.Context) error
}
