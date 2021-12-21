package main

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/grezar/revolver"
	"github.com/grezar/revolver/reporting"
	"github.com/urfave/cli/v2"
)

// These variables are set in build step.
var (
	Version  string
	Revision string
)

func main() {
	app := &cli.App{
		Commands: []*cli.Command{
			{
				Name:  "version",
				Usage: "Print version",
				Action: func(c *cli.Context) error {
					fmt.Println("revolver", Version, Revision)
					return nil
				},
			},
			{
				Name:  "rotate",
				Usage: "Rotate secrets based on configured YAML",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "config",
						Aliases:  []string{"c"},
						Usage:    "Load configuration from `FILE`",
						Required: true,
					},
					&cli.BoolFlag{
						Name:    "dry-run",
						Aliases: []string{"d"},
						Usage:   "Dry run",
					},
				},
				Action: func(c *cli.Context) error {
					runner, err := revolver.NewRunner(c.String("config"), c.Bool("dry-run"))
					if err != nil {
						return err
					}
					ok := reporting.Run(func(rptr *reporting.R) {
						runner.Run(rptr)
					})
					if !ok {
						return errors.New("failed to execute rotations")
					}
					return nil
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
