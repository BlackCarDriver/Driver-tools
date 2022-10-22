package stock

import (
	"github.com/BlackCarDriver/GoProject-api/color"
	"github.com/BlackCarDriver/GoProject-api/common/util"
)

const configPath = "./config/stock.json" // 配置文件路径

// 配置数据
var stockConfig struct {
	MyStock []monitorCfg `json:"myStock"`
}

type monitorCfg struct {
	Code      string  `json:"code"` // 股票代码
	Mode      int     `json:"mode"` // 1\0
	SellPrize float64 `json:"sell"` // 卖出触发价
	BuyPrize  float64 `json:"buy"`  // 买入触发价
	Tag       string  `json:"tag"`  // 备注
}

// 读取配置
func readConfig() (err error) {
	err = util.UnmarshalJsonFromFile(configPath, &stockConfig)
	if err != nil {
		color.Red("read config fail: err=%v path=%s", err, configPath)
		return
	}
	color.HiBlack("read config success: %+v", stockConfig)
	return
}

// 百分比涨幅计算
func countRiseRange(before, after float64) (rise float64) {
	if before == 0 {
		return 0
	}
	rise = (after - before) / before
	return rise * 100.0
}
