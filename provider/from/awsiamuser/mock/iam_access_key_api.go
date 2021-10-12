package mock

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/iam"
)

type MockIAMAccessKeyParams struct {
	ListAccessKeysOutput  *iam.ListAccessKeysOutput
	CreateAccessKeyOutput *iam.CreateAccessKeyOutput
	DeleteAccessKeyOutput *iam.DeleteAccessKeyOutput
}

// MockACMAPI is a struct that represents an ACM client.
type MockIAMAccessKeyAPI struct {
	ListAccessKeysAPI  MockListAccessKeys
	CreateAccessKeyAPI MockCreateAccessKey
	DeleteAccessKeyAPI MockDeleteAccessKey
}

// MockListAccessKeys is a type that represents a function that mock IAM's ListAccessKeys.
type MockListAccessKeys func(ctx context.Context, params *iam.ListAccessKeysInput, optFns ...func(*iam.Options)) (*iam.ListAccessKeysOutput, error)

// MockCreateAccessKey is a type that represents a function that mock IAM's ListAccessKeys.
type MockCreateAccessKey func(ctx context.Context, params *iam.CreateAccessKeyInput, optFns ...func(*iam.Options)) (*iam.CreateAccessKeyOutput, error)

// MockDeleteAccessKey is a type that represents a function that mock IAM's ListAccessKeys.
type MockDeleteAccessKey func(ctx context.Context, params *iam.DeleteAccessKeyInput, optFns ...func(*iam.Options)) (*iam.DeleteAccessKeyOutput, error)

// ListAccessKeys returns a function that mock original of IAM ListAccessKeys.
func (m MockIAMAccessKeyAPI) ListAccessKeys(ctx context.Context, params *iam.ListAccessKeysInput, optFns ...func(*iam.Options)) (*iam.ListAccessKeysOutput, error) {
	return m.ListAccessKeysAPI(ctx, params, optFns...)
}

// CreateAccessKeyAPI returns a function that mock original of ACM DescribeCertificate.
func (m MockIAMAccessKeyAPI) CreateAccessKey(ctx context.Context, params *iam.CreateAccessKeyInput, optFns ...func(*iam.Options)) (*iam.CreateAccessKeyOutput, error) {
	return m.CreateAccessKeyAPI(ctx, params, optFns...)
}

// CreateAccessKeyAPI returns a function that mock original of ACM DescribeCertificate.
func (m MockIAMAccessKeyAPI) DeleteAccessKey(ctx context.Context, params *iam.DeleteAccessKeyInput, optFns ...func(*iam.Options)) (*iam.DeleteAccessKeyOutput, error) {
	return m.DeleteAccessKeyAPI(ctx, params, optFns...)
}
