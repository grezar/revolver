package awsiamuser

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/goccy/go-yaml"
	fromprovider "github.com/grezar/revolver/provider/from"
	"github.com/grezar/revolver/secrets"
	str2duration "github.com/xhit/go-str2duration/v2"
)

const (
	name                  = "AWSIAMUser"
	keyAWSAccessKeyID     = "AWSAccessKeyID"
	keyAWSSecretAccessKey = "AWSSecretAccessKey"
)

func init() {
	fromprovider.Register(&AWSIAMUser{})
}

// fromprovider.Provider
type AWSIAMUser struct {}

func (u *AWSIAMUser) Name() string {
	return name
}

func (u *AWSIAMUser) UnmarshalSpec(bytes []byte) (fromprovider.Operator, error) {
	var s Spec
	if err := yaml.Unmarshal(bytes, &s); err != nil {
		return nil, err
	}
	if s.Expiration == "" {
		// default expiration is set to 90 days
		s.Expiration = "90d"
	}
	return &s, nil
}

// fromprovider.Operator
type Spec struct {
	AccountID     string `yaml:"accountId" validate:"required"`
	Username      string `yaml:"username" validate:"required"`
	Expiration    string `yaml:"expiration"`
	DeletableKeys []types.AccessKey
	Client        IAMAccessKeyAPI
}

func (s *Spec) buildClient(ctx context.Context) (IAMAccessKeyAPI, error) {
	if s.Client != nil {
		return s.Client, nil
	}
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion("ap-northeast-1"))
	if err != nil {
		return nil, err
	}
	return iam.NewFromConfig(cfg), nil
}

func (s *Spec) RenewKey(ctx context.Context) (secrets.Secrets, error) {
	client, err := s.buildClient(ctx)
	if err != nil {
		return nil, err
	}

	inpt := &iam.ListAccessKeysInput{
		UserName: aws.String(s.Username),
	}
	keys, err := ListAccessKeys(ctx, client, inpt)
	if err != nil {
		return nil, err
	}

	var keyCreation bool
	expiration, err := str2duration.ParseDuration(s.Expiration)
	if err != nil {
		return nil, err
	}

	switch len(keys.AccessKeyMetadata) {
	case 0:
		keyCreation = true
	case 1:
		if expiration <= time.Since(aws.ToTime(keys.AccessKeyMetadata[0].CreateDate)) {
			keyCreation = true
			s.DeletableKeys = append(s.DeletableKeys, types.AccessKey{
				AccessKeyId: keys.AccessKeyMetadata[0].AccessKeyId,
				UserName:    keys.AccessKeyMetadata[0].UserName,
			})
		}
	case 2:
		// FIXME: improve text format
		fmt.Printf(`[WARN] Two keys were found for AWS IAM User: %s
Revolver cannot create a new key and proceed with the key rotation process,
so manually delete one or more keys and try again.
`, s.Username)
	default:
		return nil, errors.New("Unhandled number of access keys has found")
	}

	if keyCreation {
		input := &iam.CreateAccessKeyInput{
			UserName: aws.String(s.Username),
		}
		output, err := CreateAccessKey(ctx, client, input)
		if err != nil {
			return nil, err
		}

		return secrets.Secrets{
			keyAWSAccessKeyID:     aws.ToString(output.AccessKey.AccessKeyId),
			keyAWSSecretAccessKey: aws.ToString(output.AccessKey.SecretAccessKey),
		}, nil
	}

	return nil, nil
}

func (s *Spec) DeleteKey(ctx context.Context) error {
	client, err := s.buildClient(ctx)
	if err != nil {
		return err
	}
	for _, k := range s.DeletableKeys {
		input := &iam.DeleteAccessKeyInput{
			AccessKeyId: k.AccessKeyId,
			UserName:    k.UserName,
		}
		_, err := DeleteAccessKey(ctx, client, input)
		if err != nil {
			return err
		}
	}
	return nil
}
