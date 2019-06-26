package cmd

import (
	"fmt"
	"runtime"

	"github.com/solution9th/S3Adapter/app"

	"github.com/spf13/cobra"
)

var versionCMD = &cobra.Command{
	Use:   "version",
	Short: "Show version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf(`%s:
  version     : %s
  build date  : %s
  go version  : %s
  go compiler : %s
  platform    : %s/%s
`, app.BuildAppName, app.BuildVersion, app.BuildTime,
			runtime.Version(), runtime.Compiler, runtime.GOOS, runtime.GOARCH)
	},
}
