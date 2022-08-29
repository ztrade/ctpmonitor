package monitor

import (
	"encoding/json"

	"github.com/sirupsen/logrus"
	"github.com/ztrade/ctp"
	"github.com/ztrade/ctpmonitor/config"
	"github.com/ztrade/ctpmonitor/model"
)

type mdSpi struct {
	db              *model.DB
	connectCallback func()
	loginCallback   func()
}

func NewMdSpi(cfg *config.Config) (spi *mdSpi, err error) {
	spi = new(mdSpi)
	spi.db, err = model.NewDB(cfg.Taos)
	return
}

func (s *mdSpi) OnFrontConnected() {
	logrus.Println("mdSpi OnFrontConnected")
	if s.connectCallback != nil {
		s.connectCallback()
	}
}
func (s *mdSpi) OnFrontDisconnected(nReason int) {
	logrus.Println("mdSpi OnFrontDisconnected:", nReason)
}
func (s *mdSpi) OnHeartBeatWarning(nTimeLapse int) {
	logrus.Println("mdSpi OnHeartBeatWarning:", nTimeLapse)
}
func (s *mdSpi) OnRspUserLogin(pRspUserLogin *ctp.CThostFtdcRspUserLoginField, pRspInfo *ctp.CThostFtdcRspInfoField, nRequestID int, bIsLast bool) {
	// logrus.Infof("%d isLast: %t login success: %s, user %s, time: %s, systemName: %s, frontID: %s, session: %d", nRequestID, bIsLast, pRspUserLogin.TradingDay, pRspUserLogin.UserID, pRspUserLogin.SystemName, pRspUserLogin.FrontID, pRspUserLogin.SessionID)

	if pRspInfo.ErrorID == 0 && s.loginCallback != nil {
		s.loginCallback()
	}
	logrus.Infof("login error: %d %s", pRspInfo.ErrorID, pRspInfo.ErrorMsg)
}
func (s *mdSpi) OnRspUserLogout(pUserLogout *ctp.CThostFtdcUserLogoutField, pRspInfo *ctp.CThostFtdcRspInfoField, nRequestID int, bIsLast bool) {
	buf, _ := json.Marshal(pUserLogout)
	logrus.Infof("%d isLast: %t login success: %s", nRequestID, bIsLast, string(buf))
	logrus.Infof("login error: %d %s", pRspInfo.ErrorID, pRspInfo.ErrorMsg)
}
func (s *mdSpi) OnRspQryMulticastInstrument(pMulticastInstrument *ctp.CThostFtdcMulticastInstrumentField, pRspInfo *ctp.CThostFtdcRspInfoField, nRequestID int, bIsLast bool) {
	buf, _ := json.Marshal(pMulticastInstrument)
	logrus.Infof("%d isLast: %t logout success: %s", nRequestID, bIsLast, string(buf))
	logrus.Infof("logout error: %d %s", pRspInfo.ErrorID, pRspInfo.ErrorMsg)
}
func (s *mdSpi) OnRspError(pRspInfo *ctp.CThostFtdcRspInfoField, nRequestID int, bIsLast bool) {
	logrus.Warnf("%d resp error: %t:ErrorID: %d ErrorMsg:%s", nRequestID, bIsLast, pRspInfo.ErrorID, pRspInfo.ErrorMsg)
}
func (s *mdSpi) OnRspSubMarketData(pSpecificInstrument *ctp.CThostFtdcSpecificInstrumentField, pRspInfo *ctp.CThostFtdcRspInfoField, nRequestID int, bIsLast bool) {
	logrus.Info("onSubMarketData:", pSpecificInstrument.InstrumentID)
	if pRspInfo != nil && pRspInfo.ErrorID != 0 {
		logrus.Warnf("%d onSubMarketData: %t: ErrorID: %d ErrorMsg:%s", nRequestID, bIsLast, pRspInfo.ErrorID, pRspInfo.ErrorMsg)
	}
}
func (s *mdSpi) OnRspUnSubMarketData(pSpecificInstrument *ctp.CThostFtdcSpecificInstrumentField, pRspInfo *ctp.CThostFtdcRspInfoField, nRequestID int, bIsLast bool) {
	logrus.Info("onUnSubMarketData:", pSpecificInstrument.InstrumentID)
	logrus.Warnf("%d onUnSubMarketData: %t: ErrorID: %d ErrorMsg:%s", nRequestID, bIsLast, pRspInfo.ErrorID, pRspInfo.ErrorMsg)
}
func (s *mdSpi) OnRspSubForQuoteRsp(pSpecificInstrument *ctp.CThostFtdcSpecificInstrumentField, pRspInfo *ctp.CThostFtdcRspInfoField, nRequestID int, bIsLast bool) {
	logrus.Info("onSubForQuoteRsp:", pSpecificInstrument.InstrumentID)
	logrus.Warnf("%d onSubForQuoteRsp: %t: ErrorID: %d ErrorMsg:%s", nRequestID, bIsLast, pRspInfo.ErrorID, pRspInfo.ErrorMsg)
}
func (s *mdSpi) OnRspUnSubForQuoteRsp(pSpecificInstrument *ctp.CThostFtdcSpecificInstrumentField, pRspInfo *ctp.CThostFtdcRspInfoField, nRequestID int, bIsLast bool) {
	logrus.Info("onUnSubForQuoteRsp:", pSpecificInstrument.InstrumentID)
	logrus.Warnf("%d onUnSubForQuoteRsp: %t: ErrorID: %d ErrorMsg:%s", nRequestID, bIsLast, pRspInfo.ErrorID, pRspInfo.ErrorMsg)
}
func (s *mdSpi) OnRtnDepthMarketData(pDepthMarketData *ctp.CThostFtdcDepthMarketDataField) {
	if pDepthMarketData == nil {
		logrus.Errorf("marketdata is nil")
		return
	}
	err := s.db.AddMarketData(pDepthMarketData)
	if err != nil {
		logrus.Errorf("add marketdata failed: %s", err.Error())
	}
}
func (s *mdSpi) OnRtnForQuoteRsp(pForQuoteRsp *ctp.CThostFtdcForQuoteRspField) {
	buf, _ := json.Marshal(pForQuoteRsp)
	logrus.Info("onForQuoteRsp:", string(buf))
}

func (s *mdSpi) Close() error {
	return s.db.Close()
}
