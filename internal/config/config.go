// Package config loads srectl configuration from config/srectl.yaml with
// SRECTL_* environment variable overrides.
package config

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// EnvPrefix namespaces every environment override (e.g. SRECTL_CLUSTER_NAME).
const EnvPrefix = "SRECTL"

// Config is the root srectl configuration surface.
type Config struct {
	Cluster    Cluster    `mapstructure:"cluster"`
	Components Components `mapstructure:"components"`
	Registry   string     `mapstructure:"registry"`
}

// Cluster describes the kind cluster topology.
type Cluster struct {
	Name  string `mapstructure:"name"`
	Nodes int    `mapstructure:"nodes"`
}

// Components toggles which parts of the platform the CLI manages.
type Components struct {
	Argocd        bool `mapstructure:"argocd"`
	Observability bool `mapstructure:"observability"`
	Security      bool `mapstructure:"security"`
	Chaos         bool `mapstructure:"chaos"`
	Apps          bool `mapstructure:"apps"`
	Projects      bool `mapstructure:"projects"`
}

// NewViper returns a viper instance configured for srectl: defaults, a YAML
// config file (optional) and SRECTL_* environment overrides.
func NewViper() *viper.Viper {
	v := viper.New()
	v.SetConfigName("srectl")
	v.SetConfigType("yaml")
	v.AddConfigPath("config")
	v.AddConfigPath(".")

	v.SetEnvPrefix(EnvPrefix)
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	setDefaults(v)
	return v
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("cluster.name", "srectl")
	v.SetDefault("cluster.nodes", 3)
	v.SetDefault("components.argocd", true)
	v.SetDefault("components.observability", true)
	v.SetDefault("components.security", true)
	v.SetDefault("components.chaos", true)
	v.SetDefault("components.apps", true)
	v.SetDefault("components.projects", true)
	v.SetDefault("registry", "localhost:5000")
}

// Load reads configuration from the given viper instance and unmarshals it into
// a Config. A missing config file is not an error: the CLI must still boot on
// defaults and environment variables alone.
func Load(v *viper.Viper) (*Config, error) {
	if err := v.ReadInConfig(); err != nil {
		var notFound viper.ConfigFileNotFoundError
		if !errors.As(err, &notFound) {
			return nil, fmt.Errorf("reading config: %w", err)
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshaling config: %w", err)
	}
	return &cfg, nil
}
