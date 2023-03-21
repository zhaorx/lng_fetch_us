package main

import (
	"log"

	"github.com/robfig/cron/v3"
	"lng_fetch_us/config"
	_ "lng_fetch_us/config"
	_ "lng_fetch_us/database"
	"lng_fetch_us/job"
)

var cfg = config.Cfg

func main() {
	log.Println("current profile: " + cfg.Profile)

	registerJob() // 注册cron
	RunJob()      // 立即执行一次

	select {}
}

func registerJob() {
	if len(cfg.Cron) == 0 {
		panic("Cron表达式为空!")
	}

	c := newWithSeconds()
	_, err := c.AddFunc(cfg.Cron, func() {
		RunJob()
	})
	if err != nil {
		panic(err)
	}
	c.Start()
}

func RunJob() {
	job.FetchAndStore("inject", true)
	job.FetchAndStore("withdraw", true)
}

// 返回一个支持至 秒 级别的 cron
func newWithSeconds() *cron.Cron {
	secondParser := cron.NewParser(cron.Second | cron.Minute |
		cron.Hour | cron.Dom | cron.Month | cron.DowOptional | cron.Descriptor)
	return cron.New(cron.WithParser(secondParser), cron.WithChain())
}
