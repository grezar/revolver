package circleci

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/grezar/go-circleci"
	mock "github.com/grezar/go-circleci/mocks"
)

func TestSpec_UpdateProjectVariables(t *testing.T) {
	type fields struct {
		Owner            string
		ProjectVariables []*ProjectVariable
		ContextVariables []*Context
		Projects         func(t *testing.T, ctrl *gomock.Controller) *mock.MockProjects
		Contexts         func(t *testing.T, ctrl *gomock.Controller) *mock.MockContexts
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockCircleCIAPI := &circleci.Client{}
			mockCircleCIAPI.Projects = tt.fields.Projects(t, ctrl)
			s := &Spec{
				Owner:            tt.fields.Owner,
				ProjectVariables: tt.fields.ProjectVariables,
			}
			if err := s.UpdateProjectVariables(context.Background(), mockCircleCIAPI); (err != nil) != tt.wantErr {
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
					}).Return(&circleci.ContextList{
						Items: []*circleci.Context{},
					}, nil)
					mock.EXPECT().AddOrUpdateVariable(ctx, "", "SECRET1", circleci.AddOrUpdateVariableOptions{
						Value: circleci.String("111"),
					})
					mock.EXPECT().AddOrUpdateVariable(ctx, "", "SECRET2", circleci.AddOrUpdateVariableOptions{
						Value: circleci.String("222"),
					})
					return mock
				},
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
					}).Return(&circleci.ContextList{
						Items: []*circleci.Context{
							{
								ID:   "ctx-1",
								Name: "ctx1",
							},
						},
					}, nil)
					mock.EXPECT().AddOrUpdateVariable(ctx, "ctx-1", "SECRET1", circleci.AddOrUpdateVariableOptions{
						Value: circleci.String("111"),
					})
					return mock
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockCircleCIAPI := &circleci.Client{}
			mockCircleCIAPI.Contexts = tt.fields.Contexts(t, ctrl)
			s := &Spec{
				Owner:    tt.fields.Owner,
				Contexts: tt.fields.ContextVariables,
			}
			if err := s.UpdateContexts(context.Background(), mockCircleCIAPI); (err != nil) != tt.wantErr {
				t.Errorf("Spec.UpdateContexts() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
