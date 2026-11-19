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
	Argocd     Argocd     `mapstructure:"argocd"`
	Monitoring Monitoring `mapstructure:"monitoring"`
	SLO        SLO        `mapstructure:"slo"`
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

// Argocd configures the GitOps engine and where it sources the repo.
type Argocd struct {
	Namespace string `mapstructure:"namespace"`
	Version   string `mapstructure:"version"`
	Repo      string `mapstructure:"repo"`
	Revision  string `mapstructure:"revision"`
}

// Monitoring configures the observability stack (kube-prometheus-stack).
type Monitoring struct {
	Namespace string  `mapstructure:"namespace"`
	Grafana   Grafana `mapstructure:"grafana"`
}

// Grafana identifies where the Grafana admin credentials live.
type Grafana struct {
	Secret string `mapstructure:"secret"`
}

// SLO configures the slo-generator integration (CronJob and Prometheus).
type SLO struct {
	Namespace string `mapstructure:"namespace"`
	CronJob   string `mapstructure:"cronjob"`
	Service   string `mapstructure:"service"`
	Port      int    `mapstructure:"port"`
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
	v.SetDefault("argocd.namespace", "argocd")
	v.SetDefault("argocd.version", "10.18.0")
	v.SetDefault("argocd.repo", "https://github.com/fabianoflorentino/srectl")
	v.SetDefault("argocd.revision", "main")
	v.SetDefault("monitoring.namespace", "monitoring")
	v.SetDefault("monitoring.grafana.secret", "kube-prometheus-stack-grafana")
	v.SetDefault("slo.namespace", "monitoring")
	v.SetDefault("slo.cronjob", "slo-generator")
	v.SetDefault("slo.service", "kube-prometheus-stack-prometheus")
	v.SetDefault("slo.port", 9090)
	v.SetDefault("registry", "localhost:5000")
}

// Load reads configuration from the given viper instance and unmarshals it into
// a Config. A missing config file is not an error: the CLI must still boot on
// defaults and environment variables alone.
func Load(v *viper.Viper) (*Config, error) {
	if err := v.ReadInConfig(); err != nil {
		if _, ok := errors.AsType[viper.ConfigFileNotFoundError](err); !ok {
			return nil, fmt.Errorf("reading config: %w", err)
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshaling config: %w", err)
	}
	return &cfg, nil
}
