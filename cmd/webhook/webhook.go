package webhook

import (
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	goshopify "github.com/bold-commerce/go-shopify"
	"github.com/shopnado/cli/profile"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

func Command() *cli.Command {
	return &cli.Command{
		Name:    "webhook",
		Usage:   "",
		Action:  list,
		Aliases: []string{"wh"},
		Flags:   []cli.Flag{},
		Before:  profile.FromContext,
		Subcommands: []*cli.Command{
			{
				Name:   "list",
				Action: list,
			},
			{
				Name:   "create",
				Action: create,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "address",
						Aliases: []string{"a"},
						Usage:   "Website URL Address for the webhook to POST, https:// required",
					},
					&cli.StringFlag{
						Name:    "topic",
						Aliases: []string{"t"},
						Usage:   "Name of Shopify webhook topic to subscribe",
					},
					&cli.StringFlag{
						Name:    "format",
						Aliases: []string{"f"},
						Usage:   "Fromat for the webhook, default JSON",
						Value:   "json",
					},
				},
			},
			{
				Name:   "read",
				Action: read,
			},
			{
				Name:   "update",
				Action: update,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "id",
						Usage: "Webhook ID to update",
					},
					&cli.StringFlag{
						Name:    "address",
						Aliases: []string{"a"},
						Usage:   "Website URL Address for the webhook to POST, https:// required",
					},
					&cli.StringFlag{
						Name:    "topic",
						Aliases: []string{"t"},
						Usage:   "Name of Shopify webhook topic to subscribe",
					},
					&cli.StringFlag{
						Name:    "format",
						Aliases: []string{"f"},
						Usage:   "Fromat for the webhook, default JSON",
						Value:   "json",
					},
				},
			},
			{
				Name:   "delete",
				Action: delete,
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "all",
						Aliases: []string{"a", "A"},
						Usage:   "Delete all webhooks",
					},
				},
			},
			{
				Name:   "topics",
				Action: topics,
				Usage:  "shopnado webhook topics <api version>, version defaults to \"stable\"",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "all",
						Aliases: []string{"a", "A"},
						Usage:   "List all the topics from all API versions",
					},
				},
			},
		},
	}
}

func list(c *cli.Context) error {
	client := profile.NewShopifyClient()
	webhooks, err := client.Webhook.List(goshopify.ListOptions{})
	if err != nil {
		return err
	}

	if len(webhooks) == 0 {
		logrus.Println("no webhooks found")
		return nil
	}

	for _, wh := range webhooks {
		logrus.Printf("%d - %s\n", wh.ID, wh.Topic)
	}

	return nil
}

func create(c *cli.Context) error {
	address := c.String("address")
	topic := c.String("topic")
	format := c.String("format")
	if topic == "" || address == "" || format == "" {
		return errors.New("address, topic and format are all required to make a webhook")
	}

	client := profile.NewShopifyClient()
	wh, err := client.Webhook.Create(goshopify.Webhook{
		Address: address,
		Topic:   topic,
		Format:  format,
	})

	if err != nil {
		return err
	}

	logrus.Printf("webhook created: %d\n", wh.ID)
	return nil
}

func read(c *cli.Context) error {
	webhookId, err := strconv.ParseInt(c.Args().First(), 10, 64)
	if err != nil {
		return err
	}

	client := profile.NewShopifyClient()

	wh, err := client.Webhook.Get(webhookId, nil)
	if err != nil {
		return err
	}

	logrus.Printf("%d : %s - %s", wh.ID, wh.Topic, wh.Address)
	return nil
}

func update(c *cli.Context) error {
	webhookId, err := strconv.ParseInt(c.Args().First(), 10, 64)
	if err != nil {
		return err
	}

	client := profile.NewShopifyClient()

	wh, err := client.Webhook.Get(webhookId, nil)
	if err != nil {
		return err
	}

	address := c.String("address")
	if address != "" {
		wh.Address = address
	}
	logrus.Println(address)
	topic := c.String("topic")
	if topic != "" {
		wh.Topic = topic
	}
	format := c.String("format")
	if format != "" {
		wh.Format = format
	}

	logrus.Println(wh)
	_, err = client.Webhook.Update(*wh)
	return err
}

func delete(c *cli.Context) error {
	if c.Bool("all") {
		return clear(c)
	}
	webhookId, err := strconv.ParseInt(c.Args().First(), 10, 64)
	if err != nil {
		return err
	}

	client := profile.NewShopifyClient()
	if err := client.Webhook.Delete(webhookId); err != nil {
		return err
	}

	logrus.Printf("webhook deleted %d", webhookId)
	return nil
}

func clear(c *cli.Context) error {
	client := profile.NewShopifyClient()
	webhooks, err := client.Webhook.List(goshopify.ListOptions{})
	if err != nil {
		return err
	}

	for _, wh := range webhooks {
		if err := client.Webhook.Delete(wh.ID); err != nil {
			logrus.Printf("error deleting webhook %d", wh.ID)
		} else {
			logrus.Printf("webhook deleted %d", wh.ID)
		}
	}

	return nil
}

func topics(c *cli.Context) error {
	all := make(map[string][]string)
	versions := []string{"stable", "2019-04", "2019-07", "2019-10", "2020-01", "2020-04"}
	var cli *goshopify.Client

	filter := c.Args().First()
	if filter == "" {
		filter = "stable"
	}
	if sort.SearchStrings(versions, filter) != len(versions) {
		return fmt.Errorf("invalid filter %s, allowed: %s", filter, strings.Join(versions, " "))
	}

	if c.Bool("all") {
		filter = ""
	}

	for _, version := range versions {
		if filter != "" && filter != version {
			continue
		}

		if version == "stable" {
			cli = profile.NewShopifyClient()
		} else {
			cli = profile.NewShopifyVersionedClient(version)
		}
		_, err := cli.Webhook.Create(goshopify.Webhook{
			Address: "https://asdf.com",
			Topic:   "asdf/asdf",
			Format:  "json",
		})

		logrus.Debugln(err)
		re := regexp.MustCompile("[a-z_]+/[a-z_]+")
		matches := re.FindAllStringSubmatch(err.Error(), -1)
		for _, t := range matches {
			for _, match := range t {
				all[version] = append(all[version], strings.Trim(match, " "))
			}
		}
		sort.Strings(all[version])
	}

	i := len(all)
	if len(all) == 1 {
		logrus.Printf("%s", strings.Join(all[filter], "\n"))
		return nil
	}

	for api, topics := range all {
		logrus.Printf("API: %s\n%s\n", api, strings.Join(topics, "\n"))
		if i > 1 {
			logrus.Printf("\n")
		}
		i--
	}

	return nil
}
