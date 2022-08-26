package util

import (
	"time"

	"github.com/ztrade/ctpmonitor/model"
	"github.com/ztrade/trademodel"
)

type CTPKline struct {
	cur          *trademodel.Candle
	prevVolume   int64
	prevTurnover float64
}

func NewCTPKline() *CTPKline {
	k := new(CTPKline)
	return k
}

func (k *CTPKline) Update(data *model.MarketData) (candle *trademodel.Candle) {
	defer func() {
		k.prevTurnover = data.Turnover
		k.prevVolume = int64(data.Volume)
	}()
	t := data.Ts
	price := data.LastPrice
	tStart := (t.Unix() / 60) * 60
	// 第一条记录不知道prevVolume和prevTurnover，无法计算这两个值
	if k.cur == nil {
		k.cur = &trademodel.Candle{
			Start:    tStart,
			Open:     price,
			High:     price,
			Low:      price,
			Close:    price,
			Volume:   0,
			Turnover: 0,
		}
		return
	}

	volume := data.Volume - int(k.prevVolume)
	turnover := data.Turnover - k.prevTurnover
	if t.Sub(k.cur.Time()) >= time.Minute {
		candle = k.cur
		k.cur = &trademodel.Candle{
			Start:    tStart,
			Open:     price,
			High:     price,
			Low:      price,
			Close:    price,
			Volume:   float64(volume),
			Turnover: turnover,
		}
		return
	}
	k.cur.Close = price
	k.cur.Volume += float64(volume)
	k.cur.Turnover += turnover

	if price > k.cur.High {
		k.cur.High = price
	}
	if price < k.cur.Low {
		k.cur.Low = price
	}
	return
}
