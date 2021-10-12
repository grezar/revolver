package mock

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
)

func NewMockIAMAccessKeyAPI(mockParams MockIAMAccessKeyParams) MockIAMAccessKeyAPI {
	return MockIAMAccessKeyAPI{
		ListAccessKeysAPI:  NewMockListAccessKeysAPI(),
		CreateAccessKeyAPI: NewMockCreateAccessKeyAPI(),
		DeleteAccessKeyAPI: NewMockDeleteAccessKeyAPI(),
	}
}

func NewMockListAccessKeysAPI() MockListAccessKeys {
	return MockListAccessKeys(func(ctx context.Context, params *iam.ListAccessKeysInput, optFns ...func(*iam.Options)) (*iam.ListAccessKeysOutput, error) {
		return &iam.ListAccessKeysOutput{
			AccessKeyMetadata: []types.AccessKeyMetadata{
				{
					AccessKeyId: aws.String("AAAAAAAAAAAA"),
					CreateDate:  aws.Time(time.Now().Add(-24 * time.Hour)),
					UserName:    aws.String("test-iam-user"),
				},
			},
		}, nil
	})
}

func NewMockCreateAccessKeyAPI() MockCreateAccessKey {
	return MockCreateAccessKey(func(ctx context.Context, params *iam.CreateAccessKeyInput, optFns ...func(*iam.Options)) (*iam.CreateAccessKeyOutput, error) {
		return &iam.CreateAccessKeyOutput{
			AccessKey: &types.AccessKey{
				AccessKeyId:     aws.String("BBBBBBBBBBBB"),
				SecretAccessKey: aws.String("CCCCCCCCCCCC"),
			},
		}, nil
	})
}

func NewMockDeleteAccessKeyAPI() MockDeleteAccessKey {
	return MockDeleteAccessKey(func(ctx context.Context, params *iam.DeleteAccessKeyInput, optFns ...func(*iam.Options)) (*iam.DeleteAccessKeyOutput, error) {
		return &iam.DeleteAccessKeyOutput{}, nil
	})
}
