package store

import (
	"time"

	"github.com/godror/godror"
	_ "github.com/godror/godror"
	"lng_fetch_us/database"
	"lng_fetch_us/util"
)

var db = database.DB

func StoreDatas(datas []*Data, key string) error {

	for _, data := range datas {
		err := storeSingle(data, key)
		if err != nil {
			return err
		}
	}

	return nil
}

func storeSingle(data *Data, key string) error {
	// 判断是否重复
	sql := "SELECT * FROM INT_LNG_US WHERE rq = ? AND type = ?"
	countErr := db.QueryRow(sql, data.RQ, key, godror.FetchArraySize(1))
	if countErr != nil { // 有重复数据 删除旧数据
		_, err := db.Exec("DELETE FROM INT_LNG_US WHERE rq = :1 AND type = :2", data.RQ, key)
		if err != nil {
			return err
		}
	}

	// 单位转换
	total := util.MMCFToWM3(data.TOTAL)
	lower48 := util.MMCFToWM3(data.LOWER48)
	east := util.MMCFToWM3(data.EAST)
	south := util.MMCFToWM3(data.SOUTH)
	midwest := util.MMCFToWM3(data.MIDWEST)
	mountain := util.MMCFToWM3(data.MOUNTAIN)
	pacific := util.MMCFToWM3(data.PACIFIC)

	_, err := db.Exec(`INSERT INTO INT_LNG_US(
                       rq, 
                       TYPE,
                       TOTAL,
                       LOWER48,
                       EAST,
                       SOUTH,
                       MIDWEST,
                       MOUNTAIN,
                       PACIFIC
                   ) VALUES (:1,:2,:3,:4,:5,:6,:7,:8,:9)`,
		data.RQ,
		key,
		total,
		lower48,
		east,
		south,
		midwest,
		mountain,
		pacific)
	if err != nil {
		return err
	}
	return err
}

type Data struct {
	RQ       time.Time `json:"rq"`
	TYPE     string    `json:"type"`
	TOTAL    int64     `json:"total"`
	LOWER48  int64     `json:"lower48"`
	EAST     int64     `json:"east"`
	SOUTH    int64     `json:"south"`
	MIDWEST  int64     `json:"midwest"`
	MOUNTAIN int64     `json:"mountain"`
	PACIFIC  int64     `json:"pacific"`
}
