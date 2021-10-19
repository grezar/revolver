package tfe

import (
	"context"
	"errors"
	"os"

	"github.com/goccy/go-yaml"
	toprovider "github.com/grezar/revolver/provider/to"
	"github.com/grezar/revolver/secrets"
	tfe "github.com/hashicorp/go-tfe"
)

const (
	name                                 = "Tfe"
	revolverTfeTokenKey                  = "REVOLVER_TFE_TOKEN"
	categoryEnv         tfe.CategoryType = "env"
	categoryTerraform   tfe.CategoryType = "terraform"
)

var categoryTypes map[string]tfe.CategoryType = map[string]tfe.CategoryType{
	"env":       categoryEnv,
	"terraform": categoryTerraform,
}

func init() {
	toprovider.Register(&Tfe{})
}

func (t *Tfe) Name() string {
	return name
}

// toprovider.Provider
type Tfe struct {
	Token string
}

func (t *Tfe) UnmarshalSpec(bytes []byte) (toprovider.Operator, error) {
	var s Spec
	if err := yaml.Unmarshal(bytes, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

// toprovider.Operator
type Spec struct {
	Organization string `yaml:"owner" validate:"required"`
	Workspace    string `yaml:"workspace" validate:"required"`
	Secrets      []Secret
	Client       *tfe.Client
}

type Secret struct {
	Name     string `yaml:"name" validate:"required"`
	Value    string `yaml:"value" validate:"required"`
	Category string `yaml:"category"`
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

// UpdateSecret implements toprovider.Operator interface
func (s *Spec) UpdateSecret(ctx context.Context) error {
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

	if len(ws.Items) > 1 {
		return errors.New("Multiple workspaces were found. Please specify an exact matching name")
	}

	workspaceID := ws.Items[0].ID

	workspaceVariables, err := listWorkspaceVariables(ctx, api, workspaceID)
	if err != nil {
		return err
	}

	workspaceVariableList := make(map[string]*tfe.Variable)

	for _, wv := range workspaceVariables {
		for _, item := range wv.Items {
			workspaceVariableList[item.Key] = item
		}
	}

	for _, s := range s.Secrets {
		categoryType := categoryTypes[s.Category]
		if categoryType == "" {
			return errors.New("Unsupported category specified. Only \"env\" or \"terraform\" are available")
		}

		secret, err := secrets.ExecuteTemplate(ctx, s.Value)
		if err != nil {
			return err
		}

		wv := workspaceVariableList[s.Name]
		if wv != nil && (categoryType == wv.Category) {
			_, err := api.Variables.Update(ctx, workspaceID, wv.ID, tfe.VariableUpdateOptions{
				Key:       tfe.String(s.Name),
				Value:     tfe.String(secret),
				Sensitive: tfe.Bool(true),
			})
			if err != nil {
				return err
			}
		} else {
			_, err := api.Variables.Create(ctx, workspaceID, tfe.VariableCreateOptions{
				Key:       tfe.String(s.Name),
				Value:     tfe.String(secret),
				Category:  tfe.Category(categoryType),
				Sensitive: tfe.Bool(true),
			})
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func listWorkspaceVariables(ctx context.Context, api *tfe.Client, workspaceID string) ([]*tfe.VariableList, error) {
	var workspaceVariables []*tfe.VariableList

	workspaceVariable, err := api.Variables.List(ctx, workspaceID, tfe.VariableListOptions{})
	if err != nil {
		return nil, err
	}
	workspaceVariables = append(workspaceVariables, workspaceVariable)

	for workspaceVariable.CurrentPage < workspaceVariable.NextPage {
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
