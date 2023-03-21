package database

import (
	"database/sql"
	"fmt"

	_ "github.com/godror/godror"
	"lng_fetch_us/config"
)

var cfg = config.Cfg
var DB *sql.DB

func init() {
	dsn := fmt.Sprintf(`user="%s" password="%s" connectString="%s:%d/%s"`, cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.ServiceName)
	var err error
	DB, err = sql.Open("godror", dsn)
	if err != nil {
		panic(err)
	}
	// defer DB.Close()
	err = DB.Ping()
	if err != nil {
		fmt.Println(err.Error())
		panic(err)
	}

	DB.SetMaxOpenConns(1)
	DB.SetMaxIdleConns(0)
	fmt.Println("db链接成功")
}
