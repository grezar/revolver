package revolver

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	mockedfp "github.com/grezar/revolver/provider/from/mocks"
	mockedtp "github.com/grezar/revolver/provider/to/mocks"
	"github.com/grezar/revolver/reporting"
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
					mockedFromOperator.EXPECT().Summary().Return("mocked from operator")
					mockedFromOperator.EXPECT().Do(ctx).Return(expectedSecrets, nil)
					ctx = secrets.WithSecrets(ctx, expectedSecrets)
					mockedToOperator.EXPECT().Summary().Return("mocked to operator")
					mockedToOperator.EXPECT().Do(ctx)
					mockedFromOperator.EXPECT().Cleanup(ctx)

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
					mockedFromOperator.EXPECT().Summary().Return("mocked from operator")
					mockedFromOperator.EXPECT().Do(ctx).Return(expectedSecrets, nil)

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
					mockedFromOperator.EXPECT().Summary().Return("mocked from operator")
					mockedFromOperator.EXPECT().Do(ctx).Return(nil, errFakeRunnerTest)

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
			name: "One or more to provider returns an error and the cleanup is invoked",
			fields: fields{
				mockedRotations: func(t *testing.T, ctrl *gomock.Controller) []*schema.Rotation {
					t.Helper()

					ctx := context.Background()
					expectedSecrets := secrets.Secrets{
						"KEY_ID": "key1",
					}

					mockedFromOperator := mockedfp.NewMockOperator(ctrl)
					mockedToOperator := mockedtp.NewMockOperator(ctrl)
					mockedFromOperator.EXPECT().Summary().Return("mocked from operator")
					mockedFromOperator.EXPECT().Do(ctx).Return(expectedSecrets, nil)
					ctx = secrets.WithSecrets(ctx, expectedSecrets)
					mockedToOperator.EXPECT().Summary().Return("mocked to operator")
					mockedToOperator.EXPECT().Do(ctx).Return(errFakeRunnerTest)
					mockedToOperator.EXPECT().Summary().Return("mocked to operator")
					mockedToOperator.EXPECT().Do(ctx).Return(nil)
					mockedFromOperator.EXPECT().Cleanup(ctx)

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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			r := &Runner{
				rotations: tt.fields.mockedRotations(t, ctrl),
			}

			reporting.Run(func (rptr *reporting.R) {
				r.Run(rptr)
			})
		})
	}
}
