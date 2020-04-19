package main

import (
	"io/ioutil"
	"log"
	"os"

	"github.com/shopnado/cli/cmd"
	"github.com/shopnado/cli/profile"
	"github.com/sirupsen/logrus"
	easy "github.com/t-tomalak/logrus-easy-formatter"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:        "shopnado",
		Usage:       "",
		Description: "CLI tool for ",
		Commands:    cmd.Commands(),
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "config",
				Aliases: []string{"c"},
				Usage:   "Filepath to the config yaml file, default ~/.shopnado/config.yaml",
				Value:   profile.DefaultFilename,
			},
			&cli.StringFlag{
				Name:    "profile",
				Aliases: []string{"p"},
				Usage:   "Profile to use from the config.yaml file.",
				Value:   "default",
			},
			&cli.StringFlag{
				Name:  "apikey",
				Usage: "API Key for Shopify",
			},
			&cli.StringFlag{
				Name:  "password",
				Usage: "API Password for Shopify",
			},
			&cli.StringFlag{
				Name:  "shopname",
				Usage: "Shopify shop name, eg <shopname>.myshopify.com",
			},
			&cli.BoolFlag{
				Name:    "debug",
				Aliases: []string{"d"},
				Usage:   "Turn on verbose debug logging",
			},
			&cli.BoolFlag{
				Name:    "quiet",
				Aliases: []string{"q"},
				Usage:   "Turn on off all logging",
			},
		},
		Before: func(c *cli.Context) error {
			if c.Bool("debug") {
				logrus.SetLevel(logrus.DebugLevel)
			} else {
				// treat logrus like fmt.Print
				logrus.SetFormatter(&easy.Formatter{
					LogFormat: "%msg%",
				})
			}
			if c.Bool("quiet") {
				logrus.SetOutput(ioutil.Discard)
			}

			return nil
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
