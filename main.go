package main

import (
	"os"

	"github.com/haozibi/zlog"
	"github.com/solution9th/S3Adapter/cmd"
)

func init() {
	zlog.NewBasicLog(os.Stdout)
}

func main() {

	cmd.Execute()
}
