package cmd

import (
	"blog/config"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"blog/service"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{Use: "blog"}

var log = logrus.StandardLogger()

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("rootCmd.Execute() failed, %s", err)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "./config.yaml", "config file path")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	rootCmd.AddCommand(&cobra.Command{
		Use:   "run",
		Short: "start running blog service",
		Run: func(cmd *cobra.Command, args []string) {
			svc := service.NewBackendService(config.Get())
			svc.Start()
			WaitStop()
			svc.Shutdown()
		},
	})
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	viper.SetConfigFile(cfgFile)
	viper.AutomaticEnv() // read in environment variables that match

	log.Formatter = &logrus.JSONFormatter{
		TimestampFormat:"2006-01-02 15:04:05",
	}
	log.Out = os.Stdout
	// if a config file is found, read it in.
	err := viper.ReadInConfig()

	if err == nil {
		log.Infof("using config file %s", viper.ConfigFileUsed())
	} else {
		log.Fatalf("viper.ReadInConfig() failed, %s", err)
	}

	level := viper.GetString("log.level")
	lvl, err := logrus.ParseLevel(level)
	if err != nil {
		log.Errorf("set log level failed, %s", err)
	}
	log.Infof("set log level %s", level)
	log.SetLevel(lvl)

	path := viper.GetString("log.path")
	if path == "" {
		path = "tmp/blog.log"
	}
	os.MkdirAll(filepath.Dir(path), 0777)
	log.Infof("set logfile path %s", path)

	// 设置 rotatelogs
	logWriter, err := rotatelogs.New(
		// 分割后的文件名称
		path + ".%Y%m%d.log",
		// 生成软链，指向最新日志文件
		rotatelogs.WithLinkName(path),
		// 设置最大保存时间(7天)
		rotatelogs.WithMaxAge(7*24*time.Hour),
		// 设置日志切割时间间隔(1天)
		rotatelogs.WithRotationTime(24*time.Hour),
	)
	if err != nil {
		log.Errorf("failed to set rotatelogs, %s", err)
	}
	log.SetOutput(logWriter)
}

func WaitStop() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	sig := <- sigs
	log.Infof("recv signal %s", sig)
}
