package config

import (
	"github.com/spf13/viper"
	"sync"
)

type Config struct {
	Website struct {
		Title    string `mapstructure:"title" json:"title"`
		Description   string `mapstructure:"description" json:"description"`
	} `mapstructure:"website"`

	HTTP struct {
		Addr        string        `mapstructure:"addr"`
		AccessLog bool `mapstructure:"access_log"`
		AcLogPath string        `mapstructure:"ac_log_path"`
	} `mapstructure:"http"`

	MySQL struct {
		Source       string `mapstructure:"source"`
		MaxIdleConns int    `mapstructure:"max_idle_conns"`
		LogMode bool `mapstructure:"log_mode"`
	} `mapstructure:"mysql"`

	Storage struct {
		VuePressBlogPath string `mapstructure:"vue_press_blog_path"`
	} `mapstructure:"storage"`

	Log struct {
		Level string `mapstructure:"level"`
		Path  string `mapstructure:"path"`
	} `mapstructure:"log"`

	Auth struct {
		Enable  bool   `mapstructure:"enable"`
		JwtKey  string `mapstructure:"jwt_key"`
	} `mapstructure:"auth"`

	Script struct{
		VisitorBuildScript string `mapstructure:"visitor_build_script"`
	} `mapstructure:"script"`
}

var c *Config
var o sync.Once

func Get() *Config {
	o.Do(func() {
		c = new(Config)
		err := viper.Unmarshal(c)
		if err != nil {
			panic(err) // config file corrupted, we need to restart
		}
	})
	return c
}
