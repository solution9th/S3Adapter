package cmd

import (
	"os"
	"reflect"

	"github.com/solution9th/S3Adapter/app"
	"github.com/solution9th/S3Adapter/internal/config"

	"github.com/davecgh/go-spew/spew"
	"github.com/haozibi/zlog"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var httpPort, pprofPort, endpoint, logpath, region string
var isdebug, show bool

var mysqlUser, mysqlPasswd, mysqlHost, mysqlDBName string
var mysqlPort int

func init() {

	webCMD.Flags().BoolVarP(&show, "show", "", false, "show config")

	webCMD.Flags().BoolVarP(&isdebug, "debug", "", false, "is debug")
	viper.BindPFlag("server.isdebug", webCMD.Flags().Lookup("debug"))

	webCMD.Flags().StringVarP(&httpPort, "httpport", "", "9091", "http port")
	viper.BindPFlag("server.httpport", webCMD.Flags().Lookup("httpport"))

	webCMD.Flags().StringVarP(&pprofPort, "pprofport", "", "9092", "pprof port")
	viper.BindPFlag("server.pprofport", webCMD.Flags().Lookup("pprofport"))

	webCMD.Flags().StringVarP(&endpoint, "endpoint", "", "", "http end point")
	viper.BindPFlag("server.endpoint", webCMD.Flags().Lookup("endpoint"))

	webCMD.Flags().StringVarP(&logpath, "logpath", "", "", "log file path")
	viper.BindPFlag("server.logpath", webCMD.Flags().Lookup("logpath"))

	webCMD.Flags().StringVarP(&region, "region", "", "", "os service region")
	viper.BindPFlag("server.region", webCMD.Flags().Lookup("region"))

	webCMD.Flags().IntVarP(&mysqlPort, "mysqlport", "", 3306, "mysql port")
	viper.BindPFlag("mysql.port", webCMD.Flags().Lookup("mysqlport"))

	webCMD.Flags().StringVarP(&mysqlUser, "mysqluser", "", "root", "mysql username")
	viper.BindPFlag("mysql.user", webCMD.Flags().Lookup("mysqluser"))

	webCMD.Flags().StringVarP(&mysqlPasswd, "mysqlpassword", "", "", "mysql password")
	viper.BindPFlag("mysql.passwd", webCMD.Flags().Lookup("mysqlpassword"))

	webCMD.Flags().StringVarP(&mysqlHost, "mysqlhost", "", "127.0.0.1", "mysql host")
	viper.BindPFlag("mysql.host", webCMD.Flags().Lookup("mysqlhost"))

	webCMD.Flags().StringVarP(&mysqlDBName, "mysqldbname", "", "", "mysql databases")
	viper.BindPFlag("mysql.dbname", webCMD.Flags().Lookup("mysqldbname"))

}

var webCMD = &cobra.Command{
	Use:   "web",
	Short: "Start Web Server",
	Run: func(cmd *cobra.Command, args []string) {

		var p config.Config
		err := viper.Unmarshal(&p)
		if err != nil {
			panic(err)
		}

		if show {
			spew.Dump(p)
		}

		if p.Server.LogPath == "" {
			p.Server.LogPath = "-"
		}

		if v := isNil(p.Server); v != "" {
			zlog.ZError().Str("Field", v).Msg("[config] value is nil")
			os.Exit(1)
		}

		if v := isNil(p.MySQL); v != "" {
			zlog.ZError().Str("Field", v).Msg("[config] value is nil")
			os.Exit(1)
		}

		app.Run(p)
	},
}

func isNil(p interface{}) string {
	v := reflect.ValueOf(p)
	t := reflect.TypeOf(p)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
		t = t.Elem()
	}

	for i := 0; i < v.NumField(); i++ {
		value := v.Field(i)
		if value.Kind() == reflect.Bool {
			continue
		}

		if reflect.DeepEqual(value.Interface(),
			reflect.Zero(value.Type()).Interface()) {
			return t.Field(i).Name
		}
	}
	return ""
}
