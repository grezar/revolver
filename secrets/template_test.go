package secrets

import (
	"context"
	"testing"
)

func TestExecuteTemplate(t *testing.T) {
	type args struct {
		ctx  context.Context
		node string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "Valid template",
			args: args{
				ctx: WithSecrets(context.Background(), Secrets{
					"AWSAccessKeyID":     "SAMPLE_ID",
					"AWSSecretAccessKey": "SAMPLE_SECRET",
				}),
				node: "{{ .AWSAccessKeyID }}",
			},
			want: "SAMPLE_ID",
		},
		{
			name: "Symbols should NOT be escaped",
			args: args{
				ctx: WithSecrets(context.Background(), Secrets{
					"AWSSecretAccessKey": "!@#$%^&*()_+-\\SECRET",
				}),
				node: "{{ .AWSSecretAccessKey }}",
			},
			want: "!@#$%^&*()_+-\\SECRET",
		},
		{
			name: "Invalid template",
			args: args{
				ctx: WithSecrets(context.Background(), Secrets{
					"AWSAccessKeyID":     "SAMPLE_ID",
					"AWSSecretAccessKey": "SAMPLE_SECRET",
				}),
				node: "{{ undefinedFunc .AWSAccessKeyID }}",
			},
			wantErr: true,
		},
		{
			name: "Pure string",
			args: args{
				ctx: WithSecrets(context.Background(), Secrets{
					"AWSAccessKeyID":     "SAMPLE_ID",
					"AWSSecretAccessKey": "SAMPLE_SECRET",
				}),
				node: "PURE_STRING",
			},
			want: "PURE_STRING",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ExecuteTemplate(tt.args.ctx, tt.args.node)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExecuteTemplate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ExecuteTemplate() = %v, want %v", got, tt.want)
			}
		})
	}
}
