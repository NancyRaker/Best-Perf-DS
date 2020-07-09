/*
@Time : 2019/11/26 17:20 
@Author : yanKoo
@File : modle
@Software: GoLand
@Description:
*/
package model

type StateServerConfig struct {
	RedisConf   *Redis       `toml:"redis_config"`
	KafkaConfig *KafkaConfig `toml:"kafka"`
	Common      *Common      `toml:"common"`
}

// 公共配置
type Common struct {
	TimeTemplate string `toml:"time_layout"`
	LogLevel     string `toml:"log_level"`
	PPRofAddr    string `toml:"pprof_addr"` // pprof地址
	WorkerCount  int    `toml:"worker_count"`
}

// redis conf
type Redis struct {
	Host         string `toml:"host"`
	Port         int    `toml:"port"`
	Password     string `toml:"password"`
	Timeout      int    `toml:"timeout"`
	DB           int    `toml:"db"`
	MaxIdle      int    `toml:"max_idle"`
	IdleTimeout  int    `toml:"idle_timeout"`
	MaxActive    int    `toml:"max_active"`
	SentinelAddr string `toml:"sentinel_addr"`
}

// kafka配置
type KafkaConfig struct {
	Consumers []*ConsumerConfig `toml:"consumers"`
	Producers []*ProducerConfig `toml:"producers"`
}

// kafka消费者配置
type ConsumerConfig struct {
	Addr      string `toml:"consumer_addr"`
	Group     string `toml:"consumer_group"`
	Topic     string `toml:"consumer_topic"`
	Partition int32  `toml:"consumer_partition"`
}

// kafka生产者配置
type ProducerConfig struct {
	Addr  string `toml:"producer_addr"`
	Topic string `toml:"producer_topic"`
}

// 消息协议
type StateMessage struct {
	MsgType   int        `json:"msg_type"`   // 1 是更新设备状态消息
	Ip        string     `json:"ip"`         // 消息生产者的服务ip
	UserState *UserState `json:"user_state"` // 设备或者调度员状态消息
}

type UserState struct {
	StateCode    int          `json:"state_code"`    // 状态编号  值为1 代表更改设备是否进入休眠状态
	SleepingInfo SleepingInfo `json:"sleeping_info"` // 休眠状态对应的信息
}

type SleepingInfo struct {
	Id    int `json:"id"`    // 设备或者调度员id
	State int `json:"state"` // 1代表正常状态 2代表休眠状态
}
