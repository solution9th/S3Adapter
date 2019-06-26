package mysql

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/solution9th/S3Adapter/internal/db"

	_ "github.com/go-sql-driver/mysql"
	"github.com/haozibi/gendry/builder"
	"github.com/haozibi/gendry/manager"
	"github.com/haozibi/gendry/scanner"
	"github.com/haozibi/zlog"
)

var (
	ErrMissParams    = errors.New("miss params")
	ErrParamsTypeErr = errors.New("params type error")
)

func init() {
	scanner.SetTagName("json")
}

// MySQLFunc database operation
type MySQLFunc struct {
	tableNameInfo string
	client        *sql.DB
}

var defaultDB *sql.DB

// NewDB new MySQLFunc
func NewDB(table string) db.DB {
	return &MySQLFunc{
		tableNameInfo: table,
		client:        defaultDB,
	}
}

// LinkDB test mysql link
func (d *MySQLFunc) LinkDB(config map[string]interface{}) error {

	dbname, ok := config["dbname"].(string)
	if !ok {
		zlog.ZError().Str("Params", "dbname").Interface("Value", config["dbname"]).Msg("[MySQL] param type error")
		return ErrParamsTypeErr
	}
	user, ok := config["user"].(string)
	if !ok {
		zlog.ZError().Str("Params", "user").Interface("Value", config["user"]).Msg("[MySQL] param type error")
		return ErrParamsTypeErr
	}
	passwd, ok := config["password"].(string)
	if !ok {
		zlog.ZError().Str("Params", "password").Interface("Value", config["password"]).Msg("[MySQL] param type error")
		return ErrParamsTypeErr
	}
	host, ok := config["host"].(string)
	if !ok {
		zlog.ZError().Str("Params", "host").Interface("Value", config["host"]).Msg("[MySQL] param type error")
		return ErrParamsTypeErr
	}
	port, ok := config["port"].(int)
	if !ok {
		zlog.ZError().Str("Params", "port").Interface("Value", config["port"]).Msg("[MySQL] param type error")
		return ErrParamsTypeErr
	}

	return d.link(dbname, user, passwd, host, port)
}

func (d *MySQLFunc) link(dbName, user, passwd, host string, port int) (err error) {

	if defaultDB != nil {
		d.client = defaultDB
		return d.client.Ping()
	}
	d.client, err = manager.New(dbName, user, passwd, host).Set(
		manager.SetCharset("utf8mb4"),
		manager.SetAllowCleartextPasswords(true),
		manager.SetInterpolateParams(true),
		manager.SetParseTime(true),
		manager.SetTimeout(1*time.Second),
		manager.SetReadTimeout(1*time.Second)).Port(port).Open(true)
	if err != nil {
		zlog.ZError().Str("DBName", dbName).Str("User", user).Str("Host", host).Int("Port", port).Msg("[MySQL] error: " + err.Error())
		return err
	}

	err = d.client.Ping()
	if err != nil {
		return err
	}

	zlog.ZDebug().Str("DBName", dbName).Str("User", user).Str("Host", host).Int("Port", port).Msg("[MySQL] link")

	return nil
}

func (d *MySQLFunc) query(tableName string, where map[string]interface{}, ptr interface{}) error {

	// if d.db == nil {
	// 	if err := d.link(); err != nil {
	// 		return err
	// 	}
	// }

	if reflect.TypeOf(ptr).Kind() != reflect.Ptr {
		return fmt.Errorf("params error: query need ptr")
	}

	cond, vals, err := builder.BuildSelect(tableName, where, nil)
	if err != nil {
		return err
	}

	rows, err := d.client.Query(cond, vals...)
	if err != nil {
		return err
	}

	err = scanner.ScanClose(rows, ptr)
	return err
}

func (d *MySQLFunc) count(cond string, val ...interface{}) (int, error) {

	// if d.db == nil {
	// 	if err := d.link(); err != nil {
	// 		return 0, err
	// 	}
	// }

	rows, err := d.client.Query(cond, val...)
	if err != nil {
		return 0, err
	}

	tmp := struct {
		Count int `json:"count"`
	}{}

	err = scanner.ScanClose(rows, &tmp)

	return tmp.Count, err
}

func (d *MySQLFunc) save(tableName string, data map[string]interface{}) (id int, err error) {

	// if d.db == nil {
	// 	if err := d.link(); err != nil {
	// 		return 0, err
	// 	}
	// }

	var datas []map[string]interface{}
	datas = append(datas, data)

	cond, vals, err := builder.BuildInsert(tableName, datas)
	r, err := d.client.Exec(cond, vals...)
	if err != nil {
		return 0, err
	}

	ids, err := r.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(ids), nil
}
