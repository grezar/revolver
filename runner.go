package revolver

import (
	"context"
	"os"

	_ "github.com/grezar/revolver/provider/from/awsiamuser"
	_ "github.com/grezar/revolver/provider/to/awssharedcredentials"
	_ "github.com/grezar/revolver/provider/to/circleci"
	_ "github.com/grezar/revolver/provider/to/tfe"
	"github.com/grezar/revolver/schema"
	"github.com/grezar/revolver/secrets"
	log "github.com/sirupsen/logrus"
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

	// TOOD: Transactional operation is needed for safety key rotations
	for _, rn := range rotations {
		log.Infof("Start %s\n", rn.Name)

		ctx := context.Background()

		renewedSecrets, err := rn.From.Spec.Operator.RenewKey(ctx)
		if err != nil {
			log.WithFields(log.Fields{
				"provider": rn.From.Provider,
			}).Error(err)
			continue
		}

		ctx = secrets.WithSecrets(ctx, renewedSecrets)

		for _, to := range rn.To {
			err := to.Spec.Operator.UpdateSecret(ctx)
			if err != nil {
				log.WithFields(log.Fields{
					"provider": to.Provider,
				}).Error(err)
				continue
			}
		}

		err = rn.From.Spec.DeleteKey(ctx)
		if err != nil {
			log.WithFields(log.Fields{
				"provider": rn.From.Provider,
			}).Error(err)
			continue
		}

		log.Infof("Finish %s\n", rn.Name)
	}

	return nil
}
