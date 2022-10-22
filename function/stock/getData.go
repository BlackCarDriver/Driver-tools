package stock

import (
	"encoding/json"
	"fmt"
	"github.com/BlackCarDriver/GoProject-api/color"
	"github.com/BlackCarDriver/GoProject-api/common/util"
	"io/ioutil"
	"net/http"
	"time"
)

const reqPathOverall = "https://push2.eastmoney.com/api/qt/ulist.np/get"
const reqPathStock = "https://push2.eastmoney.com/api/qt/stock/get"

type queryOverallParams struct {
	Fields    string `json:"fields"`
	CB        string `json:"cb"`
	Fltt      int    `json:"fltt"`
	SECIDS    string `json:"secids"` // 选中的股票ID, 参考: "1.000001,0.399001"
	UT        string `json:"ut"`
	Timestamp int64  `json:"_"`
}

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

// https://push2.eastmoney.com/api/qt/stock/get?secid=1.510500&cb=jQuery351008508713402356394_1666188681731&ut=6d2ffaa6a585d612eda28417681d58fb&fields=f57,f58,f59,f152,f43,f169,f170,f60,f44,f45,f168,f50,f47,f48,f49,f46,f78,f85,f86,f169,f117,f107,f116,f117,f118,f163,f171,f113,f114,f115,f161,f162,f164,f168,f177,f180&invt=2&_=1666188681745
// https://push2.eastmoney.com/api/qt/stock/get?
// cb=jQuery351008508713402356394_1666188681731&
// ut=6d2ffaa6a585d612eda28417681d58fb&
// fields=f57,f58,f59,f152,f43,f169,f170,f60,f44,f45,f168,f50,f47,f48,f49,f46,f78,f85,f86,f169,f117,f107,f116,f117,f118,f163,f171,f113,f114,f115,f161,f162,f164,f168,f177,f180&
// secid=1.000905
// invt=2&
// _=1666188681745

type queryStockParams struct {
	CB        string `json:"cb"`
	UT        string `json:"ut"`
	Fields    string `json:"fields"`
	SECID     string `json:"secid"` // 选中的股票ID, 参考: "1.000001,0.399001"
	Invt      int    `json:"invt"`
	Timestamp int64  `json:"_"`
}

type GetStockResp struct {
	RC     int               `json:"rc"`
	RT     int               `json:"rt"`
	SVR    int               `json:"svr"`
	LT     int               `json:"lt"`
	FULL   int               `json:"full"`
	DLMKTS string            `json:"dlmkts"`
	Data   *GetStockRespData `json:"data"`
}

type GetStockRespData struct {
	F43  int64   `json:"F43"`  // 最新报价()
	F44  int64   `json:"f44"`  // 最高(分)
	F45  int64   `json:"f45"`  // 最低(分)
	F46  int64   `json:"F46"`  // 今开(分)
	F47  int64   `json:"F47"`  // 成交量(元)
	F48  float64 `json:"f48"`  // 成交额
	F49  int64   `json:"f49"`  // 外盘
	F50  int64   `json:"f50"`  // 量比(/100)
	F57  string  `json:"F57"`  // 股票代号
	F58  string  `json:"F58"`  // 股票名称
	F59  int     `json:"f59"`  // 精准度(小数点后保留多少位)
	F60  int64   `json:"f60"`  // 昨收(分)
	F86  int64   `json:"f86"`  // 数据更新时间 (秒)
	F161 int64   `json:"F161"` // 外盘
	F168 int64   `json:"f168"` // 换手率 (/100)
	F169 int64   `json:"F169"` // 涨跌 (分)
	F170 int64   `json:"F170"` // 涨跌幅 (/100)
	F171 int64   `json:"F171"` // 振幅 (/100)
}

// =============================

// 查询个股数据
func getStockData(mode int, code string) (data *GetStockRespData, err error) {
	timestamp := time.Now().Unix()
	params := queryStockParams{
		SECID:     fmt.Sprintf("%d.%s", mode, code),
		UT:        "b2884a393a59ad64002292a3e90d46a5",
		Fields:    "f43,f44,445,f46,f47,f48,f49,f50,f57,f58,f59,f60,f86,f161,f168,f169,f170,f171",
		Invt:      2,
		Timestamp: timestamp,
	}
	resp, err := util.GetRequireWithParams(reqPathStock, params)
	if err != nil {
		err = fmt.Errorf("get data fail: err=%w \n", err)
		return
	}
	if resp.StatusCode != http.StatusOK {
		color.HiRed("unexpect status: %d \n", resp.StatusCode)
		err = fmt.Errorf("status %d", resp.StatusCode)
		return
	}
	rawResp, _ := ioutil.ReadAll(resp.Body)
	var ret GetStockResp
	err = json.Unmarshal(rawResp, &ret)
	if err != nil {
		color.HiRed("unmarshal fail: err=%v rawResp=%s \n", err, string(rawResp))
		return
	}
	if ret.Data == nil {
		color.HiRed("data not fund in content: rawResp=%s", string(rawResp))
		err = fmt.Errorf("data not found")
		return
	}
	data = ret.Data
	return
}

// 查询大盘数据
func getOverAllData() (data utilRespData, err error) {
	timestamp := time.Now().Unix()
	params := queryOverallParams{
		Fltt:      2,
		SECIDS:    "1.000001",
		UT:        "b2884a393a59ad64002292a3e90d46a5",
		Fields:    "f1,f2,f3,f4,f6,f12,f13,f104,f105,f106",
		Timestamp: timestamp,
	}
	resp, err := util.GetRequireWithParams(reqPathOverall, params)
	if err != nil {
		err = fmt.Errorf("get data fail: err=%w \n", err)
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

func (d *GetStockRespData) GetDesc() (desc string) {
	return fmt.Sprintf("%s  %.2f  %.2f \n", d.F58, float64(d.F43)/100, float64(d.F170)/100)
}
