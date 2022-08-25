package model

import "time"

type MarketData struct {
	Ts                 time.Time `db:"ts"`
	TradingDay         string    `db:"trading_day"`
	InstrumentID       string    `db:"instrument_id"`
	ExchangeID         string    `db:"exchange_id"`
	ExchangeInstID     string    `db:"exchange_inst_id"`
	ActionDay          string
	LastPrice          float64
	PreSettlementPrice float64
	PreClosePrice      float64
	PreOpenInterest    float64
	OpenPrice          float64
	HighestPrice       float64
	LowestPrice        float64
	Volume             int
	Turnover           float64
	OpenInterest       float64
	ClosePrice         float64
	SettlementPrice    float64
	UpperLimitPrice    float64
	LowerLimitPrice    float64
	PreDelta           float64
	CurrDelta          float64
	BidPrice1          float64
	BidVolume1         int
	AskPrice1          float64
	AskVolume1         int
	BidPrice2          float64
	BidVolume2         int
	AskPrice2          float64
	AskVolume2         int
	BidPrice3          float64
	BidVolume3         int
	AskPrice3          float64
	AskVolume3         int
	BidPrice4          float64
	BidVolume4         int
	AskPrice4          float64
	AskVolume4         int
	BidPrice5          float64
	BidVolume5         int
	AskPrice5          float64
	AskVolume5         int
	AveragePrice       float64
}
