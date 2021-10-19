package main

import (
	"fmt"
	"os"

	"github.com/grezar/revolver"
	"github.com/urfave/cli/v2"
)

func main() {
  app := &cli.App{
    Commands: []*cli.Command{
      {
        Name:    "rotate",
        Usage:   "Rotate secrets based on configured YAML",
        Flags: []cli.Flag{
          &cli.StringFlag{
            Name:    "config",
            Aliases: []string{"c"},
            Usage:   "Load configuration from `FILE`",
            Required: true,
          },
        },
        Action:  func(c *cli.Context) error {
          runner := revolver.NewRunner(c.String("config"))
          if err := runner.Run(); err != nil {
            return err
          }
          return nil
        },
      },
    },
  }

  err := app.Run(os.Args)
  if err != nil {
    fmt.Fprintln(os.Stderr, err)
    os.Exit(1)
  }
}
