package monitor

import (
	"encoding/json"

	"github.com/sirupsen/logrus"
	"github.com/ztrade/ctp"
	"github.com/ztrade/ctpmonitor/config"
	"github.com/ztrade/ctpmonitor/model"
)

type mdSpi struct {
	db                *model.DB
	connectCallback   func()
	loginCallback     func()
	loginFailCallback func()
	l                 *logrus.Entry
}

func NewMdSpi(cfg *config.Config) (spi *mdSpi, err error) {
	spi = new(mdSpi)
	spi.l = logrus.WithField("module", "mdSpi")
	spi.db, err = model.NewDB(cfg.Taos)
	return
}

func (s *mdSpi) OnFrontConnected() {
	s.l.Println("OnFrontConnected")
	if s.connectCallback != nil {
		s.connectCallback()
	}
}
func (s *mdSpi) OnFrontDisconnected(nReason int) {
	s.l.Println("OnFrontDisconnected:", nReason)
}
func (s *mdSpi) OnHeartBeatWarning(nTimeLapse int) {
	s.l.Println("OnHeartBeatWarning:", nTimeLapse)
}
func (s *mdSpi) OnRspUserLogin(pRspUserLogin *ctp.CThostFtdcRspUserLoginField, pRspInfo *ctp.CThostFtdcRspInfoField, nRequestID int, bIsLast bool) {
	// logrus.Infof("%d isLast: %t login success: %s, user %s, time: %s, systemName: %s, frontID: %s, session: %d", nRequestID, bIsLast, pRspUserLogin.TradingDay, pRspUserLogin.UserID, pRspUserLogin.SystemName, pRspUserLogin.FrontID, pRspUserLogin.SessionID)

	if pRspInfo.ErrorID == 0 && s.loginCallback != nil {
		s.loginCallback()
	}
	if pRspInfo.ErrorID != 0 && s.loginFailCallback != nil {
		s.loginFailCallback()
	}
	s.l.Infof("login error: %d %s", pRspInfo.ErrorID, pRspInfo.ErrorMsg)
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
	return s.db.Close()
}
