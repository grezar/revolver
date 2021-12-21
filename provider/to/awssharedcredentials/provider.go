package awssharedcredentials

import (
	"bufio"
	"context"
	"fmt"
	"os"

	"github.com/goccy/go-yaml"
	toprovider "github.com/grezar/revolver/provider/to"
	"github.com/grezar/revolver/secrets"
	"gopkg.in/ini.v1"
)

const (
	name = "AWSSharedCredentials"
)

var refs = map[string]string{
	"aws_access_key_id":     "{{ .AWSAccessKeyID }}",
	"aws_secret_access_key": "{{ .AWSSecretAccessKey }}",
}

func init() {
	toprovider.Register(&AWSSharedCredentials{})
}

func (a *AWSSharedCredentials) Name() string {
	return name
}

// toprovider.Provider
type AWSSharedCredentials struct{}

func (a *AWSSharedCredentials) UnmarshalSpec(bytes []byte) (toprovider.Operator, error) {
	var s Spec
	if err := yaml.Unmarshal(bytes, &s); err != nil {
		return nil, err
	}
	s.Secrets = refs
	return &s, nil
}

// toprovider.Operator
type Spec struct {
	Path    string `yaml:"path"`
	Profile string `yaml:"profile"`
	Secrets map[string]string
}

func (s *Spec) Summary() string {
	return fmt.Sprintf("path: %s, profile: %s", s.Path, s.Profile)
}

// Do implements toprovider.Operator interface
func (s *Spec) Do(ctx context.Context) error {
	c, err := ini.Load(s.Path)
	if err != nil {
		return err
	}

	for k, v := range s.Secrets {
		secret, err := secrets.ExecuteTemplate(ctx, v)
		if err != nil {
			return err
		}
		c.Section(s.Profile).Key(k).SetValue(secret)
	}

	f, err := os.Create(s.Path)
	if err != nil {
		return err
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	_, err = c.WriteTo(w)
	if err != nil {
		return err
	}
	err = w.Flush()
	if err != nil {
		return err
	}

	return nil
}
