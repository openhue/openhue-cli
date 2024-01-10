package openhue

import (
	"crypto/tls"
	"errors"
	"fmt"
	sp "github.com/deepmap/oapi-codegen/pkg/securityprovider"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"net/http"
	"openhue-cli/openhue/gen"
	"openhue-cli/util/logger"
	"os"
	"path/filepath"
	"slices"
)

// CommandsWithNoConfig contains the list of commands that don't require the configuration to exist
var CommandsWithNoConfig = []string{"configure", "help", "discover", "auth", "version", "completion"}

type Config struct {
	// The IP of the Philips HUE Bridge
	Bridge string
	// The HUE Application Key
	Key string
}

func (c *Config) Load() {

	// Find home directory.
	home, err := os.UserHomeDir()
	cobra.CheckErr(err)

	var configPath = filepath.Join(home, "/.openhue")
	_ = os.MkdirAll(configPath, os.ModePerm)

	logger.Init(configPath)

	// Search config in home directory with name ".openhue" (without an extension).
	viper.AddConfigPath(configPath)
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	// When trying to run CLI without configuration
	configDoesNotExist := viper.ReadInConfig() != nil
	if configDoesNotExist && len(os.Args) > 1 && !slices.Contains(CommandsWithNoConfig, os.Args[1]) {
		fmt.Println("\nopenhue-cli not configured yet, please run the 'configure' command")
		os.Exit(0)
	}

	c.Bridge = viper.GetString("Bridge")
	c.Key = viper.GetString("Key")
}

func (c *Config) Save() error {

	if len(c.Bridge) == 0 {
		return errors.New("'bridge' value not set in config")
	}

	if len(c.Key) == 0 {
		return errors.New("'key' value not set in config")
	}

	viper.Set("Bridge", c.Bridge)
	viper.Set("Key", c.Key)

	err := viper.SafeWriteConfig()
	if err != nil {
		return viper.WriteConfig()
	}

	return nil
}

// NewOpenHueClient Creates a new NewClientWithResponses for a given server and hueApplicationKey.
// This function will also skip SSL verification, as the Philips HUE Bridge exposes a self-signed certificate.
func (c *Config) NewOpenHueClient() *gen.ClientWithResponses {
	p, err := sp.NewSecurityProviderApiKey("header", "hue-application-Key", c.Key)
	cobra.CheckErr(err)

	// skip SSL Verification
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	client, err := gen.NewClientWithResponses("https://"+c.Bridge, gen.WithRequestEditorFn(p.Intercept))
	cobra.CheckErr(err)

	return client
}

// NewOpenHueClientNoAuth Creates a new NewClientWithResponses for a given server and no application Key.
// This function will also skip SSL verification, as the Philips HUE Bridge exposes a self-signed certificate.
func NewOpenHueClientNoAuth(ip string) *gen.ClientWithResponses {

	// skip SSL Verification
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	client, err := gen.NewClientWithResponses("https://" + ip)
	cobra.CheckErr(err)

	return client
}
