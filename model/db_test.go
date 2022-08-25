package model

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/ztrade/ctp"
)

func TestDB(t *testing.T) {
	db := &DB{}
	db.database = "ctptest"
	db.writeOnce = 100
	db.chs = make(map[string]chan *MarketData)
	db.cache = 10000
	err := db.connect("root:123456@tcp(home.in:6030)/")
	if err != nil {
		t.Fatal(err.Error())
	}
	defer db.Close()
	str := `{"TradingDay":"20220823","InstrumentID":"al2201","ExchangeID":"","ExchangeInstID":"","LastPrice":18350,"PreSettlementPrice":18580,"PreClosePrice":18475,"PreOpenInterest":236229,"OpenPrice":18475,"HighestPrice":18490,"LowestPrice":18230,"Volume":165674,"Turnover":15192326925,"OpenInterest":250293,"ClosePrice":0,"SettlementPrice":0,"UpperLimitPrice":20065,"LowerLimitPrice":17090,"PreDelta":0,"CurrDelta":0,"UpdateTime":"23:40:51","UpdateMillisec":500,"BidPrice1":18350,"BidVolume1":25,"AskPrice1":18355,"AskVolume1":71,"BidPrice2":18345,"BidVolume2":43,"AskPrice2":18360,"AskVolume2":83,"BidPrice3":18340,"BidVolume3":59,"AskPrice3":18365,"AskVolume3":48,"BidPrice4":18335,"BidVolume4":61,"AskPrice4":18370,"AskVolume4":75,"BidPrice5":18330,"BidVolume5":69,"AskPrice5":18375,"AskVolume5":155,"AveragePrice":91700.12750944626,"ActionDay":"20220823"}`
	data := ctp.CThostFtdcDepthMarketDataField{}
	err = json.Unmarshal([]byte(str), &data)
	if err != nil {
		t.Fatal(err.Error())
	}
	db.cache = 1
	err = db.AddMarketData(&data)
	if err != nil {
		t.Fatal(err.Error())
	}
	time.Sleep(time.Second * 10)
	datas, err := db.GetDatas("al2201", time.Now().Add(time.Hour*-24), time.Now().Add(time.Hour*24))
	if err != nil {
		t.Fatal(err.Error())
	}
	buf, err := json.Marshal(datas)
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Log(string(buf))
}
