package revolver

import (
	"os"

	"github.com/grezar/revolver/schema"
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
		repo, err := rn.From.Spec.Operator.RenewKey()
		if err != nil {
			return err
		}

		for _, to := range rn.To {
			err := to.Spec.Operator.UpdateSecret(repo)
			if err != nil {
				return err
			}
		}

		err = rn.From.Spec.DeleteKey()
		if err != nil {
			return err
		}
	}

	return nil
}
