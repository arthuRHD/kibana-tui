package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

type AppConfig struct {
	FilePath   string `mapstructure:"file"`
	ESAddr     string `mapstructure:"es"`
	ESIndex    string `mapstructure:"es_index"`
	ESUser     string `mapstructure:"es_user"`
	ESPass     string `mapstructure:"es_pass"`
	ESAPIKey   string `mapstructure:"es_apikey"`
	ESCACert   string `mapstructure:"es_cacert"`
	ESInsecure         bool   `mapstructure:"es_insecure"`
	SampleSize         int    `mapstructure:"sample_size"`
	ESDefaultESQLQuery string `mapstructure:"es_default_esql_query"`
}

func (c AppConfig) SourceName() string {
	if c.ESAddr != "" {
		return fmt.Sprintf("ES:%s/%s", c.ESAddr, c.ESIndex)
	}
	if c.FilePath != "" {
		return c.FilePath
	}
	return fmt.Sprintf("sample (%d)", c.SampleSize)
}

func loadConfig() AppConfig {
	v := viper.New()

	v.SetDefault("es_index", "logs-*")
	v.SetDefault("sample_size", 500)

	v.SetConfigName("config")
	v.SetConfigType("yaml")

	xdgConfigHome := os.Getenv("XDG_CONFIG_HOME")
	if xdgConfigHome == "" {
		home, err := os.UserHomeDir()
		if err == nil {
			xdgConfigHome = filepath.Join(home, ".config")
		}
	}
	if xdgConfigHome != "" {
		v.AddConfigPath(filepath.Join(xdgConfigHome, "kibana-tui"))
	}
	v.AddConfigPath("/etc/kibana-tui")
	v.AddConfigPath(".")

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			fmt.Fprintf(os.Stderr, "Warning: error reading config: %v\n", err)
		}
	} else {
		fmt.Fprintf(os.Stderr, "Using config: %s\n", v.ConfigFileUsed())
	}

	v.SetEnvPrefix("KIBANA_TUI")
	v.BindEnv("file")
	v.BindEnv("es")
	v.BindEnv("es_index")
	v.BindEnv("es_user")
	v.BindEnv("es_pass")
	v.BindEnv("es_apikey")
	v.BindEnv("es_cacert")
	v.BindEnv("es_insecure")
	v.BindEnv("sample_size")
	v.BindEnv("es_default_esql_query")

	var cfg AppConfig
	if err := v.Unmarshal(&cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Error reading config: %v\n", err)
		os.Exit(1)
	}

	return cfg
}
