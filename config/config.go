package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

var Cfg Config

const configPath = "./config.yml"

// 加载log
func init() {
	// load cfg
	data, err := os.ReadFile(configPath)
	if err != nil {
		panic(err.Error())
	}

	Cfg = NewConfigWithDefault()
	err = yaml.Unmarshal(data, &Cfg)
	if err != nil {
		panic(err)
	}
	fmt.Printf("cfg db:\n%v\n\n", Cfg)

}

// cfg 缺省设置
func NewConfigWithDefault() Config {
	c := Config{
		Profile:      "dev",
		InjectionUrl: "",
		WithdrawUrl:  "",
		Proxy:        "",
		RetryLimit:   12,
		RetryBeat:    20,
	}
	return c
}

type Config struct {
	Profile string `yaml:"profile"` // 执行环境 dev/prod/history/org

	// 公共配置
	Proxy      string `yaml:"proxy"`      // 生产环境代理
	RetryLimit int    `yaml:"retryLimit"` // 重新抓取次数上限
	RetryBeat  int    `yaml:"retryBeat"`  // 重新抓取周期小时数

	InjectionUrl string `yaml:"injectionUrl"` // 注气file地址
	WithdrawUrl  string `yaml:"withdrawUrl"`  // 采气file地址

	// oracle数据库配置
	Host        string `yaml:"host"`
	Port        int    `yaml:"port"`
	Username    string `yaml:"username"`
	Password    string `yaml:"password"`
	ServiceName string `yaml:"serviceName"`
}
