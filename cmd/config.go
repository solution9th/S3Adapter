package cmd

import (
	"fmt"

	"github.com/solution9th/S3Adapter/internal/config"

	"github.com/davecgh/go-spew/spew"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var configCMD = &cobra.Command{
	Use:   "config",
	Short: "Just show config",
	Run: func(cmd *cobra.Command, args []string) {
		var p config.Config
		err := viper.Unmarshal(&p)
		if err != nil {
			panic(err)
		}

		s := viper.ConfigFileUsed()
		if s == "" {
			s = "default(path: conf/default.yml)"
		}

		fmt.Println("[Config]:", s)

		spew.Dump(p)
	},
}
