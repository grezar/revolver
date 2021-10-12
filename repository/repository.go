package repository

import "errors"

var secretNotFoundError = errors.New("secret not found")

type Repository struct {
	Secrets map[string]string
}

func (r *Repository) Ref(key string) (string, error) {
	v := r.Secrets[key]
	if v == "" {
		return "", secretNotFoundError
	}
	return v, nil
}
