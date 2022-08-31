package monitor

import (
	"context"
	"os"
	"sync/atomic"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/ztrade/ctp"
	"github.com/ztrade/ctpmonitor/config"
	"github.com/ztrade/ctpmonitor/util"
)

type CTPMonitor struct {
	cfg   *config.Config
	mdSpi *mdSpi
	tdSpi *TdSpi

	isConnect atomic.Bool

	symbols map[string]*ctp.CThostFtdcInstrumentField

	isStop chan int
}

func NewCTPMonitor(cfg *config.Config) (m *CTPMonitor) {
	m = new(CTPMonitor)
	m.cfg = cfg
	m.isStop = make(chan int, 1)
	os.MkdirAll("md", os.ModePerm)
	os.MkdirAll("td", os.ModePerm)
	return
}

func (m *CTPMonitor) Start() (err error) {
	go m.loop()
	return
}

func (m *CTPMonitor) Stop() (err error) {
	close(m.isStop)
	logrus.Info("close mdSpi")
	if m.mdSpi != nil {
		m.mdSpi.Close()
	}
	logrus.Info("close mdSpi success")
	if m.tdSpi != nil {
		m.tdSpi.Close()
	}
	logrus.Info("Stop success")
	return
}

func (m *CTPMonitor) reconnect() (err error) {
	if m.tdSpi != nil {
		m.tdSpi.Close()
		m.tdSpi = nil
	}
	if m.mdSpi != nil {
		m.mdSpi.Close()
		m.tdSpi = nil
	}
	m.tdSpi = NewTdSpi(m.cfg)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	err = m.tdSpi.Connect(ctx)
	cancel()
	if err != nil {
		return
	}

	m.mdSpi, err = NewMdSpi(m.cfg)
	if err != nil {
		return
	}
	m.mdSpi.loginCallback = m.watchAll
	ctx, cancel = context.WithTimeout(context.Background(), time.Second*10)
	err = m.mdSpi.Connect(ctx)
	if err != nil {
		return
	}
	// go func() {
	// 	defer m.isConnect.Store(false)
	// 	m.marketApi.Join()
	// 	logrus.Info("marketApi finished")
	// }()

	m.isConnect.Store(true)
	// go func() {
	// 	defer m.isConnect.Store(false)
	// 	m.tradeApi.Join()
	// 	logrus.Info("tradeApi finished")
	// }()
	// _, err = m.refreshSymbols()
	return
}

func (m *CTPMonitor) watchAll(api *ctp.CThostFtdcMdApi) {
	m.symbols = m.tdSpi.GetSymbols()
	for k := range m.symbols {
		logrus.Info("SubscribeMarketData:", k)
		api.SubscribeMarketData([]string{k})
	}
}

func (m *CTPMonitor) loop() {
	var weekDay time.Weekday
	var err error
	var t time.Time
	var needConnect bool
Out:
	for {
		needConnect = false
		select {
		case <-m.isStop:
			break Out
		default:
		}
		t = time.Now()
		weekDay = t.Weekday()
		if weekDay == time.Sunday || weekDay == time.Saturday {
			time.Sleep(time.Hour)
			continue
		}
		if m.isConnect.Load() {
			time.Sleep(time.Minute)
			continue
		}
		timeMinute := util.DayMinute(t.Hour()*60 + t.Minute())
		for _, v := range util.TradeTime {
			// 提前5分钟重连
			if (v.Start-5) < timeMinute && (v.End+5) > timeMinute {
				needConnect = true
				break
			}
		}
		needConnect = true
		if !needConnect {
			logrus.Println("wait time")
			time.Sleep(time.Minute)
			continue
		}
		// 重连
		logrus.Infof("reconnect")
		for {
			err = m.reconnect()
			logrus.Info("reconnect status:", err)
			if err == nil {
				time.Sleep(time.Minute)
				break
			}
		}
	}
}
