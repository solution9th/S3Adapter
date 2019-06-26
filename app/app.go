package app

import (
	"net/http"
	_ "net/http/pprof"
	"os"

	"github.com/solution9th/S3Adapter/internal/config"
	"github.com/solution9th/S3Adapter/internal/db"
	"github.com/solution9th/S3Adapter/internal/db/mysql"

	"github.com/gorilla/mux"
	"github.com/haozibi/zlog"
)

var (
	// GlobalRegion service Region
	GlobalRegion = ""
	// EndPointDomain endpoint
	EndPointDomain = ""
)

// Run start http
func Run(cfg config.Config) error {

	EndPointDomain = cfg.Server.EndPoint
	GlobalRegion = cfg.Server.Region
	pprofPort := cfg.Server.PprofPort

	a, err := NewAPP(cfg)
	if err != nil {
		zlog.ZError().Msg("[Init] error:" + err.Error())
		return err
	}

	if pprofPort != "" {
		go func() {
			zlog.ZDebug().Str("pprof", pprofPort).Msg("[pprof]")
			http.ListenAndServe(":"+pprofPort, nil)
		}()
	}

	httpPort := cfg.Server.HTTPPort

	r := mux.NewRouter()
	NewAPIRouter(r, a)

	zlog.ZInfo().Str("Port", httpPort).Msg("listen...")
	err = http.ListenAndServe(":"+httpPort, r)
	if err != nil {
		zlog.ZFatal().Msg(err.Error())
		return err
	}
	return nil
}

type API struct {
	DB db.DB
}

// NewAPP 初始化 APP
var NewAPP = func(cfg config.Config) (*API, error) {

	if cfg.Server.LogPath != "-" {
		filename := cfg.Server.LogPath
		f, err := os.OpenFile(filename, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
		if err != nil {
			zlog.ZError().Msg(err.Error())
			return nil, err
		}
		zlog.ZDebug().Str("LogPath", filename).Msg("[Log]")
		zlog.NewBasicLog(f)
	}

	a := &API{}

	dbType := "mysql"

	apiConfig := make(map[string]interface{})

	switch dbType {
	case "mysql":
		apiConfig = map[string]interface{}{
			"dbname":   cfg.MySQL.DBName,
			"user":     cfg.MySQL.User,
			"password": cfg.MySQL.Passwd,
			"host":     cfg.MySQL.Host,
			"port":     cfg.MySQL.Port,
		}
	}

	tableName := "info"

	a.DB = mysql.NewDB(tableName)
	err := a.DB.LinkDB(apiConfig)
	if err != nil {
		return nil, err
	}

	err = a.DB.AddTable()
	if err != nil {
		return nil, err
	}

	return a, nil
}
