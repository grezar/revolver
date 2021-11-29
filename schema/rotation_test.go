package schema

import (
	"os"
	"testing"

	_ "github.com/grezar/revolver/provider/from/awsiamuser"
	_ "github.com/grezar/revolver/provider/from/stdin"
	_ "github.com/grezar/revolver/provider/to/awssharedcredentials"
	_ "github.com/grezar/revolver/provider/to/circleci"
	_ "github.com/grezar/revolver/provider/to/stdout"
	_ "github.com/grezar/revolver/provider/to/tfe"
)

func TestLoadRotations(t *testing.T) {
	tests := []string{"./../testdata/valid.yml"}

	for _, test := range tests {
		f, err := os.Open(test)
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()

		rotations, err := LoadRotations(f)
		if err != nil {
			t.Error(err)
		}
		if len(rotations) == 0 {
			t.Error("no rotations found")
		}
	}
}
