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
	AccountID  string `yaml:"accountId" validate:"required"`
	Username   string `yaml:"username" validate:"required"`
	Expiration string `yaml:"expiration"`
	Client     IAMAccessKeyAPI
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

func (s *Spec) Do(ctx context.Context, dryRun bool) (_ secrets.Secrets, doErr error) {
	client, doErr := s.buildClient(ctx)
	if doErr != nil {
		return nil, doErr
	}

	inpt := &iam.ListAccessKeysInput{
		UserName: aws.String(s.Username),
	}
	keys, doErr := ListAccessKeys(ctx, client, inpt)
	if doErr != nil {
		return nil, doErr
	}

	expiration, doErr := str2duration.ParseDuration(s.Expiration)
	if doErr != nil {
		return nil, doErr
	}

	switch len(keys.AccessKeyMetadata) {
	case 0:
		// Only to proceed to the next step.
	case 1:
		if expiration <= time.Since(aws.ToTime(keys.AccessKeyMetadata[0].CreateDate)) {

			defer func() {
				doErr = s.cleanup(ctx, dryRun, types.AccessKey{
					AccessKeyId: keys.AccessKeyMetadata[0].AccessKeyId,
					UserName:    keys.AccessKeyMetadata[0].UserName,
				})
			}()
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
		output, doErr := CreateAccessKey(ctx, client, input)
		if doErr != nil {
			return nil, doErr
		}

		return secrets.Secrets{
			keyAWSAccessKeyID:     aws.ToString(output.AccessKey.AccessKeyId),
			keyAWSSecretAccessKey: aws.ToString(output.AccessKey.SecretAccessKey),
		}, nil
	}

	return nil, nil
}

func (s *Spec) cleanup(ctx context.Context, dryRun bool, deletableKey types.AccessKey) error {
	client, err := s.buildClient(ctx)
	if err != nil {
		return err
	}
	input := &iam.DeleteAccessKeyInput{
		AccessKeyId: deletableKey.AccessKeyId,
		UserName:    deletableKey.UserName,
	}
	if !dryRun {
		_, err := DeleteAccessKey(ctx, client, input)
		if err != nil {
			return err
		}
	}
	return nil
}
