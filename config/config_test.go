package config

import (
	"github.com/spf13/viper"
	"testing"
)

func TestConfig(t *testing.T) {
	cfgFile := "../config_example.yaml"
	viper.SetConfigFile(cfgFile)
	viper.AutomaticEnv() // read in environment variables that match

	// if a config file is found, read it in.
	err := viper.ReadInConfig()
	if err == nil {
		t.Log("using config file", "cfg", viper.ConfigFileUsed())
	} else {
		t.Error("viper.ReadInConfig()", "error", err)
		t.FailNow()
	}
	cfg := Get()
	t.Logf("config unmarshal result: %v", cfg)
}
