/**
 * 公共配置类，用于加载数据库配置等
 * Author: tesion
 * Date: 20th March 2019
 * Note:
 *		redis客户端配置受redis服务器配置影响
 */
package conf

import (
	"fmt"
	"github.com/go-ini/ini"
)

const (
	DEFAULT_DB_PORT            = 3306
	DEFAULT_DB_MAX_CONN        = 10
	DEFAULT_REDIS_PORT         = 6379
	DEFAULT_REDIS_TIMEOUT      = 10
	DEFAULT_REDIS_DB           = 0
	DEFAULT_REDIS_MAX_IDLE     = 100
	DEFAULT_REDIS_IDLE_TIMEOUT = 0
	DEFAULT_REDIS_MAX_ACTIVE   = 500
)

type DBConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DB       string
	MaxConn  int
	Driver   string
}

type RedisConfig struct {
	Host         string
	Port         int
	Password     string
	Timeout      int
	DB           int
	MaxIdle      int
	IdleTimeout  int
	MaxActive    int
	SentinelAddr string
}

func NewDBConfig() *DBConfig {
	return new(DBConfig)
}

func (cfg *DBConfig) LoadConfig(section, configPath string) error {
	if cfg == nil {
		return fmt.Errorf("config obj is null")
	}

	config, err := ini.Load(configPath)
	if err != nil {
		return err
	}

	sec := config.Section(section)
	if sec == nil {
		return fmt.Errorf("section(%s) not exist", section)
	}

	cfg.Host = sec.Key("host").String()
	cfg.Port = sec.Key("port").MustInt(DEFAULT_DB_PORT)
	cfg.User = sec.Key("user").String()
	cfg.Password = sec.Key("password").String()
	cfg.DB = sec.Key("db").String()
	cfg.MaxConn = sec.Key("max_conn").MustInt(DEFAULT_DB_MAX_CONN)
	cfg.Driver = sec.Key("driver").String()

	return nil
}

func (cfg *DBConfig) GetDSN() string {
	str := ""
	if cfg != nil {
		str = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DB)
	}
	return str
}

func (cfg *DBConfig) GetDriver() string {
	if cfg != nil {
		return cfg.Driver
	}

	return ""
}

func NewRedisConfig() *RedisConfig {
	return new(RedisConfig)
}

func (cfg *RedisConfig) LoadConfig(section, path string) error {
	config, err := ini.Load(path)
	if err != nil {
		return fmt.Errorf("load config(%s) error: %s", path, err)
	}

	sec := config.Section(section)
	if sec == nil {
		return fmt.Errorf("section(%s) not exist", section)
	}

	cfg.Host = sec.Key("host").String()
	cfg.Port = sec.Key("port").MustInt(DEFAULT_REDIS_PORT)
	cfg.Password = sec.Key("password").String()
	cfg.SentinelAddr = sec.Key("sentinel_addr").String()

	cfg.Timeout = DEFAULT_REDIS_TIMEOUT
	key := sec.Key("timeout")
	if key != nil {
		cfg.Timeout = key.MustInt(DEFAULT_REDIS_TIMEOUT)
	}

	cfg.DB = DEFAULT_REDIS_DB
	key = sec.Key("db")
	if key != nil {
		cfg.DB = key.MustInt(DEFAULT_REDIS_DB)
	}

	cfg.MaxIdle = DEFAULT_REDIS_MAX_IDLE
	key = sec.Key("max_idle")
	if key != nil {
		cfg.MaxIdle = key.MustInt(DEFAULT_REDIS_MAX_IDLE)
	}

	cfg.IdleTimeout = DEFAULT_REDIS_IDLE_TIMEOUT
	key = sec.Key("idle_timeout")
	if key != nil {
		cfg.IdleTimeout = key.MustInt(DEFAULT_REDIS_IDLE_TIMEOUT)
	}

	cfg.MaxActive = DEFAULT_REDIS_MAX_ACTIVE
	key = sec.Key("max_active")
	if key != nil {
		cfg.MaxActive = key.MustInt(DEFAULT_REDIS_MAX_ACTIVE)
	}

	return nil
}
