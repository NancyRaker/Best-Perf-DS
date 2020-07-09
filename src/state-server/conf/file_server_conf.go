/*
@Time : 2019/4/20 17:11
@Author : yanKoo
@File : conf
@Software: GoLand
@Description:
*/
package conf

import (
	"flag"
	"fmt"
	"github.com/BurntSushi/toml"
	"os"
	"state-server/model"
)


const (
	DEFAULT_CONFIG = "state-server-conf.toml"
)

var Config *model.StateServerConfig

func init() {
	filePath := flag.String("config", DEFAULT_CONFIG, "web api configuration file path")
	flag.Parse()

	Config = &model.StateServerConfig{}
	_, err := toml.DecodeFile(*filePath, &Config) // 编译之后的执行文件所在位置的相对位置
	if err != nil {
		fmt.Printf("Fail to read file: %v", err)
		os.Exit(1)
	}

	fmt.Printf(" common config: %+v\n", Config.Common)
	fmt.Printf("  redis config: %+v\n", Config.RedisConf)
	for _, consumer := range Config.KafkaConfig.Consumers {
		fmt.Printf("kafka consumer: %+v\n", consumer)
	}

	for _, producer := range Config.KafkaConfig.Producers {
		fmt.Printf("kafka producer: %+v\n", producer)
	}
}
