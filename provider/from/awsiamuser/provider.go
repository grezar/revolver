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
	log "github.com/sirupsen/logrus"
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
	s.Logger = log.WithFields(log.Fields{
		"provider": name,
	})
	return &s, nil
}

// fromprovider.Operator
type Spec struct {
	AccountID     string `yaml:"accountId" validate:"required"`
	Username      string `yaml:"username" validate:"required"`
	Expiration    string `yaml:"expiration"`
	DeletableKeys []types.AccessKey
	Client        IAMAccessKeyAPI
	Logger        log.FieldLogger
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

func (s *Spec) Do(ctx context.Context) (secrets.Secrets, error) {
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
		s.Logger.Info("Create a new access key since no key is found.")
	case 1:
		if expiration <= time.Since(aws.ToTime(keys.AccessKeyMetadata[0].CreateDate)) {
			s.Logger.Info("Create a new access key since the existing key is expired")

			s.DeletableKeys = append(s.DeletableKeys, types.AccessKey{
				AccessKeyId: keys.AccessKeyMetadata[0].AccessKeyId,
				UserName:    keys.AccessKeyMetadata[0].UserName,
			})
		} else {
			s.Logger.Info("The key hasn't expired yet. Do nothing.")
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
	output, err := CreateAccessKey(ctx, client, input)
	if err != nil {
		return nil, err
	}
	s.Logger.Info("Successfully created a new key.")

	return secrets.Secrets{
		keyAWSAccessKeyID:     aws.ToString(output.AccessKey.AccessKeyId),
		keyAWSSecretAccessKey: aws.ToString(output.AccessKey.SecretAccessKey),
	}, nil
}

func (s *Spec) Cleanup(ctx context.Context) error {
	client, err := s.buildClient(ctx)
	if err != nil {
		return err
	}
	for _, k := range s.DeletableKeys {
		s.Logger.Info("Delete the existing key to rotate.")

		input := &iam.DeleteAccessKeyInput{
			AccessKeyId: k.AccessKeyId,
			UserName:    k.UserName,
		}
		_, err := DeleteAccessKey(ctx, client, input)
		if err != nil {
			return err
		}

		s.Logger.Info("Successfully deleted the existing key.")
	}
	return nil
}
