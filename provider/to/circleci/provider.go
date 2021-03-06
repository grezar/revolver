package circleci

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/goccy/go-yaml"
	"github.com/grezar/go-circleci"
	toprovider "github.com/grezar/revolver/provider/to"
	"github.com/grezar/revolver/secrets"
	"go.uber.org/ratelimit"
)

const (
	name                     = "CircleCI"
	revolverCircleCITokenKey = "REVOLVER_CIRCLECI_TOKEN"
	// Ref: https://circleci.com/docs/2.0/api-developers-guide/#rate-limits
	// Since there are different protections and limits in place for different parts of
	// the API, we cannot use a concrete value for this so I assumed this value is enough
	// small to avoid rate limiting errors.
	apiRateLimit = 5
)

func init() {
	toprovider.Register(&CircleCI{
		RateLimit: ratelimit.New(apiRateLimit),
	})
}

func (t *CircleCI) Name() string {
	return name
}

// toprovider.Provider
type CircleCI struct {
	Token     string
	RateLimit ratelimit.Limiter
}

func (t *CircleCI) UnmarshalSpec(bytes []byte) (toprovider.Operator, error) {
	var s Spec
	if err := yaml.Unmarshal(bytes, &s); err != nil {
		return nil, err
	}
	s.RateLimit = t.RateLimit
	return &s, nil
}

// toprovider.Operator
type Spec struct {
	Owner            string             `yaml:"owner"`
	ProjectVariables []*ProjectVariable `yaml:"projectVariables"`
	Contexts         []*Context         `yaml:"contexts"`
	Client           *circleci.Client
	RateLimit        ratelimit.Limiter
}

type ProjectVariable struct {
	Project   string      `yaml:"project"`
	Variables []*Variable `yaml:"variables"`
}

type Context struct {
	Name      string      `yaml:"name"`
	Variables []*Variable `yaml:"variables"`
}

type Variable struct {
	Name  string `yaml:"name"`
	Value string `yaml:"value"`
}

func (s *Spec) Summary() string {
	summary := fmt.Sprintf("owner: %s", s.Owner)

	if len(s.Contexts) > 0 {
		var contexts []string
		for _, c := range s.Contexts {
			contexts = append(contexts, c.Name)
		}
		summary += fmt.Sprintf(", contexts: %s", strings.Join(contexts, ", "))
	}

	if len(s.ProjectVariables) > 0 {
		var projectVariables []string
		for _, pv := range s.ProjectVariables {
			projectVariables = append(projectVariables, pv.Project)
		}
		summary += fmt.Sprintf(", projects: %s", strings.Join(projectVariables, ", "))
	}

	return summary
}

func (s *Spec) buildClient() (*circleci.Client, error) {
	config := &circleci.Config{
		Token: os.Getenv(revolverCircleCITokenKey),
	}
	client, err := circleci.NewClient(config)
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

	// Update project variables if any
	if len(s.ProjectVariables) > 0 {
		err = s.UpdateProjectVariables(ctx, dryRun, api, s.RateLimit)
		if err != nil {
			return err
		}
	}

	// Update context variables if any
	if len(s.Contexts) > 0 {
		err = s.UpdateContexts(ctx, dryRun, api, s.RateLimit)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Spec) UpdateProjectVariables(ctx context.Context, dryRun bool, api *circleci.Client, ratelimit ratelimit.Limiter) error {
	for _, pv := range s.ProjectVariables {
		ratelimit.Take()
		pvl, err := api.Projects.ListVariables(ctx, pv.Project)
		if err != nil {
			return err
		}

		projectVariableList := make(map[string]*circleci.ProjectVariable)
		for _, pv := range pvl.Items {
			projectVariableList[pv.Name] = pv
		}

		for _, v := range pv.Variables {
			// if the project variable already exists with the same name, delete it before creating a new one
			if projectVariableList[v.Name] != nil {
				if !dryRun {
					ratelimit.Take()
					err := api.Projects.DeleteVariable(ctx, pv.Project, v.Name)
					if err != nil {
						return err
					}
				}
			}

			variableValue, err := secrets.ExecuteTemplate(ctx, v.Value)
			if err != nil {
				return err
			}

			if !dryRun {
				ratelimit.Take()
				_, err = api.Projects.CreateVariable(ctx, pv.Project, circleci.ProjectCreateVariableOptions{
					Name:  circleci.String(v.Name),
					Value: circleci.String(variableValue),
				})
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (s *Spec) UpdateContexts(ctx context.Context, dryRun bool, api *circleci.Client, ratelimit ratelimit.Limiter) error {
	var contexts []*circleci.Context
	var err error
	cl := &circleci.ContextList{
		NextPageToken: "",
	}

	for {
		ratelimit.Take()
		cl, err = api.Contexts.List(ctx, circleci.ContextListOptions{
			OwnerSlug: circleci.String(s.Owner),
			PageToken: circleci.String(cl.NextPageToken),
		})
		if err != nil {
			return err
		}

		contexts = append(contexts, cl.Items...)

		if cl.NextPageToken == "" {
			break
		}
	}

	contextList := make(map[string]*circleci.Context)
	for _, c := range contexts {
		contextList[c.Name] = c
	}

	for _, c := range s.Contexts {
		for _, v := range c.Variables {
			var contextID string
			if v, ok := contextList[c.Name]; ok {
				contextID = v.ID
			} else {
				return fmt.Errorf("circleci context not found: %s", c.Name)
			}

			variableValue, err := secrets.ExecuteTemplate(ctx, v.Value)
			if err != nil {
				return err
			}

			if !dryRun {
				ratelimit.Take()
				_, err = api.Contexts.AddOrUpdateVariable(ctx, contextID, v.Name, circleci.ContextAddOrUpdateVariableOptions{
					Value: circleci.String(variableValue),
				})
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}
