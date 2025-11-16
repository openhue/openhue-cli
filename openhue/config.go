package openhue

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/openhue/openhue-go"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"net/http"
	"openhue-cli/util/logger"
	"os"
	"path/filepath"
	"slices"
)

// CommandsWithNoConfig contains the list of commands that don't require the configuration to exist
var CommandsWithNoConfig = []string{"setup", "config", "help", "discover", "auth", "version", "completion"}
var configPath string

func init() {

	// If the $XDG_CONFIG_HOME env var is set, use it instead of the home user home dir
	xdgConfigHome := os.Getenv("XDG_CONFIG_HOME")
	if len(xdgConfigHome) > 0 {
		configPath = filepath.Join(xdgConfigHome, "openhue")
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)
		configPath = filepath.Join(home, ".openhue")
	}

	_ = os.MkdirAll(configPath, os.ModePerm)
}

type Config struct {
	// The IP of the Philips HUE Bridge
	Bridge string
	// The HUE Application Key
	Key      string
	LogLevel string
}

func (c *Config) GetConfigFile() string {
	return filepath.Join(configPath, "config.yaml")
}

func (c *Config) Load() {
	viper.SetConfigFile(c.GetConfigFile())

	// When trying to run CLI without configuration
	configDoesNotExist := viper.ReadInConfig() != nil
	if configDoesNotExist && len(os.Args) > 1 && !slices.Contains(CommandsWithNoConfig, os.Args[1]) {
		fmt.Println("\nopenhue-cli not configured yet, please run the 'setup' command")
		os.Exit(0)
	}

	c.Bridge = viper.GetString("Bridge")
	c.Key = viper.GetString("Key")
	c.LogLevel = viper.GetString("log_level")
	logger.Init(filepath.Join(configPath, "openhue.log"), c.LogLevel)
}

func (c *Config) Save() (string, error) {

	if len(c.Bridge) == 0 {
		return "", errors.New("'bridge' value not set in config")
	}

	if len(c.Key) == 0 {
		return "", errors.New("'key' value not set in config")
	}

	viper.Set("Bridge", c.Bridge)
	viper.Set("Key", c.Key)

	err := viper.SafeWriteConfig()
	if err != nil {
		return c.GetConfigFile(), viper.WriteConfig()
	}

	return c.GetConfigFile(), nil
}

// NewOpenHueClient Creates a new NewClientWithResponses for a given server and hueApplicationKey.
// This function will also skip SSL verification, as the Philips HUE Bridge exposes a self-signed certificate.
func (c *Config) NewOpenHueClient() *openhue.ClientWithResponses {
	apiKeyAuth := func(ctx context.Context, req *http.Request) error {
		req.Header.Set("hue-application-key", c.Key)
		return nil
	}

	// skip SSL Verification
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	client, err := openhue.NewClientWithResponses("https://"+c.Bridge, openhue.WithRequestEditorFn(apiKeyAuth))
	cobra.CheckErr(err)

	return client
}

// NewOpenHueClientNoAuth Creates a new NewClientWithResponses for a given server and no application Key.
// This function will also skip SSL verification, as the Philips HUE Bridge exposes a self-signed certificate.
func NewOpenHueClientNoAuth(ip string) *openhue.ClientWithResponses {

	// skip SSL Verification
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	client, err := openhue.NewClientWithResponses("https://" + ip)
	cobra.CheckErr(err)

	return client
}
