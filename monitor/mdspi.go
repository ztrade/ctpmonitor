package monitor

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/ztrade/ctp"
	"github.com/ztrade/ctpmonitor/config"
	"github.com/ztrade/ctpmonitor/model"
)

type mdSpi struct {
	db *model.DB

	l   *logrus.Entry
	api *ctp.CThostFtdcMdApi
	cfg *config.Config

	loginCallback func(*ctp.CThostFtdcMdApi)
}

func NewMdSpi(cfg *config.Config) (spi *mdSpi, err error) {
	spi = new(mdSpi)
	spi.cfg = cfg
	spi.l = logrus.WithField("module", "mdSpi")
	spi.db, err = model.NewDB(cfg.Taos)
	return
}

func (s *mdSpi) Connect(ctx context.Context) (err error) {
	s.api = ctp.MdCreateFtdcMdApi("./md/", false, false)
	s.api.RegisterFront(fmt.Sprintf("tcp://%s", s.cfg.MdServer))
	s.api.RegisterSpi(s)
	s.api.Init()
	return
}

func (s *mdSpi) OnFrontConnected() {
	n := s.api.ReqUserLogin(&ctp.CThostFtdcReqUserLoginField{UserID: s.cfg.User, BrokerID: s.cfg.BrokerID, Password: s.cfg.Password}, 0)
	s.l.Println("OnFrontConnected:", n)
}

func (s *mdSpi) OnFrontDisconnected(nReason int) {
	s.l.Println("OnFrontDisconnected:", nReason)
}

func (s *mdSpi) OnHeartBeatWarning(nTimeLapse int) {
	s.l.Println("OnHeartBeatWarning:", nTimeLapse)
}

func (s *mdSpi) OnRspUserLogin(pRspUserLogin *ctp.CThostFtdcRspUserLoginField, pRspInfo *ctp.CThostFtdcRspInfoField, nRequestID int, bIsLast bool) {
	// logrus.Infof("%d isLast: %t login success: %s, user %s, time: %s, systemName: %s, frontID: %s, session: %d", nRequestID, bIsLast, pRspUserLogin.TradingDay, pRspUserLogin.UserID, pRspUserLogin.SystemName, pRspUserLogin.FrontID, pRspUserLogin.SessionID)

	s.l.Infof("login: %d %s", pRspInfo.ErrorID, pRspInfo.ErrorMsg)
	if pRspInfo.ErrorID == 0 {
		if s.loginCallback != nil {
			s.loginCallback(s.api)
		}
	} else {
		s.l.Println("OnRspUserLogin fail, retry after 10s")
		go func() {
			time.Sleep(10 * time.Second)
			n := s.api.ReqUserLogin(&ctp.CThostFtdcReqUserLoginField{UserID: s.cfg.User, BrokerID: s.cfg.BrokerID, Password: s.cfg.Password}, 0)
			s.l.Info("OnRspUserLogin fail, do ReqUserLogin:", n)
		}()
	}

}

func (s *mdSpi) OnRspUserLogout(pUserLogout *ctp.CThostFtdcUserLogoutField, pRspInfo *ctp.CThostFtdcRspInfoField, nRequestID int, bIsLast bool) {
	buf, _ := json.Marshal(pUserLogout)
	s.l.Infof("%d isLast: %t login success: %s", nRequestID, bIsLast, string(buf))
	s.l.Infof("logout error: %d %s", pRspInfo.ErrorID, pRspInfo.ErrorMsg)
}

func (s *mdSpi) OnRspQryMulticastInstrument(pMulticastInstrument *ctp.CThostFtdcMulticastInstrumentField, pRspInfo *ctp.CThostFtdcRspInfoField, nRequestID int, bIsLast bool) {
	buf, _ := json.Marshal(pMulticastInstrument)
	s.l.Infof("%d isLast: %t logout success: %s", nRequestID, bIsLast, string(buf))
	s.l.Infof("OnRspQryMulticastInstrument error: %d %s", pRspInfo.ErrorID, pRspInfo.ErrorMsg)
}
func (s *mdSpi) OnRspError(pRspInfo *ctp.CThostFtdcRspInfoField, nRequestID int, bIsLast bool) {
	s.l.Warnf("%d resp error: %t:ErrorID: %d ErrorMsg:%s", nRequestID, bIsLast, pRspInfo.ErrorID, pRspInfo.ErrorMsg)
}
func (s *mdSpi) OnRspSubMarketData(pSpecificInstrument *ctp.CThostFtdcSpecificInstrumentField, pRspInfo *ctp.CThostFtdcRspInfoField, nRequestID int, bIsLast bool) {
	s.l.Info("onSubMarketData:", pSpecificInstrument.InstrumentID)
	if pRspInfo != nil && pRspInfo.ErrorID != 0 {
		s.l.Warnf("%d onSubMarketData: %t: ErrorID: %d ErrorMsg:%s", nRequestID, bIsLast, pRspInfo.ErrorID, pRspInfo.ErrorMsg)
	}
}
func (s *mdSpi) OnRspUnSubMarketData(pSpecificInstrument *ctp.CThostFtdcSpecificInstrumentField, pRspInfo *ctp.CThostFtdcRspInfoField, nRequestID int, bIsLast bool) {
	s.l.Info("onUnSubMarketData:", pSpecificInstrument.InstrumentID)
	s.l.Warnf("%d onUnSubMarketData: %t: ErrorID: %d ErrorMsg:%s", nRequestID, bIsLast, pRspInfo.ErrorID, pRspInfo.ErrorMsg)
}
func (s *mdSpi) OnRspSubForQuoteRsp(pSpecificInstrument *ctp.CThostFtdcSpecificInstrumentField, pRspInfo *ctp.CThostFtdcRspInfoField, nRequestID int, bIsLast bool) {
	s.l.Info("onSubForQuoteRsp:", pSpecificInstrument.InstrumentID)
	s.l.Warnf("%d onSubForQuoteRsp: %t: ErrorID: %d ErrorMsg:%s", nRequestID, bIsLast, pRspInfo.ErrorID, pRspInfo.ErrorMsg)
}
func (s *mdSpi) OnRspUnSubForQuoteRsp(pSpecificInstrument *ctp.CThostFtdcSpecificInstrumentField, pRspInfo *ctp.CThostFtdcRspInfoField, nRequestID int, bIsLast bool) {
	s.l.Info("onUnSubForQuoteRsp:", pSpecificInstrument.InstrumentID)
	s.l.Warnf("%d onUnSubForQuoteRsp: %t: ErrorID: %d ErrorMsg:%s", nRequestID, bIsLast, pRspInfo.ErrorID, pRspInfo.ErrorMsg)
}
func (s *mdSpi) OnRtnDepthMarketData(pDepthMarketData *ctp.CThostFtdcDepthMarketDataField) {
	if pDepthMarketData == nil {
		s.l.Error("marketdata is nil")
		return
	}
	// buf, _ := json.Marshal(pDepthMarketData)
	// fmt.Println(string(buf))
	err := s.db.AddMarketData(pDepthMarketData)
	if err != nil {
		s.l.Errorf("add marketdata failed: %s", err.Error())
	}
}
func (s *mdSpi) OnRtnForQuoteRsp(pForQuoteRsp *ctp.CThostFtdcForQuoteRspField) {
	buf, _ := json.Marshal(pForQuoteRsp)
	s.l.Info("onForQuoteRsp:", string(buf))
}

func (s *mdSpi) Close() error {
	api := s.api
	s.api = nil
	if api != nil {
		api.Release()
		s.l.Info("release marketApi success")
	}
	return s.db.Close()
}
