package tfe

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/goccy/go-yaml"
	toprovider "github.com/grezar/revolver/provider/to"
	"github.com/grezar/revolver/secrets"
	tfe "github.com/hashicorp/go-tfe"
	"go.uber.org/ratelimit"
)

const (
	name                                 = "Tfe"
	revolverTfeTokenKey                  = "REVOLVER_TFE_TOKEN"
	categoryEnv         tfe.CategoryType = "env"
	categoryTerraform   tfe.CategoryType = "terraform"
	// Ref: https://www.terraform.io/enterprise/admin/application/general#api-rate-limiting
	apiRateLimit = 30
)

var categoryTypes map[string]tfe.CategoryType = map[string]tfe.CategoryType{
	"env":       categoryEnv,
	"terraform": categoryTerraform,
}

func init() {
	toprovider.Register(&Tfe{
		RateLimit: ratelimit.New(apiRateLimit),
	})
}

func (t *Tfe) Name() string {
	return name
}

// toprovider.Provider
type Tfe struct {
	Token     string
	RateLimit ratelimit.Limiter
}

func (t *Tfe) UnmarshalSpec(bytes []byte) (toprovider.Operator, error) {
	var s Spec
	if err := yaml.Unmarshal(bytes, &s); err != nil {
		return nil, err
	}
	s.RateLimit = t.RateLimit
	return &s, nil
}

// toprovider.Operator
type Spec struct {
	Organization string `yaml:"organization" validate:"required"`
	Workspace    string `yaml:"workspace" validate:"required"`
	Secrets      []Secret
	Client       *tfe.Client
	RateLimit    ratelimit.Limiter
}

type Secret struct {
	Name      string `yaml:"name" validate:"required"`
	Value     string `yaml:"value" validate:"required"`
	Category  string `yaml:"category"`
	Sensitive bool   `yaml:"sensitive"`
}

func (s *Spec) Summary() string {
	return fmt.Sprintf("organization: %s, workspace: %s", s.Organization, s.Workspace)
}

func (s *Spec) buildClient() (*tfe.Client, error) {
	if s.Client != nil {
		return s.Client, nil
	}

	config := &tfe.Config{
		Token: os.Getenv(revolverTfeTokenKey),
	}

	client, err := tfe.NewClient(config)
	if err != nil {
		return nil, err
	}
	return client, nil
}

// Do implements toprovider.Operator interface
func (s *Spec) Do(ctx context.Context, dryRun bool) error {
	api, err := s.buildClient()
	if err != nil {
		return err
	}

	ws, err := api.Workspaces.List(ctx, s.Organization, tfe.WorkspaceListOptions{
		Search: tfe.String(s.Workspace),
	})
	if err != nil {
		return err
	}

	var workspaceID string

	for _, w := range ws.Items {
		if w.Name == s.Workspace {
			workspaceID = w.ID
		}
	}

	if workspaceID == "" {
		return fmt.Errorf("Exactly matching workspace with the name %s was not found", s.Workspace)
	}

	workspaceVariables, err := listWorkspaceVariables(ctx, api, s.RateLimit, workspaceID)
	if err != nil {
		return err
	}

	workspaceVariableList := make(map[string]*tfe.Variable)

	for _, wv := range workspaceVariables {
		for _, item := range wv.Items {
			workspaceVariableList[item.Key] = item
		}
	}

	for _, secret := range s.Secrets {
		categoryType := categoryTypes[secret.Category]
		if categoryType == "" {
			return errors.New("Unsupported category specified. Only \"env\" or \"terraform\" are available")
		}

		secretValue, err := secrets.ExecuteTemplate(ctx, secret.Value)
		if err != nil {
			return err
		}

		wv := workspaceVariableList[secret.Name]
		if wv != nil && (categoryType == wv.Category) {
			if !dryRun {
				s.RateLimit.Take()
				_, err := api.Variables.Update(ctx, workspaceID, wv.ID, tfe.VariableUpdateOptions{
					Key:       tfe.String(secret.Name),
					Value:     tfe.String(secretValue),
					Sensitive: tfe.Bool(secret.Sensitive),
				})
				if err != nil {
					return err
				}
			}
		} else {
			if !dryRun {
				s.RateLimit.Take()
				_, err := api.Variables.Create(ctx, workspaceID, tfe.VariableCreateOptions{
					Key:       tfe.String(secret.Name),
					Value:     tfe.String(secretValue),
					Category:  tfe.Category(categoryType),
					Sensitive: tfe.Bool(secret.Sensitive),
				})
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func listWorkspaceVariables(ctx context.Context, api *tfe.Client, ratelimit ratelimit.Limiter, workspaceID string) ([]*tfe.VariableList, error) {
	var workspaceVariables []*tfe.VariableList

	ratelimit.Take()
	workspaceVariable, err := api.Variables.List(ctx, workspaceID, tfe.VariableListOptions{})
	if err != nil {
		return nil, err
	}
	workspaceVariables = append(workspaceVariables, workspaceVariable)

	for workspaceVariable.CurrentPage < workspaceVariable.NextPage {
		ratelimit.Take()
		workspaceVariable, err = api.Variables.List(ctx, workspaceID, tfe.VariableListOptions{
			ListOptions: tfe.ListOptions{
				PageNumber: workspaceVariable.NextPage,
			},
		})
		if err != nil {
			return nil, err
		}
		workspaceVariables = append(workspaceVariables, workspaceVariable)
	}

	return workspaceVariables, nil
}
