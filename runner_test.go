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
		mockedRotations func(t *testing.T, ctrl *gomock.Controller, dryRun bool) []*schema.Rotation
		dryRun          bool
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "Secrets are expired, do rotation",
			fields: fields{
				mockedRotations: func(t *testing.T, ctrl *gomock.Controller, dryRun bool) []*schema.Rotation {
					t.Helper()

					ctx := context.Background()

					mockedFromOperator := mockedfp.NewMockOperator(ctrl)
					mockedToOperator := mockedtp.NewMockOperator(ctrl)

					// advance dry-run
					mockedFromOperator.EXPECT().Summary().Return("mocked from operator")
					mockedFromOperator.EXPECT().Do(ctx, true).Return(nil, nil)
					mockedToOperator.EXPECT().Summary().Return("mocked to operator")
					mockedToOperator.EXPECT().Do(ctx, true)

					// actual run
					mockedFromOperator.EXPECT().Summary().Return("mocked from operator")
					expectedSecrets := secrets.Secrets{
						"KEY_ID": "key1",
						"SECRET": "secret1",
					}
					mockedFromOperator.EXPECT().Do(ctx, dryRun).Return(expectedSecrets, nil)
					ctx = secrets.WithSecrets(ctx, expectedSecrets)
					mockedToOperator.EXPECT().Summary().Return("mocked to operator")
					mockedToOperator.EXPECT().Do(ctx, dryRun)

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
				mockedRotations: func(t *testing.T, ctrl *gomock.Controller, dryRun bool) []*schema.Rotation {
					t.Helper()

					ctx := context.Background()
					expectedSecrets := secrets.Secrets{}

					mockedFromOperator := mockedfp.NewMockOperator(ctrl)

					// advance dry-run
					mockedFromOperator.EXPECT().Summary().Return("mocked from operator")
					mockedFromOperator.EXPECT().Do(ctx, true).Return(nil, nil)

					// actual run
					mockedFromOperator.EXPECT().Summary().Return("mocked from operator")
					mockedFromOperator.EXPECT().Do(ctx, dryRun).Return(expectedSecrets, nil)

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
				mockedRotations: func(t *testing.T, ctrl *gomock.Controller, dryRun bool) []*schema.Rotation {
					t.Helper()

					ctx := context.Background()

					mockedFromOperator := mockedfp.NewMockOperator(ctrl)

					// advance dry-run
					mockedFromOperator.EXPECT().Summary().Return("mocked from operator")
					mockedFromOperator.EXPECT().Do(ctx, true).Return(nil, errFakeRunnerTest)

					// actual run
					mockedFromOperator.EXPECT().Summary().Return("mocked from operator")
					mockedFromOperator.EXPECT().Do(ctx, dryRun).Return(nil, errFakeRunnerTest)

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
			wantErr: true,
		},
		{
			name: "One or more to provider returns an error and the cleanup is invoked",
			fields: fields{
				mockedRotations: func(t *testing.T, ctrl *gomock.Controller, dryRun bool) []*schema.Rotation {
					t.Helper()

					ctx := context.Background()
					expectedSecrets := secrets.Secrets{
						"KEY_ID": "key1",
					}

					mockedFromOperator := mockedfp.NewMockOperator(ctrl)
					mockedToOperator1 := mockedtp.NewMockOperator(ctrl)
					mockedToOperator2 := mockedtp.NewMockOperator(ctrl)

					// advance dry-run
					mockedFromOperator.EXPECT().Summary().Return("mocked from operator")
					mockedFromOperator.EXPECT().Do(ctx, true).Return(nil, nil)
					mockedToOperator1.EXPECT().Summary().Return("mocked to operator 1")
					mockedToOperator1.EXPECT().Do(ctx, true).Return(errFakeRunnerTest)
					mockedToOperator2.EXPECT().Summary().Return("mocked to operator 2")
					mockedToOperator2.EXPECT().Do(ctx, true).Return(nil)

					// actual run
					mockedFromOperator.EXPECT().Summary().Return("mocked from operator")
					mockedFromOperator.EXPECT().Do(ctx, dryRun).Return(expectedSecrets, nil)
					ctx = secrets.WithSecrets(ctx, expectedSecrets)
					mockedToOperator1.EXPECT().Summary().Return("mocked to operator 1")
					mockedToOperator1.EXPECT().Do(ctx, dryRun).Return(errFakeRunnerTest)
					mockedToOperator2.EXPECT().Summary().Return("mocked to operator 2")
					mockedToOperator2.EXPECT().Do(ctx, dryRun).Return(nil)

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
										Operator: mockedToOperator1,
									},
								},
								{
									Spec: schema.ToProviderSpec{
										Operator: mockedToOperator2,
									},
								},
							},
						},
					}

					return rotations
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			r := &Runner{
				rotations: tt.fields.mockedRotations(t, ctrl, tt.fields.dryRun),
			}

			ok := reporting.Run(func(rptr *reporting.R) {
				r.Run(rptr)
			})
			if !ok != tt.wantErr {
				t.Fail()
			}
		})
	}
}
