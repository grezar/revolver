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
	"github.com/grezar/revolver/secrets"
)

func TestSpec_Do(t *testing.T) {
	type fields struct {
		AccountID                 string
		Username                  string
		Expiration                string
		ForceDeleteAllExpiredKeys bool
		MockIAMAccessKeyAPI       mock.MockIAMAccessKeyAPI
		dryRun                    bool
	}
	tests := []struct {
		name    string
		fields  fields
		want    secrets.Secrets
		wantErr bool
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
			want: secrets.Secrets{
				"AWSAccessKeyID":     "BBBBBBBBBBBB",
				"AWSSecretAccessKey": "CCCCCCCCCCCC",
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
			want: nil,
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
					DeleteAccessKeyAPI: mock.NewMockDeleteAccessKeyAPI(),
				},
			},
			want: secrets.Secrets{
				"AWSAccessKeyID":     "BBBBBBBBBBBB",
				"AWSSecretAccessKey": "CCCCCCCCCCCC",
			},
		},
		{
			name: "Do nothing, but require to delete the key",
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
			wantErr: true,
		},
		{
			name: "It doesn't do destructive changes in dry-run mode",
			fields: fields{
				AccountID:  "0123456789",
				Username:   "test-iam-user",
				Expiration: "15m",
				MockIAMAccessKeyAPI: mock.MockIAMAccessKeyAPI{
					ListAccessKeysAPI: mock.NewMockListAccessKeysAPI(),
				},
				dryRun: true,
			},
			want: nil,
		},
		{
			name: "Delete both of expired keys if they're expired with forceDeleteAllExpiredKeys enabled",
			fields: fields{
				AccountID:                 "0123456789",
				Username:                  "test-iam-user",
				Expiration:                "90d",
				ForceDeleteAllExpiredKeys: true,
				MockIAMAccessKeyAPI: mock.MockIAMAccessKeyAPI{
					ListAccessKeysAPI: mock.MockListAccessKeys(
						func(ctx context.Context, params *iam.ListAccessKeysInput, optFns ...func(*iam.Options)) (*iam.ListAccessKeysOutput, error) {
							return &iam.ListAccessKeysOutput{
								AccessKeyMetadata: []types.AccessKeyMetadata{
									{
										AccessKeyId: aws.String("EXPIRED1"),
										CreateDate:  aws.Time(time.Now().Add(-24 * 200 * time.Hour)),
										UserName:    aws.String("test-iam-user"),
									},
									{
										AccessKeyId: aws.String("EXPIRED2"),
										CreateDate:  aws.Time(time.Now().Add(-24 * 180 * time.Hour)),
										UserName:    aws.String("test-iam-user"),
									},
								},
							}, nil
						},
					),
					CreateAccessKeyAPI: mock.NewMockCreateAccessKeyAPI(),
					DeleteAccessKeyAPI: mock.NewMockDeleteAccessKeyAPI(),
				},
				dryRun: false,
			},
			want: secrets.Secrets{
				"AWSAccessKeyID":     "BBBBBBBBBBBB",
				"AWSSecretAccessKey": "CCCCCCCCCCCC",
			},
		},
		{
			name: "DO NOT delete any keys if all of them aren't expired with forceDeleteAllExpiredKeys enabled",
			fields: fields{
				AccountID:                 "0123456789",
				Username:                  "test-iam-user",
				Expiration:                "90d",
				ForceDeleteAllExpiredKeys: true,
				MockIAMAccessKeyAPI: mock.MockIAMAccessKeyAPI{
					ListAccessKeysAPI: mock.MockListAccessKeys(
						func(ctx context.Context, params *iam.ListAccessKeysInput, optFns ...func(*iam.Options)) (*iam.ListAccessKeysOutput, error) {
							return &iam.ListAccessKeysOutput{
								AccessKeyMetadata: []types.AccessKeyMetadata{
									{
										AccessKeyId: aws.String("NOTEXPIRED1"),
										CreateDate:  aws.Time(time.Now().Add(-24 * 1 * time.Hour)),
										UserName:    aws.String("test-iam-user"),
									},
									{
										AccessKeyId: aws.String("NOTEXPIRED2"),
										CreateDate:  aws.Time(time.Now().Add(-24 * 1 * time.Hour)),
										UserName:    aws.String("test-iam-user"),
									},
								},
							}, nil
						},
					),
				},
				dryRun: false,
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Spec{
				AccountID:                 tt.fields.AccountID,
				Username:                  tt.fields.Username,
				Expiration:                tt.fields.Expiration,
				ForceDeleteAllExpiredKeys: tt.fields.ForceDeleteAllExpiredKeys,
				Client:                    tt.fields.MockIAMAccessKeyAPI,
			}
			ctx := context.Background()
			got, err := s.Do(ctx, tt.fields.dryRun)
			if (err != nil) != tt.wantErr {
				t.Errorf("Spec.Do() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(got, tt.want) {
				fmt.Println(got)
				t.Errorf("Spec.Do() = %v, want %v", got, tt.want)
			}
		})
	}
}
