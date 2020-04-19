package cmd

import (
	"github.com/shopnado/cli/cmd/profile"
	"github.com/shopnado/cli/cmd/webhook"
	"github.com/urfave/cli/v2"
)

func Commands() cli.Commands {
	return cli.Commands{
		webhook.Command(),
		profile.Command(),
	}
}
