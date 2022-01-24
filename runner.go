package revolver

import (
	"context"
	"fmt"
	"os"
	"strconv"

	_ "github.com/grezar/revolver/provider/from/awsiamuser"
	_ "github.com/grezar/revolver/provider/from/stdin"
	_ "github.com/grezar/revolver/provider/to/awssharedcredentials"
	_ "github.com/grezar/revolver/provider/to/circleci"
	_ "github.com/grezar/revolver/provider/to/stdout"
	_ "github.com/grezar/revolver/provider/to/tfe"
	"github.com/grezar/revolver/reporting"
	"github.com/grezar/revolver/schema"
	"github.com/grezar/revolver/secrets"
	"go.uber.org/ratelimit"
)

var revolverRateLimit = 5

type Runner struct {
	rotations []*schema.Rotation
	dryRun    bool
}

func NewRunner(path string, dryRun bool) (*Runner, error) {
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
		dryRun:    dryRun,
	}, nil
}

func (r *Runner) Run(rptr *reporting.R) {
	if v, ok := os.LookupEnv("REVOLVER_RATE_LIMIT"); ok {
		var err error
		revolverRateLimit, err = strconv.Atoi(v)
		if err != nil {
			panic(err)
		}
	}
	rl := ratelimit.New(revolverRateLimit)

	for _, rn := range r.rotations {
		rl.Take()
		rn := rn
		rptr.Run(rn.Name, func(rptr *reporting.R) {
			rptr.Parallel()
			ctx := context.Background()
			// Always run advance dry-run in order not to rotate the from provider's
			// resource when the to provider is unavailable.
			ok := r.run(ctx, rptr, rn, true)
			if !ok {
				return
			}
			if !r.dryRun {
				rptr.ResetChildren()
				_ = r.run(ctx, rptr, rn, false)
			}
		})
	}
}

func (r *Runner) run(ctx context.Context, rptr *reporting.R, rn *schema.Rotation, dryRun bool) bool {
	rptr.Run(fmt.Sprintf("From/%s", rn.From.Provider), func(rptr *reporting.R) {
		rptr.Summary(rn.From.Spec.Operator.Summary())
		newSecrets, err := rn.From.Spec.Operator.Do(ctx, dryRun)
		if err != nil {
			rptr.Fail(err)
			return
		}
		if len(newSecrets) > 0 {
			rptr.Success()
			ctx = secrets.WithSecrets(ctx, newSecrets)
		} else {
			if dryRun {
				rptr.Success()
			} else {
				rptr.Skip()
			}
		}
	})

	for _, to := range rn.To {
		to := to
		rptr.Run(fmt.Sprintf("To/%s", to.Provider), func(rptr *reporting.R) {
			rptr.Parallel()
			rptr.Summary(to.Spec.Operator.Summary())
			if len(secrets.GetSecrets(ctx)) == 0 && !dryRun {
				rptr.Skip()
				return
			}

			err := to.Spec.Operator.Do(ctx, dryRun)
			if err != nil {
				rptr.Fail(err)
				return
			}
			rptr.Success()
		})
	}

	return true
}
