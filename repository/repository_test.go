package repository

import "testing"

func TestRef(t *testing.T) {
	t.Run("key found", func(t *testing.T) {
		repo := &Repository{}
		testKey := "key"
		expected := "value"
		repo.Secrets = map[string]string{
			testKey: expected,
		}
		actual, err := repo.Ref(testKey)
		if err != nil {
			t.Fatal(err)
		}
		if actual != expected {
			t.Fatal(err)
		}
	})

	t.Run("key not found", func(t *testing.T) {
		repo := &Repository{}
		expected := "value"
		repo.Secrets = map[string]string{
			"key": expected,
		}
		_, err := repo.Ref("notFoundKey")
		if err != secretNotFoundError {
			t.Fatal(err)
		}
	})
}
