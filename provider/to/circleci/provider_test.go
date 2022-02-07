package circleci

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/grezar/go-circleci"
	mock "github.com/grezar/go-circleci/mocks"
	"go.uber.org/ratelimit"
)

func TestSpec_Summary(t *testing.T) {
	type fields struct {
		Owner            string
		ProjectVariables []*ProjectVariable
		Contexts         []*Context
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "Summary returns string with contexts and projects combined",
			fields: fields{
				Owner: "org1",
				ProjectVariables: []*ProjectVariable{
					{
						Project: "prj1",
					},
					{
						Project: "prj2",
					},
				},
				Contexts: []*Context{
					{
						Name: "ctx1",
					},
					{
						Name: "ctx2",
					},
				},
			},
			want: "owner: org1, contexts: ctx1, ctx2, projects: prj1, prj2",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Spec{
				Owner:            tt.fields.Owner,
				ProjectVariables: tt.fields.ProjectVariables,
				Contexts:         tt.fields.Contexts,
				RateLimit:        ratelimit.New(apiRateLimit),
			}
			if got := s.Summary(); got != tt.want {
				t.Errorf("Spec.Summary() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSpec_UpdateProjectVariables(t *testing.T) {
	type fields struct {
		Owner            string
		ProjectVariables []*ProjectVariable
		ContextVariables []*Context
		Projects         func(t *testing.T, ctrl *gomock.Controller) *mock.MockProjects
		Contexts         func(t *testing.T, ctrl *gomock.Controller) *mock.MockContexts
		dryRun           bool
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "Create new project variables",
			fields: fields{
				Owner: "org1",
				ProjectVariables: []*ProjectVariable{
					{
						Project: "prj1",
						Variables: []*Variable{
							{
								Name:  "SECRET1",
								Value: "111",
							},
							{
								Name:  "SECRET2",
								Value: "222",
							},
						},
					},
				},
				Projects: func(t *testing.T, ctrl *gomock.Controller) *mock.MockProjects {
					t.Helper()

					ctx := context.Background()
					project := "prj1"
					mock := mock.NewMockProjects(ctrl)
					mock.EXPECT().ListVariables(ctx, project).Return(&circleci.ProjectVariableList{
						Items: []*circleci.ProjectVariable{},
					}, nil)
					mock.EXPECT().CreateVariable(ctx, project, circleci.ProjectCreateVariableOptions{
						Name:  circleci.String("SECRET1"),
						Value: circleci.String("111"),
					})
					mock.EXPECT().CreateVariable(ctx, project, circleci.ProjectCreateVariableOptions{
						Name:  circleci.String("SECRET2"),
						Value: circleci.String("222"),
					})
					return mock
				},
			},
		},
		{
			name: "Met conditions for creating new project variables but doesn't do destructive changes in dry-run mode",
			fields: fields{
				Owner: "org1",
				ProjectVariables: []*ProjectVariable{
					{
						Project: "prj1",
						Variables: []*Variable{
							{
								Name:  "SECRET1",
								Value: "111",
							},
							{
								Name:  "SECRET2",
								Value: "222",
							},
						},
					},
				},
				Projects: func(t *testing.T, ctrl *gomock.Controller) *mock.MockProjects {
					t.Helper()

					ctx := context.Background()
					project := "prj1"
					mock := mock.NewMockProjects(ctrl)
					mock.EXPECT().ListVariables(ctx, project).Return(&circleci.ProjectVariableList{
						Items: []*circleci.ProjectVariable{},
					}, nil)
					return mock
				},
				dryRun: true,
			},
		},
		{
			name: "Update an existing project variable by deleting it and creating a new one",
			fields: fields{
				Owner: "org1",
				ProjectVariables: []*ProjectVariable{
					{
						Project: "prj1",
						Variables: []*Variable{
							{
								Name:  "SECRET1",
								Value: "111",
							},
						},
					},
				},
				Projects: func(t *testing.T, ctrl *gomock.Controller) *mock.MockProjects {
					t.Helper()

					ctx := context.Background()
					project := "prj1"
					mock := mock.NewMockProjects(ctrl)
					mock.EXPECT().ListVariables(ctx, project).Return(&circleci.ProjectVariableList{
						Items: []*circleci.ProjectVariable{
							{
								Name:  "SECRET1",
								Value: "000",
							},
						},
					}, nil)
					mock.EXPECT().DeleteVariable(ctx, project, "SECRET1")
					mock.EXPECT().CreateVariable(ctx, project, circleci.ProjectCreateVariableOptions{
						Name:  circleci.String("SECRET1"),
						Value: circleci.String("111"),
					})
					return mock
				},
			},
		},
		{
			name: "It returns an error when the matching project is not found",
			fields: fields{
				Owner: "org1",
				ProjectVariables: []*ProjectVariable{
					{
						Project: "prj1",
						Variables: []*Variable{
							{
								Name:  "SECRET1",
								Value: "111",
							},
						},
					},
				},
				Projects: func(t *testing.T, ctrl *gomock.Controller) *mock.MockProjects {
					t.Helper()

					ctx := context.Background()
					project := "prj1"
					mock := mock.NewMockProjects(ctrl)
					mock.EXPECT().ListVariables(ctx, project).Return(nil, errors.New("project is not found"))
					return mock
				},
			},
			wantErr: true,
		},
		{
			name: "Met conditions for updating project variables but doesn't do destructive changes in dry-run mode",
			fields: fields{
				Owner: "org1",
				ProjectVariables: []*ProjectVariable{
					{
						Project: "prj1",
						Variables: []*Variable{
							{
								Name:  "SECRET1",
								Value: "111",
							},
						},
					},
				},
				Projects: func(t *testing.T, ctrl *gomock.Controller) *mock.MockProjects {
					t.Helper()

					ctx := context.Background()
					project := "prj1"
					mock := mock.NewMockProjects(ctrl)
					mock.EXPECT().ListVariables(ctx, project).Return(&circleci.ProjectVariableList{
						Items: []*circleci.ProjectVariable{
							{
								Name:  "SECRET1",
								Value: "000",
							},
						},
					}, nil)
					return mock
				},
				dryRun: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockCircleCIAPI := &circleci.Client{}
			mockCircleCIAPI.Projects = tt.fields.Projects(t, ctrl)
			s := &Spec{
				Owner:            tt.fields.Owner,
				ProjectVariables: tt.fields.ProjectVariables,
				RateLimit:        ratelimit.New(apiRateLimit),
			}
			if err := s.UpdateProjectVariables(context.Background(), tt.fields.dryRun, mockCircleCIAPI, s.RateLimit); (err != nil) != tt.wantErr {
				t.Errorf("Spec.UpdateContexts() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSpec_UpdateiContexts(t *testing.T) {
	type fields struct {
		Owner            string
		ContextVariables []*Context
		Contexts         func(t *testing.T, ctrl *gomock.Controller) *mock.MockContexts
		dryRun           bool
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "Create new context variables",
			fields: fields{
				Owner: "org1",
				ContextVariables: []*Context{
					{
						Name: "ctx1",
						Variables: []*Variable{
							{
								Name:  "SECRET1",
								Value: "111",
							},
							{
								Name:  "SECRET2",
								Value: "222",
							},
						},
					},
				},
				Contexts: func(t *testing.T, ctrl *gomock.Controller) *mock.MockContexts {
					t.Helper()

					ctx := context.Background()
					mock := mock.NewMockContexts(ctrl)
					mock.EXPECT().List(ctx, circleci.ContextListOptions{
						OwnerSlug: circleci.String("org1"),
						PageToken: circleci.String(""),
					}).Return(&circleci.ContextList{
						Items: []*circleci.Context{
							{
								ID:   "ctx-1",
								Name: "ctx1",
							},
						},
					}, nil)
					mock.EXPECT().AddOrUpdateVariable(ctx, "ctx-1", "SECRET1", circleci.ContextAddOrUpdateVariableOptions{
						Value: circleci.String("111"),
					})
					mock.EXPECT().AddOrUpdateVariable(ctx, "ctx-1", "SECRET2", circleci.ContextAddOrUpdateVariableOptions{
						Value: circleci.String("222"),
					})
					return mock
				},
			},
		},
		{
			name: "Met conditions for creating new context variables but doesn't do destructive changes in dry-run mode",
			fields: fields{
				Owner: "org1",
				ContextVariables: []*Context{
					{
						Name: "ctx1",
						Variables: []*Variable{
							{
								Name:  "SECRET1",
								Value: "111",
							},
							{
								Name:  "SECRET2",
								Value: "222",
							},
						},
					},
				},
				Contexts: func(t *testing.T, ctrl *gomock.Controller) *mock.MockContexts {
					t.Helper()

					ctx := context.Background()
					mock := mock.NewMockContexts(ctrl)
					mock.EXPECT().List(ctx, circleci.ContextListOptions{
						OwnerSlug: circleci.String("org1"),
						PageToken: circleci.String(""),
					}).Return(&circleci.ContextList{
						Items: []*circleci.Context{
							{
								ID:   "ctx-1",
								Name: "ctx1",
							},
						},
					}, nil)
					return mock
				},
				dryRun: true,
			},
		},
		{
			name: "Update an existing context variable",
			fields: fields{
				Owner: "org1",
				ContextVariables: []*Context{
					{
						Name: "ctx1",
						Variables: []*Variable{
							{
								Name:  "SECRET1",
								Value: "111",
							},
						},
					},
				},
				Contexts: func(t *testing.T, ctrl *gomock.Controller) *mock.MockContexts {
					t.Helper()

					ctx := context.Background()
					mock := mock.NewMockContexts(ctrl)
					mock.EXPECT().List(ctx, circleci.ContextListOptions{
						OwnerSlug: circleci.String("org1"),
						PageToken: circleci.String(""),
					}).Return(&circleci.ContextList{
						Items: []*circleci.Context{
							{
								ID:   "ctx-1",
								Name: "ctx1",
							},
						},
					}, nil)
					mock.EXPECT().AddOrUpdateVariable(ctx, "ctx-1", "SECRET1", circleci.ContextAddOrUpdateVariableOptions{
						Value: circleci.String("111"),
					})
					return mock
				},
			},
		},
		{
			name: "Met conditions for updating an existing context variable but doesn't do destructive changes in dry-run mode",
			fields: fields{
				Owner: "org1",
				ContextVariables: []*Context{
					{
						Name: "ctx1",
						Variables: []*Variable{
							{
								Name:  "SECRET1",
								Value: "111",
							},
						},
					},
				},
				Contexts: func(t *testing.T, ctrl *gomock.Controller) *mock.MockContexts {
					t.Helper()

					ctx := context.Background()
					mock := mock.NewMockContexts(ctrl)
					mock.EXPECT().List(ctx, circleci.ContextListOptions{
						OwnerSlug: circleci.String("org1"),
						PageToken: circleci.String(""),
					}).Return(&circleci.ContextList{
						Items: []*circleci.Context{
							{
								ID:   "ctx-1",
								Name: "ctx1",
							},
						},
					}, nil)
					return mock
				},
				dryRun: true,
			},
		},
		{
			name: "It returns an error when the matching context is not found",
			fields: fields{
				Owner: "org1",
				ContextVariables: []*Context{
					{
						Name: "ctx-not-found",
						Variables: []*Variable{
							{
								Name:  "SECRET1",
								Value: "111",
							},
						},
					},
				},
				Contexts: func(t *testing.T, ctrl *gomock.Controller) *mock.MockContexts {
					t.Helper()

					ctx := context.Background()
					mock := mock.NewMockContexts(ctrl)
					mock.EXPECT().List(ctx, circleci.ContextListOptions{
						OwnerSlug: circleci.String("org1"),
						PageToken: circleci.String(""),
					}).Return(&circleci.ContextList{
						Items: []*circleci.Context{
							{
								ID:   "ctx-1",
								Name: "ctx1",
							},
						},
					}, nil)
					return mock
				},
				dryRun: true,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockCircleCIAPI := &circleci.Client{}
			mockCircleCIAPI.Contexts = tt.fields.Contexts(t, ctrl)
			s := &Spec{
				Owner:     tt.fields.Owner,
				Contexts:  tt.fields.ContextVariables,
				RateLimit: ratelimit.New(apiRateLimit),
			}
			if err := s.UpdateContexts(context.Background(), tt.fields.dryRun, mockCircleCIAPI, s.RateLimit); (err != nil) != tt.wantErr {
				t.Errorf("Spec.UpdateContexts() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
