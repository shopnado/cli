package profile

import (
	"errors"
	"fmt"
	"os"
	"os/exec"

	"github.com/shopnado/cli/profile"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

func Command() *cli.Command {
	return &cli.Command{
		Name:    "profile",
		Usage:   "",
		Action:  list,
		Aliases: []string{"p"},
		Flags:   []cli.Flag{},
		Subcommands: []*cli.Command{
			{
				Name:   "list",
				Action: list,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "config",
						Aliases: []string{"c"},
						Usage:   "Config file, default ~/.shopnado/config.yaml",
						Value:   profile.DefaultFilename,
					},
				},
			},
			{
				Name:   "create",
				Action: create,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "config",
						Aliases: []string{"c"},
						Usage:   "Config file, default ~/.shopnado/config.yaml",
						Value:   profile.DefaultFilename,
					},
					&cli.StringFlag{
						Name:    "name",
						Aliases: []string{"n"},
						Usage:   "Name for the profile",
					},
					&cli.StringFlag{
						Name:    "apikey",
						Aliases: []string{"k", "key"},
						Usage:   "API Key for Shopify",
					},
					&cli.StringFlag{
						Name:    "password",
						Aliases: []string{"p", "pass"},
						Usage:   "API Password for Shopify",
					},
					&cli.StringFlag{
						Name:    "shopname",
						Aliases: []string{"s", "shop"},
						Usage:   "Shopify shop name, eg <shopname>.myshopify.com",
					},
				},
			},
			{
				Name:   "read",
				Action: read,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "config",
						Aliases: []string{"c"},
						Usage:   "Config file, default ~/.shopnado/config.yaml",
						Value:   profile.DefaultFilename,
					},
					&cli.StringFlag{
						Name:    "name",
						Aliases: []string{"n"},
						Usage:   "Name for the profile",
					},
				},
			},
			{
				Name:   "update",
				Action: update,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "config",
						Aliases: []string{"c"},
						Usage:   "Config file, default ~/.shopnado/config.yaml",
						Value:   profile.DefaultFilename,
					},
					&cli.StringFlag{
						Name:    "name",
						Aliases: []string{"n"},
						Usage:   "Name for the profile",
					},
				},
			},
			{
				Name:   "delete",
				Action: del,
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "all",
						Aliases: []string{"a", "A"},
						Usage:   "Delete all entries in the config file",
					},
					&cli.StringFlag{
						Name:    "name",
						Aliases: []string{"n"},
						Usage:   "Name for the profile",
					},
					&cli.StringFlag{
						Name:    "config",
						Aliases: []string{"c"},
						Usage:   "Config file, default ~/.shopnado/config.yaml",
						Value:   profile.DefaultFilename,
					},
				},
			},
			{
				Name:   "edit",
				Action: edit,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "editor",
						Aliases: []string{"e"},
						Usage:   "Which file editor to use, default vim",
						Value:   "vim",
					},
					&cli.StringFlag{
						Name:    "config",
						Aliases: []string{"c"},
						Usage:   "Config file, default ~/.shopnado/config.yaml",
						Value:   profile.DefaultFilename,
					},
				},
			},
		},
	}
}

func list(c *cli.Context) error {
	config, err := profile.GetConfig(c.String("config"))
	if err != nil {
		return err
	}

	for name, profile := range config {
		logrus.Printf("%s [%s]\n", name, profile.ShopName)
	}

	return nil
}

func create(c *cli.Context) error {
	shopname := c.String("shopname")
	apikey := c.String("apikey")
	password := c.String("password")
	profileName := c.String("name")
	filename := c.String("config")

	if profileName == "" {
		return errors.New("profile name is required")
	}

	if err := profile.ConfigTouch(filename); err != nil {
		return err
	}

	config, err := profile.GetConfig(filename)
	if err != nil {
		return err
	}

	if shopname != "" && apikey != "" && password != "" {
		config[profileName] = *profile.NewProfile(shopname, apikey, password)
		if err := profile.WriteConfig(config, filename); err != nil {
			return err
		}
	}

	return nil
}

func read(c *cli.Context) error {
	profileName := c.String("name")
	if profileName == "" {
		return errors.New("profile name is required")
	}

	config, err := profile.GetConfig(c.String("config"))
	if err != nil {
		return err
	}

	p, ok := config[profileName]
	if !ok {
		return fmt.Errorf("profile not found: %s", profileName)
	}

	logrus.Printf("shopname: %s\napikey: %s\npassword: %s",
		p.ShopName, p.ApiKey, p.Password)

	return nil
}

func update(c *cli.Context) error {
	// todo
	return nil
}

func del(c *cli.Context) error {
	profileName := c.String("name")
	filename := c.String("config")

	if c.Bool("all") {
		logrus.Printf("deleting all profiles in %s", filename)
		return profile.DeleteConfig(filename)
	}

	if profileName == "" {
		return errors.New("profile name is required")
	}

	logrus.Printf("deleting profile %s from %s", profileName, filename)
	config, err := profile.GetConfig(filename)
	if err != nil {
		return err
	}

	if _, ok := config[profileName]; !ok {
		return fmt.Errorf("no profile with name %s", profileName)
	}

	delete(config, profileName)
	if err := profile.WriteConfig(config, filename); err != nil {
		return err
	}

	return nil
}

func edit(c *cli.Context) error {
	filename, err := profile.GetConfigFilename(c.String("config"))
	if err != nil {
		return err
	}

	cmd := exec.Command(c.String("editor"), filename)
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
