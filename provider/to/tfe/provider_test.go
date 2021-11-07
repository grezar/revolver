package tfe

import (
	"context"
	"io"
	"testing"

	"github.com/golang/mock/gomock"
	tfe "github.com/hashicorp/go-tfe"
	log "github.com/sirupsen/logrus"
)

func defaultWorkspaces(t *testing.T, ctrl *gomock.Controller, organization string, workspace string, workspaceID string) *tfe.MockWorkspaces {
	t.Helper()

	ctx := context.Background()
	mock := tfe.NewMockWorkspaces(ctrl)
	mock.EXPECT().
		List(ctx, organization, tfe.WorkspaceListOptions{
			Search: tfe.String(workspace),
		}).
		Return(&tfe.WorkspaceList{
			Items: []*tfe.Workspace{
				{
					ID: workspaceID,
				},
			},
		}, nil)
	return mock
}

func TestSpec_Do(t *testing.T) {
	type fields struct {
		Organization string
		Workspace    string
		Secrets      []Secret
		Variables    func(t *testing.T, ctrl *gomock.Controller, workspaceID string) *tfe.MockVariables
		Workspaces   func(t *testing.T, ctrl *gomock.Controller, organization string, workspace string, workspaceID string) *tfe.MockWorkspaces
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "Create new env secrets",
			fields: fields{
				Organization: "org1",
				Workspace:    "ws1",
				Secrets: []Secret{
					{
						Name:     "SECRET1",
						Value:    "111",
						Category: "env",
					},
					{
						Name:     "SECRET2",
						Value:    "222",
						Category: "env",
					},
				},
				Variables: func(t *testing.T, ctrl *gomock.Controller, workspaceID string) *tfe.MockVariables {
					t.Helper()

					ctx := context.Background()
					mock := tfe.NewMockVariables(ctrl)
					mock.EXPECT().List(ctx, workspaceID, tfe.VariableListOptions{}).Return(&tfe.VariableList{
						Pagination: &tfe.Pagination{
							NextPage:    0,
							CurrentPage: 0,
						},
						Items: []*tfe.Variable{},
					}, nil)
					mock.EXPECT().Create(ctx, workspaceID, tfe.VariableCreateOptions{
						Key:       tfe.String("SECRET1"),
						Value:     tfe.String("111"),
						Category:  tfe.Category(categoryEnv),
						Sensitive: tfe.Bool(true),
					}).
						Return(&tfe.Variable{}, nil)
					mock.EXPECT().Create(ctx, workspaceID, tfe.VariableCreateOptions{
						Key:       tfe.String("SECRET2"),
						Value:     tfe.String("222"),
						Category:  tfe.Category(categoryEnv),
						Sensitive: tfe.Bool(true),
					}).
						Return(&tfe.Variable{}, nil)
					return mock
				},
				Workspaces: defaultWorkspaces,
			},
		},
		{
			name: "Create new terraform secrets",
			fields: fields{
				Organization: "org1",
				Workspace:    "ws1",
				Secrets: []Secret{
					{
						Name:     "SECRET1",
						Value:    "111",
						Category: "terraform",
					},
					{
						Name:     "SECRET2",
						Value:    "222",
						Category: "terraform",
					},
				},
				Variables: func(t *testing.T, ctrl *gomock.Controller, workspaceID string) *tfe.MockVariables {
					t.Helper()

					ctx := context.Background()
					mock := tfe.NewMockVariables(ctrl)
					mock.EXPECT().List(ctx, workspaceID, tfe.VariableListOptions{}).Return(&tfe.VariableList{
						Pagination: &tfe.Pagination{
							NextPage:    0,
							CurrentPage: 0,
						},
						Items: []*tfe.Variable{},
					}, nil)
					mock.EXPECT().Create(ctx, workspaceID, tfe.VariableCreateOptions{
						Key:       tfe.String("SECRET1"),
						Value:     tfe.String("111"),
						Category:  tfe.Category(categoryTerraform),
						Sensitive: tfe.Bool(true),
					}).
						Return(&tfe.Variable{}, nil)
					mock.EXPECT().Create(ctx, workspaceID, tfe.VariableCreateOptions{
						Key:       tfe.String("SECRET2"),
						Value:     tfe.String("222"),
						Category:  tfe.Category(categoryTerraform),
						Sensitive: tfe.Bool(true),
					}).
						Return(&tfe.Variable{}, nil)
					return mock
				},
				Workspaces: defaultWorkspaces,
			},
		},
		{
			name: "Try to create an unknown category secret and fail",
			fields: fields{
				Organization: "org1",
				Workspace:    "ws1",
				Secrets: []Secret{
					{
						Name:     "SECRET1",
						Value:    "111",
						Category: "unknown",
					},
				},
				Variables: func(t *testing.T, ctrl *gomock.Controller, workspaceID string) *tfe.MockVariables {
					t.Helper()

					ctx := context.Background()
					mock := tfe.NewMockVariables(ctrl)
					mock.EXPECT().List(ctx, workspaceID, tfe.VariableListOptions{}).Return(&tfe.VariableList{
						Pagination: &tfe.Pagination{
							NextPage:    0,
							CurrentPage: 0,
						},
						Items: []*tfe.Variable{},
					}, nil)
					return mock
				},
				Workspaces: defaultWorkspaces,
			},
			wantErr: true,
		},
		{
			name: "Update an env secret",
			fields: fields{
				Organization: "org1",
				Workspace:    "ws1",
				Secrets: []Secret{
					{
						Name:     "SECRET1",
						Value:    "NEWVALUE",
						Category: "env",
					},
				},
				Variables: func(t *testing.T, ctrl *gomock.Controller, workspaceID string) *tfe.MockVariables {
					t.Helper()

					ctx := context.Background()
					variableID := "var-1"
					mock := tfe.NewMockVariables(ctrl)
					mock.EXPECT().List(ctx, workspaceID, tfe.VariableListOptions{}).Return(&tfe.VariableList{
						Pagination: &tfe.Pagination{
							NextPage:    1,
							CurrentPage: 1,
						},
						Items: []*tfe.Variable{
							{
								ID:       variableID,
								Key:      "SECRET1",
								Value:    "OLDVALUE",
								Category: categoryEnv,
							},
						},
					}, nil)
					mock.EXPECT().Update(ctx, workspaceID, variableID, tfe.VariableUpdateOptions{
						Key:       tfe.String("SECRET1"),
						Value:     tfe.String("NEWVALUE"),
						Sensitive: tfe.Bool(true),
					}).
						Return(&tfe.Variable{}, nil)
					return mock
				},
				Workspaces: defaultWorkspaces,
			},
		},
		{
			name: "Update an terraform secret",
			fields: fields{
				Organization: "org1",
				Workspace:    "ws1",
				Secrets: []Secret{
					{
						Name:     "SECRET1",
						Value:    "NEWVALUE",
						Category: "terraform",
					},
				},
				Variables: func(t *testing.T, ctrl *gomock.Controller, workspaceID string) *tfe.MockVariables {
					t.Helper()

					ctx := context.Background()
					variableID := "var-1"
					mock := tfe.NewMockVariables(ctrl)
					mock.EXPECT().List(ctx, workspaceID, tfe.VariableListOptions{}).Return(&tfe.VariableList{
						Pagination: &tfe.Pagination{
							NextPage:    1,
							CurrentPage: 1,
						},
						Items: []*tfe.Variable{
							{
								ID:       variableID,
								Key:      "SECRET1",
								Value:    "OLDVALUE",
								Category: categoryTerraform,
							},
						},
					}, nil)
					mock.EXPECT().Update(ctx, workspaceID, variableID, tfe.VariableUpdateOptions{
						Key:       tfe.String("SECRET1"),
						Value:     tfe.String("NEWVALUE"),
						Sensitive: tfe.Bool(true),
					}).
						Return(&tfe.Variable{}, nil)
					return mock
				},
				Workspaces: defaultWorkspaces,
			},
		},
		{
			name: "Ambiguious workspace name provided and fail",
			fields: fields{
				Organization: "org1",
				Workspace:    "ws",
				Secrets: []Secret{
					{
						Name:     "SECRET1",
						Value:    "NEWVALUE",
						Category: "terraform",
					},
				},
				Variables: func(t *testing.T, ctrl *gomock.Controller, workspaceID string) *tfe.MockVariables {
					t.Helper()

					mock := tfe.NewMockVariables(ctrl)
					return mock
				},
				Workspaces: func(t *testing.T, ctrl *gomock.Controller, organization string, workspace string, workspaceID string) *tfe.MockWorkspaces {
					t.Helper()

					ctx := context.Background()
					mock := tfe.NewMockWorkspaces(ctrl)
					mock.EXPECT().
						List(ctx, organization, tfe.WorkspaceListOptions{
							Search: tfe.String(workspace),
						}).
						Return(&tfe.WorkspaceList{
							Items: []*tfe.Workspace{
								{
									ID: "ws-1",
								},
								{
									ID: "ws-2",
								},
							},
						}, nil)
					return mock
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		nullLogger := log.New()
		nullLogger.SetOutput(io.Discard)

		t.Run(tt.name, func(t *testing.T) {
			mockTfeAPI := &tfe.Client{}
			ctrl := gomock.NewController(t)
			workspaceID := "ws-1"
			mockTfeAPI.Variables = tt.fields.Variables(t, ctrl, workspaceID)
			mockTfeAPI.Workspaces = tt.fields.Workspaces(t, ctrl, tt.fields.Organization, tt.fields.Workspace, workspaceID)

			s := &Spec{
				Organization: tt.fields.Organization,
				Workspace:    tt.fields.Workspace,
				Secrets:      tt.fields.Secrets,
				Client:       mockTfeAPI,
				Logger:       nullLogger,
			}

			ctx := context.Background()

			if err := s.Do(ctx); (err != nil) != tt.wantErr {
				t.Errorf("Spec.Do() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
