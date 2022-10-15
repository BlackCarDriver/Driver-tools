package stock

import (
	"encoding/json"
	"fmt"
	"github.com/BlackCarDriver/GoProject-api/common/util"
	"io/ioutil"
	"net/http"
	"time"
)

const reqPaht = "https://push2.eastmoney.com/api/qt/ulist.np/get"

type utilRespLayer1 struct {
	Data utilRespLayer2 `json:"data"`
}

type utilRespLayer2 struct {
	Total int            `json:"total"`
	Diff  []utilRespData `json:"diff"`
}

type utilRespData struct {
	F2   float64 `json:"f2"`   // 现价
	F3   float64 `json:"f3"`   // 增幅
	F4   float64 `json:"f4"`   // 增值
	F6   float64 `json:"f6"`   // 总市值
	F12  string  `json:"f12"`  // 股票代码
	F104 int     `json:"f104"` // 涨_数量
	F105 int     `json:"f105"` // 跌_数量
	F106 int     `json:"f106"` // 平_数量
}

type queryParams struct {
	Fields    string `json:"fields"`
	CB        string `json:"cb"`
	Fltt      int    `json:"fltt"`
	SECIDS    string `json:"secids"` // 选中的股票ID, 参考: "1.000001,0.399001"
	UT        string `json:"ut"`
	Timestamp int64  `json:"_"`
}

// =============================

func getOverAllData() (data utilRespData, err error) {
	timestamp := time.Now().Unix()
	params := queryParams{
		Fltt:      2,
		SECIDS:    "1.000001",
		UT:        "b2884a393a59ad64002292a3e90d46a5",
		Fields:    "f1,f2,f3,f4,f6,f12,f13,f104,f105,f106",
		Timestamp: timestamp,
	}
	resp, err := util.GetRequireWithParams(reqPaht, params)
	if err != nil {
		err = fmt.Errorf("get data fail: err=%v \n", err)
		return
	}
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("unexpect status: %d \n", resp.StatusCode)
		err = fmt.Errorf("status %d", resp.StatusCode)
		return
	}
	rawResp, _ := ioutil.ReadAll(resp.Body)
	var ret utilRespLayer1
	err = json.Unmarshal(rawResp, &ret)
	if err != nil {
		fmt.Printf("unmarshal fail: err=%v rawResp=%s \n", err, string(rawResp))
		return
	}
	if ret.Data.Total < 1 {
		fmt.Printf("unexpect data: data=%+v \n", ret.Data)
		return
	}
	data = ret.Data.Diff[0]
	return
}
