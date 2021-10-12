package awsiamuser

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/grezar/revolver/provider/from/awsiamuser/mock"
	"github.com/grezar/revolver/repository"
)

func TestSpec_RenewKey(t *testing.T) {
	type fields struct {
		AccountID           string
		Username            string
		Expiration          string
		MockIAMAccessKeyAPI mock.MockIAMAccessKeyAPI
	}
	tests := []struct {
		name   string
		fields fields
		want   *repository.Repository
	}{
		{
			name: "Create a new key due to no key was found",
			fields: fields{
				AccountID:  "0123456789",
				Username:   "test-iam-user",
				Expiration: "1d",
				MockIAMAccessKeyAPI: mock.MockIAMAccessKeyAPI{
					ListAccessKeysAPI: mock.MockListAccessKeys(
						func(ctx context.Context, params *iam.ListAccessKeysInput, optFns ...func(*iam.Options)) (*iam.ListAccessKeysOutput, error) {
							return &iam.ListAccessKeysOutput{
								AccessKeyMetadata: []types.AccessKeyMetadata{},
							}, nil
						},
					),
					CreateAccessKeyAPI: mock.NewMockCreateAccessKeyAPI(),
				},
			},
			want: &repository.Repository{
				Secrets: map[string]string{
					"AWSAccessKeyID":     "BBBBBBBBBBBB",
					"AWSSecretAccessKey": "CCCCCCCCCCCC",
				},
			},
		},
		{
			name: "Do nothing, there's an available key and no expired key",
			fields: fields{
				AccountID:  "0123456789",
				Username:   "test-iam-user",
				Expiration: "90d",
				MockIAMAccessKeyAPI: mock.MockIAMAccessKeyAPI{
					ListAccessKeysAPI: mock.NewMockListAccessKeysAPI(),
				},
			},
			want: &repository.Repository{},
		},
		{
			name: "Renew a key, there's an expired key",
			fields: fields{
				AccountID:  "0123456789",
				Username:   "test-iam-user",
				Expiration: "15m",
				MockIAMAccessKeyAPI: mock.MockIAMAccessKeyAPI{
					ListAccessKeysAPI:  mock.NewMockListAccessKeysAPI(),
					CreateAccessKeyAPI: mock.NewMockCreateAccessKeyAPI(),
				},
			},
			want: &repository.Repository{
				Secrets: map[string]string{
					"AWSAccessKeyID":     "BBBBBBBBBBBB",
					"AWSSecretAccessKey": "CCCCCCCCCCCC",
				},
			},
		},
		{
			name: "Do nothing but warn to require manual operation due to there are two keys",
			fields: fields{
				AccountID:  "0123456789",
				Username:   "test-iam-user",
				Expiration: "90d",
				MockIAMAccessKeyAPI: mock.MockIAMAccessKeyAPI{
					ListAccessKeysAPI: mock.MockListAccessKeys(
						func(ctx context.Context, params *iam.ListAccessKeysInput, optFns ...func(*iam.Options)) (*iam.ListAccessKeysOutput, error) {
							return &iam.ListAccessKeysOutput{
								AccessKeyMetadata: []types.AccessKeyMetadata{
									{
										AccessKeyId: aws.String("NOTEXPIRED1"),
										CreateDate:  aws.Time(time.Now().Add(-24 * time.Hour)),
										UserName:    aws.String("test-iam-user"),
									},
									{
										AccessKeyId: aws.String("EXPIRED1"),
										CreateDate:  aws.Time(time.Now().Add(-24 * 180 * time.Hour)),
										UserName:    aws.String("test-iam-user"),
									},
								},
							}, nil
						},
					),
				},
			},
			want: &repository.Repository{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Spec{
				AccountID:  tt.fields.AccountID,
				Username:   tt.fields.Username,
				Expiration: tt.fields.Expiration,
				Client:     tt.fields.MockIAMAccessKeyAPI,
			}
			got, _ := s.RenewKey()
			if !reflect.DeepEqual(got, tt.want) {
				fmt.Println(got)
				t.Errorf("Spec.RenewKey() = %v, want %v", got, tt.want)
			}
		})
	}
}
