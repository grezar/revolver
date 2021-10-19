package revolver

import (
	"context"
	"os"

	"github.com/grezar/revolver/schema"
	"github.com/grezar/revolver/secrets"
)

type Runner struct {
	config string
}

func NewRunner(path string) *Runner {
	return &Runner{
		config: path,
	}
}

func (r *Runner) Run() error {
	f, err := os.Open(r.config)
	if err != nil {
		return err
	}
	defer f.Close()
	rotations, err := schema.LoadRotations(f)
	if err != nil {
		return err
	}
	for _, rn := range rotations {
		ctx := context.Background()

		renewedSecrets, err := rn.From.Spec.Operator.RenewKey(ctx)
		if err != nil {
			return err
		}

		ctx = secrets.WithSecrets(ctx, renewedSecrets)

		for _, to := range rn.To {
			err := to.Spec.Operator.UpdateSecret(ctx)
			if err != nil {
				return err
			}
		}

		err = rn.From.Spec.DeleteKey(ctx)
		if err != nil {
			return err
		}
	}

	return nil
}
