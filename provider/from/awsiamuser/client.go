package awsiamuser

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/iam"
)

type IAMAccessKeyAPI interface {
	ListAccessKeys(ctx context.Context,
		params *iam.ListAccessKeysInput,
		optFns ...func(*iam.Options)) (*iam.ListAccessKeysOutput, error)
	CreateAccessKey(ctx context.Context,
		params *iam.CreateAccessKeyInput,
		optFns ...func(*iam.Options)) (*iam.CreateAccessKeyOutput, error)
	DeleteAccessKey(ctx context.Context,
		params *iam.DeleteAccessKeyInput,
		optFns ...func(*iam.Options)) (*iam.DeleteAccessKeyOutput, error)
}

func ListAccessKeys(c context.Context, api IAMAccessKeyAPI, input *iam.ListAccessKeysInput) (*iam.ListAccessKeysOutput, error) {
	return api.ListAccessKeys(c, input)
}

func CreateAccessKey(ctx context.Context, api IAMAccessKeyAPI, input *iam.CreateAccessKeyInput) (*iam.CreateAccessKeyOutput, error) {
	return api.CreateAccessKey(ctx, input)
}

func DeleteAccessKey(ctx context.Context, api IAMAccessKeyAPI, input *iam.DeleteAccessKeyInput) (*iam.DeleteAccessKeyOutput, error) {
	return api.DeleteAccessKey(ctx, input)
}
