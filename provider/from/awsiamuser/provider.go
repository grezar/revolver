package awsiamuser

import (
	"context"
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
	awsDefaultRegion      = "us-east-1"
)

func init() {
	fromprovider.Register(&AWSIAMUser{})
}

// fromprovider.Provider
type AWSIAMUser struct{}

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

func (s *Spec) Summary() string {
	return fmt.Sprintf("account: %s, username: %s", s.AccountID, s.Username)
}

func (s *Spec) buildClient(ctx context.Context) (IAMAccessKeyAPI, error) {
	if s.Client != nil {
		return s.Client, nil
	}
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(awsDefaultRegion))
	if err != nil {
		return nil, err
	}
	return iam.NewFromConfig(cfg), nil
}

func (s *Spec) Do(ctx context.Context, dryRun bool) (secrets.Secrets, error) {
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

	expiration, err := str2duration.ParseDuration(s.Expiration)
	if err != nil {
		return nil, err
	}

	switch len(keys.AccessKeyMetadata) {
	case 0:
		// Only to proceed to the next step.
	case 1:
		if expiration <= time.Since(aws.ToTime(keys.AccessKeyMetadata[0].CreateDate)) {

			s.DeletableKeys = append(s.DeletableKeys, types.AccessKey{
				AccessKeyId: keys.AccessKeyMetadata[0].AccessKeyId,
				UserName:    keys.AccessKeyMetadata[0].UserName,
			})
		} else {
			return nil, nil
		}
	case 2:
		return nil, fmt.Errorf(`The user "%s" already has two access keys. Revolver cannot create a new key and cannot continue with the key rotation process. Please delete at least one of the existing keys and try again.`, s.Username)
	default:
		panic("never reach here")
	}

	input := &iam.CreateAccessKeyInput{
		UserName: aws.String(s.Username),
	}

	if !dryRun {
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

func (s *Spec) Cleanup(ctx context.Context, dryRun bool) error {
	client, err := s.buildClient(ctx)
	if err != nil {
		return err
	}
	for _, k := range s.DeletableKeys {
		input := &iam.DeleteAccessKeyInput{
			AccessKeyId: k.AccessKeyId,
			UserName:    k.UserName,
		}
		if !dryRun {
			_, err := DeleteAccessKey(ctx, client, input)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
