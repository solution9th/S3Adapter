package cmd

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/solution9th/S3Adapter/app"
	"github.com/solution9th/S3Adapter/conf"

	"github.com/haozibi/zlog"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	// ConfigName default config name
	ConfigName = "osconfig"

	// ENVPrefix env prefix
	ENVPrefix = "OS"
)

var (
	cfgFile string

	rootCmd = &cobra.Command{
		Use:   app.BuildAppName,
		Short: "A AWS S3-compatible CLI",
	}
)

// Execute execute app
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ./"+ConfigName+".yml)")

	rootCmd.AddCommand(versionCMD)
	rootCmd.AddCommand(webCMD)
	rootCmd.AddCommand(configCMD)
}

func initConfig() {

	v := viper.New()
	v.AddConfigPath("conf/")
	v.SetConfigName("default")
	v.SetConfigType("yml")

	defaultBody, err := conf.Asset("conf/default.yml")
	if err != nil {
		zlog.ZError().Str("Config", "config/default.yml").Msg("[config] read default config error:" + err.Error())
		os.Exit(1)
	}

	err = v.ReadConfig(bytes.NewReader(defaultBody))
	if err != nil {
		zlog.ZError().Str("Config", "config/default.yml").Msg("[config] read default config error:" + err.Error())
		os.Exit(1)
	}

	defaultCfg := v.AllSettings()
	for k, v := range defaultCfg {
		viper.SetDefault(k, v)
	}

	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := homedir.Dir()
		if err != nil {
			zlog.ZError().Msg("[config] home dir error:" + err.Error())
			os.Exit(1)
		}
		// viper.AddConfigPath("conf/")
		viper.AddConfigPath(".")
		viper.AddConfigPath(home)
		// without extension
		viper.SetConfigName(ConfigName)
	}

	viper.SetEnvPrefix(ENVPrefix)
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.BindEnv("server.isdebug")
	viper.BindEnv("server.httpport")
	viper.BindEnv("server.pprofport")
	viper.BindEnv("server.region")
	viper.BindEnv("server.logpath")
	viper.BindEnv("server.endpoint")
	viper.BindEnv("mysql.port")
	viper.BindEnv("mysql.user")
	viper.BindEnv("mysql.passwd")
	viper.BindEnv("mysql.host")
	viper.BindEnv("mysql.dbname")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		// zlog.ZDebug().Str("Config", viper.ConfigFileUsed()).Msg("[config] Using config file:")
	}
}
