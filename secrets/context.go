package secrets

import (
	"context"
)

type keySecrets struct{}

type Secrets map[string]string

func WithSecrets(ctx context.Context, s Secrets) context.Context {
	return context.WithValue(ctx, keySecrets{}, s)
}

func GetSecrets(ctx context.Context) Secrets {
	ss, ok := ctx.Value(keySecrets{}).(Secrets)
	if ok {
		return ss
	}
	return nil
}
