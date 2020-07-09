package db

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"web-ips/conf"
)

const (
	DEFAULT_DB_CONFIG = "conf.ini"
)

// db 包装类
type SDBManager struct {
	DB    *sql.DB
	DBCfg conf.DBConfig
}

type SDBContainer struct {
	dbc map[string]*SDBManager
}

var (
	DBC       SDBContainer
	DBHandler *sql.DB
	Err       error
)

func init() {
	if DBHandler != nil {
		return
	}

	DBHandler, Err = CreateDBHandler("db", DEFAULT_DB_CONFIG)
}

// simple version
func CreateDBHandler(section, path string) (*sql.DB, error) {
	cfg := conf.NewDBConfig()
	if err := cfg.LoadConfig(section, path); err != nil {
		return nil, fmt.Errorf("load config error(%s)", err)
	}

	db, err := sql.Open(cfg.GetDriver(), cfg.GetDSN())
	if err != nil {
		return nil, err
	}

	db.SetMaxIdleConns(cfg.MaxConn)
	return db, nil
}

// db container interface
func (Dbc *SDBContainer) Set(name string, dbm *SDBManager) error {
	if dbm == nil {
		return fmt.Errorf("db manager is nil")
	}

	if _, ok := Dbc.dbc[name]; ok {
		return fmt.Errorf("db manager name exists")
	}

	Dbc.dbc[name] = dbm
	return nil
}

func (Dbc *SDBContainer) Get(name string) (*sql.DB, error) {
	if 0 == len(Dbc.dbc) {
		return nil, fmt.Errorf("empty db container")
	}

	if _, ok := Dbc.dbc[name]; !ok {
		return nil, fmt.Errorf("no db manager(%s) exist", name)
	}

	return Dbc.dbc[name].DB, nil
}
