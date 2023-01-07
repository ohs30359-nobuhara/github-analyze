package excel

import (
	"github.com/xuri/excelize/v2"
	"strconv"
)

func Write(data [][]interface{}, fileName string, sheetName string) error {
	var f *excelize.File

	// 存在すれば追記、しなければ新規で作成
	f, e := excelize.OpenFile(fileName)
	if e != nil {
		f = excelize.NewFile()
	}

	index := f.NewSheet(sheetName)

	for rowCnt, rows := range data {
		for colCnt, col := range rows {
			axis, _ := excelize.ColumnNumberToName(colCnt + 1)
			if e := f.SetCellValue(sheetName, axis+strconv.Itoa(1+rowCnt), col); e != nil {
				return e
			}
		}
	}

	f.SetActiveSheet(index)
	if e := f.SaveAs(fileName); e != nil {
		return e
	}
	return nil
}

// WriteFromMap 連想mapをExcelに変換
func WriteFromMap(dataset map[string]map[string]int, fileName string, sheetName string) error {
	var (
		cols   [][]interface{}
		body   [][]interface{}
		header []interface{}
	)

	// headerの最初のindexは空
	header = append(header, "#")
	// 行生成に必要なデータ加工用のmapを定義
	type uniqKey struct {
		key1 string
		key2 string
	}

	xParams := make(map[uniqKey]int)
	yIndex := make(map[string]int)

	for firstKey, firstItem := range dataset {
		for secondKey, secondItem := range firstItem {
			// Y軸のindex作成 (値は使わないので適当な値を差し込む)
			if _, ok := yIndex[secondKey]; !ok {
				yIndex[secondKey] = 0
			}
			xParams[uniqKey{key1: firstKey, key2: secondKey}] = secondItem
		}
		header = append(header, firstKey)
	}

	for index, _ := range yIndex {
		var col []interface{}
		col = append(col, index)

		for _, head := range header {
			if head == "#" {
				continue
			}

			cnt, ok := xParams[uniqKey{key1: head.(string), key2: index}]
			if ok {
				col = append(col, cnt)
			} else {
				col = append(col, 0)
			}
		}
		body = append(body, col)
	}

	cols = append(cols, header)
	cols = append(cols, body...)

	return Write(cols, fileName, sheetName)
}
