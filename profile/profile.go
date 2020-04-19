package profile

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"

	goshopify "github.com/bold-commerce/go-shopify"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v2"
)

const (
	EnvShopname     = "SHOPNADO_SHOPNAME"
	EnvApikey       = "SHOPNADO_APIKEY"
	EnvPassword     = "SHOPNADO_PASSWORD"
	DefaultFilename = "~/.shopnado/config.yaml"
)

type Config map[string]Profile
type Profile struct {
	ShopName string `yaml:"shopname"`
	ApiKey   string `yaml:"apikey"`
	Password string `yaml:"password"`
}

var profile *Profile

func NewProfile(shopname, apikey, password string) *Profile {
	return &Profile{
		ShopName: shopname,
		ApiKey:   apikey,
		Password: password,
	}
}

func Set(p *Profile) {
	profile = p
}

func Get() *Profile {
	return profile
}

func NewShopifyClient() *goshopify.Client {
	if profile == nil {
		return nil
	}
	app := goshopify.App{
		ApiKey:   profile.ApiKey,
		Password: profile.Password,
	}
	return goshopify.NewClient(app, profile.ShopName, "")
}

func NewShopifyVersionedClient(version string) *goshopify.Client {
	if profile == nil {
		return nil
	}
	app := goshopify.App{
		ApiKey:   profile.ApiKey,
		Password: profile.Password,
	}
	return goshopify.NewClient(app,
		profile.ShopName,
		"", goshopify.WithVersion(version))
}

func LoadFromConfig(f, p string) (*Profile, error) {
	if f == "" || p == "" {
		return nil, fmt.Errorf("config filename and profile required: given %s %s", f, p)
	}

	c, err := GetConfig(f)
	if err != nil {
		return nil, err
	}

	loadedProfile, ok := c[p]
	if !ok {
		return nil, fmt.Errorf("no profile %s in %s", p, f)
	}

	return &loadedProfile, nil
}

func FromContext(c *cli.Context) error {
	var shopname string
	var apikey string
	var password string

	// os env vars
	shopname = os.Getenv(EnvShopname)
	apikey = os.Getenv(EnvApikey)
	password = os.Getenv(EnvPassword)
	if shopname != "" &&
		apikey != "" &&
		password != "" {
		logrus.Debugf("loading shop credentials from environment variables")
		Set(NewProfile(shopname, apikey, password))
		return nil
	}

	// cli flags
	shopname = c.String("shopname")
	apikey = c.String("apikey")
	password = c.String("password")
	if shopname != "" &&
		apikey != "" &&
		password != "" {
		logrus.Debugf("loading shop credentials from cli flags")
		Set(NewProfile(shopname, apikey, password))
		return nil
	}

	// from config.yaml
	filename, err := GetConfigFilename(c.String("config"))
	if err != nil {
		return err
	}
	if fileExists(filename) {
		logrus.Debugf("loading shop credentials from %s", c.String("config"))
		profile, err := LoadFromConfig(c.String("config"), c.String("profile"))
		if err != nil {
			return err
		}
		Set(profile)
		return nil
	}

	return fmt.Errorf(
		"unable to load credentials from environment, flags or %s",
		c.String("config"))
}

func ConfigTouch(f string) error {
	filename, err := homedir(f)
	if err != nil {
		return err
	}

	if fileExists(filename) {
		return nil // already exists
	}

	config, err := os.Create(filename)
	if err != nil {
		return err
	}

	return config.Close()
}

func GetConfigFilename(f string) (string, error) {
	filename, err := homedir(f)
	if err != nil {
		return "", err
	}
	return filename, nil
}

func GetConfig(f string) (Config, error) {
	filename, err := homedir(f)
	if err != nil {
		return nil, err
	}

	if !fileExists(filename) {
		return nil, fmt.Errorf("file does not exist: %s", filename)
	}

	yamlContents, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	c := Config{}
	if err := yaml.Unmarshal(yamlContents, &c); err != nil {
		return nil, err
	}
	return c, nil
}

func WriteConfig(c Config, f string) error {
	filename, err := homedir(f)
	if err != nil {
		return err
	}

	contents, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(filename, contents, 0644)
}

func DeleteConfig(f string) error {
	filename, err := homedir(f)
	if err != nil {
		return err
	}

	return os.Remove(filename)
}

func homedir(filename string) (string, error) {
	if strings.Contains(filename, "~/") {
		homedir, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		filename = strings.Replace(filename, "~/", "", 1)
		filename = path.Join(homedir, filename)
	}
	return filename, nil
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
