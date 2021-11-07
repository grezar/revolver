package revolver

import (
	"context"
	"os"
	"sync"

	_ "github.com/grezar/revolver/provider/from/awsiamuser"
	_ "github.com/grezar/revolver/provider/to/awssharedcredentials"
	_ "github.com/grezar/revolver/provider/to/circleci"
	_ "github.com/grezar/revolver/provider/to/tfe"
	"github.com/grezar/revolver/schema"
	"github.com/grezar/revolver/secrets"
	log "github.com/sirupsen/logrus"
)

type Runner struct {
	rotations []*schema.Rotation
}

func NewRunner(path string) (*Runner, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	rotations, err := schema.LoadRotations(f)
	if err != nil {
		return nil, err
	}

	return &Runner{
		rotations: rotations,
	}, nil
}

func (r *Runner) Run() error {
	for _, rn := range r.rotations {
		log.Infof("Start %s\n", rn.Name)

		ctx := context.Background()

		renewedSecrets, err := rn.From.Spec.Operator.Do(ctx)
		if err != nil {
			log.WithFields(log.Fields{
				"provider": rn.From.Provider,
			}).Error(err)
			continue
		}

		// Skip the following operations if the secrets aren't renewed
		if len(renewedSecrets) == 0 {
			continue
		}

		// Ensure that the cleanup process is invoked when the provider's Do
		// operation succeeds
		defer func() {
			err = rn.From.Spec.Cleanup(ctx)
			if err != nil {
				log.WithFields(log.Fields{
					"provider": rn.From.Provider,
				}).Error(err)
			}
		}()

		ctx = secrets.WithSecrets(ctx, renewedSecrets)

		var wg sync.WaitGroup
		for _, to := range rn.To {
			wg.Add(1)
			go func(to *schema.To) {
				defer wg.Done()
				err := to.Spec.Operator.Do(ctx)
				if err != nil {
					log.WithFields(log.Fields{
						"provider": to.Provider,
					}).Error(err)
				}
			}(to)
		}
		wg.Wait()

		log.Infof("Finish %s\n", rn.Name)
	}

	return nil
}
