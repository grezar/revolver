package revolver

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	mockedfp "github.com/grezar/revolver/provider/from/mocks"
	mockedtp "github.com/grezar/revolver/provider/to/mocks"
	"github.com/grezar/revolver/schema"
	"github.com/grezar/revolver/secrets"
)

var (
	errFakeRunnerTest = errors.New("runner test fake error")
)

func TestRunner_Run(t *testing.T) {
	type fields struct {
		mockedRotations func(t *testing.T, ctrl *gomock.Controller) []*schema.Rotation
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "Secrets are expired, do rotation",
			fields: fields{
				mockedRotations: func(t *testing.T, ctrl *gomock.Controller) []*schema.Rotation {
					t.Helper()

					ctx := context.Background()
					expectedSecrets := secrets.Secrets{
						"KEY_ID": "key1",
						"SECRET": "secret1",
					}

					mockedFromOperator := mockedfp.NewMockOperator(ctrl)
					mockedToOperator := mockedtp.NewMockOperator(ctrl)
					mockedFromOperator.EXPECT().RenewKey(ctx).Return(expectedSecrets, nil)
					ctx = secrets.WithSecrets(ctx, expectedSecrets)
					mockedToOperator.EXPECT().UpdateSecret(ctx)
					mockedFromOperator.EXPECT().DeleteKey(ctx)

					rotations := []*schema.Rotation{
						{
							Name: "Mocked Rotation",
							From: schema.From{
								Spec: schema.FromProviderSpec{
									Operator: mockedFromOperator,
								},
							},
							To: []*schema.To{
								{
									Spec: schema.ToProviderSpec{
										Operator: mockedToOperator,
									},
								},
							},
						},
					}

					return rotations
				},
			},
		},
		{
			name: "Secrets aren't expired",
			fields: fields{
				mockedRotations: func(t *testing.T, ctrl *gomock.Controller) []*schema.Rotation {
					t.Helper()

					ctx := context.Background()
					expectedSecrets := secrets.Secrets{}

					mockedFromOperator := mockedfp.NewMockOperator(ctrl)
					mockedFromOperator.EXPECT().RenewKey(ctx).Return(expectedSecrets, nil)

					rotations := []*schema.Rotation{
						{
							Name: "Mocked Rotation",
							From: schema.From{
								Spec: schema.FromProviderSpec{
									Operator: mockedFromOperator,
								},
							},
						},
					}

					return rotations
				},
			},
		},
		{
			name: "The from provider returns error and skip following operations",
			fields: fields{
				mockedRotations: func(t *testing.T, ctrl *gomock.Controller) []*schema.Rotation {
					t.Helper()

					ctx := context.Background()

					mockedFromOperator := mockedfp.NewMockOperator(ctrl)
					mockedFromOperator.EXPECT().RenewKey(ctx).Return(nil, errFakeRunnerTest)

					rotations := []*schema.Rotation{
						{
							Name: "Mocked Rotation",
							From: schema.From{
								Spec: schema.FromProviderSpec{
									Operator: mockedFromOperator,
								},
							},
						},
					}

					return rotations
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			r := &Runner{
				rotations: tt.fields.mockedRotations(t, ctrl),
			}

			if err := r.Run(); (err != nil) != tt.wantErr {
				t.Errorf("Runner.Run() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
