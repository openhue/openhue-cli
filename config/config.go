package config

import (
	"crypto/tls"
	sp "github.com/deepmap/oapi-codegen/pkg/securityprovider"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"net/http"
	"openhue-cli/openhue"
	"os"
	"path/filepath"
	"slices"
)

type Config struct {
	// The IP of the Philips HUE bridge
	bridge string
	// The HUE Application Key
	key string
}

func (c *Config) Load() {

	// Find home directory.
	home, err := os.UserHomeDir()
	cobra.CheckErr(err)

	var configPath = filepath.Join(home, "/.openhue")
	_ = os.MkdirAll(configPath, os.ModePerm)

	// Search config in home directory with name ".openhue" (without an extension).
	viper.AddConfigPath(configPath)
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	// List of commands that does not require configuration
	ignoredCmds := []string{"setup", "help", "discover", "auth"}

	// When trying to run CLI without configuration
	if err := viper.ReadInConfig(); err != nil && !slices.Contains(ignoredCmds, os.Args[1]) {
		color.New(color.FgRed).Add(color.Bold).Println("\nopenhue-cli not configured yet, please run the 'setup' command")
		os.Exit(0)
	}

	c.bridge = viper.GetString("bridge")
	c.key = viper.GetString("key")
}

// NewOpenHueClient Creates a new NewClientWithResponses for a given server and hueApplicationKey.
// This function will also skip SSL verification, as the Philips HUE Bridge exposes a self-signed certificate.
func (c *Config) NewOpenHueClient() *openhue.ClientWithResponses {
	p, err := sp.NewSecurityProviderApiKey("header", "hue-application-key", c.key)
	cobra.CheckErr(err)

	// skip SSL Verification
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	client, err := openhue.NewClientWithResponses("https://"+c.bridge, openhue.WithRequestEditorFn(p.Intercept))
	cobra.CheckErr(err)

	return client
}

// NewOpenHueClientNoAuth Creates a new NewClientWithResponses for a given server and no application key.
// This function will also skip SSL verification, as the Philips HUE Bridge exposes a self-signed certificate.
func NewOpenHueClientNoAuth(ip string) *openhue.ClientWithResponses {

	// skip SSL Verification
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	client, err := openhue.NewClientWithResponses("https://" + ip)
	cobra.CheckErr(err)

	return client
}
