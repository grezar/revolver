package secrets

import (
	"context"
	"strings"
	"text/template"
)

func ExecuteTemplate(ctx context.Context, node string) (string, error) {
	tmpl, err := template.New("").Parse(node)
	if err != nil {
		return "", err
	}

	writer := new(strings.Builder)
	err = tmpl.Execute(writer, GetSecrets(ctx))
	if err != nil {
		return "", err
	}
	return writer.String(), nil
}
