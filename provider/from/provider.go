package fromprovider

import (
	"github.com/grezar/revolver/repository"
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
	RenewKey() (*repository.Repository, error)
	DeleteKey() error
}
