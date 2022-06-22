package config

import (
	"fmt"
	"log"
	"os"
	"path"

	"github.com/spf13/viper"
	"github.com/strangelove-ventures/lens/client"
	"gopkg.in/yaml.v2"
)

//createConfig idempotently creates the config.
func CreateConfig(home string, debug bool) error {
	cfgPath := path.Join(home, "config.yaml")

	// If the config doesn't exist...
	if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
		// And the config folder doesn't exist...
		// And the home folder doesn't exist
		if _, err := os.Stat(home); os.IsNotExist(err) {
			// Create the home folder
			if err = os.Mkdir(home, os.ModePerm); err != nil {
				return err
			}
		}
	}

	// Then create the file...
	f, err := os.Create(cfgPath)
	if err != nil {
		return err
	}
	defer f.Close()

	// And write the default config to that location...
	if _, err = f.Write(defaultConfig(path.Join(home, "keys"), debug)); err != nil {
		return err
	}
	return nil
}

func OverwriteConfig(cfg *Config) error {
	home := viper.GetString("home")
	cfgPath := path.Join(home, "config.yaml")
	f, err := os.Create(cfgPath)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := f.Write(cfg.MustYAML()); err != nil {
		return err
	}

	log.Printf("updated lens configuration at %s", cfgPath)
	return nil
}

// Config represents the config file for the relayer
type Config struct {
	DefaultChain string                               `yaml:"default_chain" json:"default_chain"`
	Chains       map[string]*client.ChainClientConfig `yaml:"chains" json:"chains"`

	Cl map[string]*client.ChainClient
}

func (c *Config) GetDefaultClient() *client.ChainClient {
	return c.GetClient(c.DefaultChain)
}

func (c *Config) GetClient(chainID string) *client.ChainClient {
	if v, ok := c.Cl[chainID]; ok {
		return v
	}
	return nil
}

// Called to initialize the relayer.Chain types on Config
func ValidateConfig(c *Config) error {
	for _, chain := range c.Chains {
		if err := chain.Validate(); err != nil {
			return err
		}
	}
	if c.GetDefaultClient() == nil {
		return fmt.Errorf("default chain (%s) configuration not found", c.DefaultChain)
	}
	return nil
}

// MustYAML returns the yaml string representation of the Paths
func (c Config) MustYAML() []byte {
	out, err := yaml.Marshal(c)
	if err != nil {
		panic(err)
	}
	return out
}

func defaultConfig(keyHome string, debug bool) []byte {
	return Config{
		DefaultChain: "stride-testnet",
		Chains: map[string]*client.ChainClientConfig{
			"stride-testnet": client.GetStrideTestnetConfig(keyHome, debug),
			"gaia-testnet":   client.GetGaiaTestnetConfig(keyHome, debug),
		},
	}.MustYAML()
}
