package job

import (
	"errors"
	"fmt"
	"log"
	"os"
	"reflect"
	"time"

	"github.com/shakinm/xlsReader/xls"
	"github.com/spf13/cast"
	"lng_fetch_us/config"
	"lng_fetch_us/store"
	"lng_fetch_us/util"
)

var sheetName = "Data 1"
var cfg = config.Cfg

// var dataMap = map[string]map[int]string{
// 	"inject": { // 注气
// 		1:  "N5050US2",              // B2
// 		3:  "NGM_EPG0_SAI_R48_MMCF", // D2
// 		38: "N5050832",              // AM2
// 		39: "N5050842",              // AN2
// 		40: "N5050852",              // AO2
// 		41: "N5050862",              // AP2
// 		42: "N5050912",              // AQ2
// 	},
// 	"withdraw": { // 采气
// 		1:  "N5060US2",              // B2
// 		3:  "NGM_EPG0_SAW_R48_MMCF", // D2
// 		37: "N5060832",              // AL2
// 		38: "N5060842",              // AM2
// 		39: "N5060852",              // AN2
// 		40: "N5060862",              // AO2
// 		41: "N5060912",              // AP2
// 	},
// }
/*
TOTAL:    cast.ToInt64(cols[1].GetString()),
LOWER48:  cast.ToInt64(cols[3].GetString()),
EAST:     cast.ToInt64(cols[38].GetString()),
SOUTH:    cast.ToInt64(cols[39].GetString()),
MIDWEST:  cast.ToInt64(cols[40].GetString()),
MOUNTAIN: cast.ToInt64(cols[41].GetString()),
PACIFIC:  cast.ToInt64(cols[42].GetString()),

*/
var dataMap = map[string][]struct {
	Name      string
	Index     int
	Sourcekey string
}{
	"inject": {
		{Name: "TOTAL", Index: 1, Sourcekey: "N5050US2"},
		{Name: "LOWER48", Index: 3, Sourcekey: "NGM_EPG0_SAI_R48_MMCF"},
		{Name: "EAST", Index: 38, Sourcekey: "N5050832"},
		{Name: "SOUTH", Index: 39, Sourcekey: "N5050842"},
		{Name: "MIDWEST", Index: 40, Sourcekey: "N5050852"},
		{Name: "MOUNTAIN", Index: 41, Sourcekey: "N5050862"},
		{Name: "PACIFIC", Index: 42, Sourcekey: "N5050912"},
	},
	"withdraw": {
		{Name: "TOTAL", Index: 1, Sourcekey: "N5060US2"},
		{Name: "LOWER48", Index: 3, Sourcekey: "NGM_EPG0_SAW_R48_MMCF"},
		{Name: "EAST", Index: 37, Sourcekey: "N5060832"},
		{Name: "SOUTH", Index: 38, Sourcekey: "N5060842"},
		{Name: "MIDWEST", Index: 39, Sourcekey: "N5060852"},
		{Name: "MOUNTAIN", Index: 40, Sourcekey: "N5060862"},
		{Name: "PACIFIC", Index: 41, Sourcekey: "N5060912"},
	},
}

func FetchAndStore(key string, retryflag bool) (ok bool) {
	log.Printf("Fetch %s data...\n", key)

	file, err := util.GetDataFile(key)
	if err != nil {
		log.Printf("GetDataFile error: %s\n", err.Error())
		if retryflag { // 如果需要retry 就开一个retry协程
			go RetryFetchAndStore(key)
		}
		return false
	}
	defer file.Close()

	// 解析excel数据 组装[]Data
	datas, err := parseFile(file, key)
	if err != nil {
		log.Printf("parseFile error: %s\n", err.Error())
		return false
	}

	// insert table
	err = store.StoreDatas(datas, key)
	if err != nil {
		log.Printf("StoreDatas error: %s\n", err.Error())
		return false
	}

	log.Printf("Fetch %s data done!\n", key)
	return true
}

// RetryFetchAndStore 定期retry重试
func RetryFetchAndStore(key string) {
	log.Printf(" 重新获取%s数据开始...\n", key)

	ticker := time.Tick(time.Duration(cfg.RetryBeat) * time.Hour)
	count := cfg.RetryLimit
	for _ = range ticker {
		if count <= 0 {
			log.Println(" 重新获取%s数据结束，达到%d次上限\n", key, cfg.RetryLimit)
			return
		}

		if ok := FetchAndStore(key, false); ok { // 成功一次就结束携程
			return
		}
	}
}

// 解析excel数据 组装[]Data
func parseFile(file *os.File, key string) ([]*store.Data, error) {
	// OpenFile打开成功返回的是workBook对象
	wb, err := xls.OpenFile(file.Name())
	if err != nil {
		return nil, err
	}

	// 获得第二张工作表
	sheet, err := wb.GetSheet(1)
	if err != nil {
		return nil, err
	}

	err = ValidateColumn(sheet, key)
	if err != nil {
		return nil, err
	}

	// 获得工作表的行数
	rn := sheet.GetNumberRows()
	// 循环处理每一行
	datas := make([]*store.Data, 0)
	for i := 483; i < rn-1; i++ { // 从2013-01(484行)开始
		// 获取行
		if row, e := sheet.GetRow(i); e == nil && row != nil {
			// // 获取这行的所有列
			cols := row.GetCols()
			if cols == nil || len(cols) < 1 {
				continue
			}

			// 获得列数
			// colen := len(cols)
			// fmt.Println("colen", colen, i)
			// fmt.Println(cols[3].GetInt64(), cols[3].GetString())

			// 组装Data
			loc, _ := time.LoadLocation("Local") // 获取当前时区
			origin_rq, _ := time.ParseInLocation("2006-01-02", "1900-01-01", loc)
			rq := origin_rq.AddDate(0, 0, cast.ToInt(cols[0].GetString())-1)
			rq = rq.AddDate(0, 0, -rq.Day()+1) // 取当月第一天

			d := &store.Data{
				RQ:   rq,
				TYPE: key,
			}

			list := dataMap[key]
			for _, el := range list {
				// 利用反射 从dataMap中找字段name 再赋值
				fieldName := el.Name
				i := el.Index
				v := reflect.ValueOf(d)
				v = v.Elem()
				f := v.FieldByName(fieldName)
				if f.Kind() == reflect.Int64 {
					f.SetInt(cast.ToInt64(cols[i].GetString()))
				}
			}

			datas = append(datas, d)
		}
	}

	return datas, nil

}

// ValidateColumn 验证excel的列是否改变
func ValidateColumn(sheet *xls.Sheet, key string) (err error) {
	if sheet.GetName() != sheetName {
		return errors.New(fmt.Sprintln("excel 文件结构改变: sheet Sourcekey ", sheet.GetName(), "!= ", sheetName))
	}

	row, err := sheet.GetRow(1)
	if err != nil {
		return err
	}

	list := dataMap[key]
	for _, el := range list {
		cell, err := row.GetCol(el.Index)
		if err != nil {
			return err
		}

		if cell.GetString() != el.Sourcekey {
			return errors.New(fmt.Sprintln("excel 文件列结构改变: B", el.Index, " cell value is not ", el.Sourcekey))
		}
	}

	return nil
}
