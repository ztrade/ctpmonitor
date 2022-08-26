package monitor

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/ztrade/ctp"
	"github.com/ztrade/ctpmonitor/config"
	"github.com/ztrade/ctpmonitor/util"
)

type CTPMonitor struct {
	cfg       *config.Config
	tradeApi  *ctp.CThostFtdcTraderApi
	marketApi *ctp.CThostFtdcMdApi
	mdSpi     *mdSpi
	tdSpi     *TdSpi

	isConnect atomic.Bool

	symbols           map[string]*ctp.CThostFtdcInstrumentField
	lastRefreshSymbol time.Time

	isStop chan int
	mutex  sync.Mutex
}

func NewCTPMonitor(cfg *config.Config) (m *CTPMonitor) {
	m = new(CTPMonitor)
	m.cfg = cfg
	m.isStop = make(chan int, 1)
	return
}

func (m *CTPMonitor) Start() (err error) {
	go m.loop()
	return
}

func (m *CTPMonitor) Stop() (err error) {
	close(m.isStop)
	if m.tradeApi != nil {
		m.tradeApi.Release()
	}
	if m.marketApi != nil {
		m.marketApi.Release()
	}
	return
}

func (m *CTPMonitor) reconnect() (err error) {
	if m.tradeApi != nil {
		m.tradeApi.Release()
	}
	if m.marketApi != nil {
		m.marketApi.Release()
	}
	m.tdSpi = NewTdSpi()
	m.mdSpi, err = NewMdSpi(m.cfg)
	if err != nil {
		return
	}
	err = m.connectMdApi()
	if err != nil {
		return
	}
	go func() {
		defer m.isConnect.Store(false)
		m.marketApi.Join()
		logrus.Info("marketApi finished")
	}()
	err = m.connectTdApi()
	if err != nil {
		return
	}
	m.isConnect.Store(true)
	go func() {
		defer m.isConnect.Store(false)
		m.tradeApi.Join()
		logrus.Info("tradeApi finished")
	}()

	t := time.Now()
	if t.Sub(m.lastRefreshSymbol) > time.Hour*24 || t.Day() != m.lastRefreshSymbol.Day() {
		err = m.refreshSymbols()
	}
	m.mdSpi.connectCallback = func() {
		logrus.Info("onFrontendConnected: watch all")
		m.watchAll()
	}

	return
}

func (m *CTPMonitor) connectTdApi() (err error) {
	tdApi := ctp.TdCreateFtdcTraderApi("")
	tdApi.RegisterSpi(m.tdSpi)
	tdApi.RegisterFront(fmt.Sprintf("tcp://%s", m.cfg.TdServer))
	tdApi.Init()
	time.Sleep(time.Second * 3)
	tdApi.ReqAuthenticate(&ctp.CThostFtdcReqAuthenticateField{BrokerID: m.cfg.BrokerID, UserID: m.cfg.User, UserProductInfo: "", AuthCode: m.cfg.AuthCode, AppID: m.cfg.AppID}, 0)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	err = m.tdSpi.WaitAuth(ctx)
	cancel()
	if err != nil {
		logrus.Error("WaitAuth error:", err.Error())
		return
	}
	tdApi.ReqUserLogin(&ctp.CThostFtdcReqUserLoginField{UserID: m.cfg.User, BrokerID: m.cfg.BrokerID, Password: m.cfg.Password}, 0)
	ctx, cancel = context.WithTimeout(context.Background(), time.Second*5)
	err = m.tdSpi.WaitLogin(ctx)
	cancel()
	if err != nil {
		return
	}
	m.tradeApi = tdApi
	return
}

func (m *CTPMonitor) connectMdApi() (err error) {
	api := ctp.MdCreateFtdcMdApi("", false, false)
	api.RegisterFront(fmt.Sprintf("tcp://%s", m.cfg.MdServer))
	api.RegisterSpi(m.mdSpi)
	api.Init()
	time.Sleep(time.Second * 3)
	api.ReqUserLogin(&ctp.CThostFtdcReqUserLoginField{UserID: m.cfg.User, BrokerID: m.cfg.BrokerID, Password: m.cfg.Password}, 0)
	time.Sleep(time.Second * 2)
	m.marketApi = api
	return
}

func (m *CTPMonitor) refreshSymbols() (err error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.tradeApi.ReqQryInstrument(&ctp.CThostFtdcQryInstrumentField{}, 1)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	err = m.tdSpi.WaitSymbols(ctx)
	if err != nil {
		logrus.Error("ReqQryInstrument error:", err.Error())
	}
	symbols := m.tdSpi.GetSymbols()
	if len(symbols) == 0 {
		logrus.Error("refreshSymbols return empty, try again after 10s")
		go func() {
			time.Sleep(time.Second * 10)
			m.refreshSymbols()
		}()
		return
	}
	if !m.symbolsNeedUpdate(symbols) {
		return
	}
	m.symbols = symbols
	m.watchAll()
	return
}

func (m *CTPMonitor) symbolsNeedUpdate(symbols map[string]*ctp.CThostFtdcInstrumentField) bool {
	if len(symbols) == 0 {
		return false
	}
	if len(symbols) != len(m.symbols) {
		return true
	}
	var ok bool
	for k := range symbols {
		_, ok = m.symbols[k]
		if !ok {
			return true
		}
	}
	return false
}

func (m *CTPMonitor) watchAll() {
	for k := range m.symbols {
		logrus.Info("SubscribeMarketData:", k)
		m.marketApi.SubscribeMarketData([]string{k})
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
