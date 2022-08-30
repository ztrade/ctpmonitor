package monitor

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/ztrade/ctp"
)

type TdSpi struct {
	ctp.CThostFtdcTraderSpiBase
	hasAuth           bool
	hasLogin          bool
	hasSymbols        bool
	symbols           map[string]*ctp.CThostFtdcInstrumentField
	connectCallback   func()
	authCallback      func()
	authFailCallback  func()
	loginFailCallback func()
	l                 *logrus.Entry
}

func NewTdSpi() *TdSpi {
	td := new(TdSpi)
	td.symbols = make(map[string]*ctp.CThostFtdcInstrumentField)
	td.l = logrus.WithField("module", "tdSpi")
	return td
}
func (s *TdSpi) GetSymbols() (symbols map[string]*ctp.CThostFtdcInstrumentField) {
	return s.symbols
}

func (s *TdSpi) OnFrontConnected() {
	if s.connectCallback != nil {
		s.connectCallback()
	}
	s.l.Info("TdSpi OnFrontConnected")
}
func (s *TdSpi) OnFrontDisconnected(nReason int) {
	s.l.Info("TdSpi OnFrontDisconnected")

}

func (s *TdSpi) WaitSymbols(ctx context.Context) (err error) {
Out:
	for {
		select {
		case <-ctx.Done():
			return errors.New("deadline")
		default:
			if s.hasSymbols {
				break Out
			}
			time.Sleep(time.Millisecond)
		}
	}
	return
}
func (s *TdSpi) WaitAuth(ctx context.Context) (err error) {
Out:
	for {
		select {
		case <-ctx.Done():
			return errors.New("deadline")
		default:
			if s.hasAuth {
				break Out
			}
			time.Sleep(time.Millisecond)
		}
	}
	return
}

func (s *TdSpi) OnRspAuthenticate(pRspAuthenticateField *ctp.CThostFtdcRspAuthenticateField, pRspInfo *ctp.CThostFtdcRspInfoField, nRequestID int, bIsLast bool) {
	buf, _ := json.Marshal(pRspAuthenticateField)
	s.l.Info("OnRspAuthenticate", string(buf))
	if pRspInfo != nil && pRspInfo.ErrorID != 0 {
		s.l.Errorf("OnRspAuthenticate error %d %s", pRspInfo.ErrorID, pRspInfo.ErrorMsg)
		if s.authFailCallback != nil {
			s.authFailCallback()
		}
		return
	}
	s.hasAuth = true
	if s.authCallback != nil {
		s.authCallback()
	}
}
func (s *TdSpi) OnRspUserLogin(pRspUserLogin *ctp.CThostFtdcRspUserLoginField, pRspInfo *ctp.CThostFtdcRspInfoField, nRequestID int, bIsLast bool) {
	buf, _ := json.Marshal(pRspUserLogin)
	s.l.Info("OnRspUserLogin", string(buf))
	if pRspInfo != nil && pRspInfo.ErrorID != 0 {
		s.l.Errorf("OnRspUserLogin error: %d %s", pRspInfo.ErrorID, pRspInfo.ErrorMsg)
		if s.loginFailCallback != nil {
			s.loginFailCallback()
		}
		return
	}
	s.hasLogin = true
}
func (s *TdSpi) OnRtnInstrumentStatus(pInstrumentStatus *ctp.CThostFtdcInstrumentStatusField) {
	// buf, _ := json.Marshal(pInstrumentStatus)
	// fmt.Println("OnRtnInstrumentStatus:", string(buf))
}

func (s *TdSpi) OnRspQryInstrument(pInstrument *ctp.CThostFtdcInstrumentField, pRspInfo *ctp.CThostFtdcRspInfoField, nRequestID int, bIsLast bool) {
	defer func() {
		if bIsLast {
			s.hasSymbols = true
		}
	}()
	if pRspInfo != nil && pRspInfo.ErrorID != 0 {
		s.l.Error("OnRspQryInstrument error", pRspInfo.ErrorID, pRspInfo.ErrorMsg)
	}
	if pInstrument == nil {
		s.l.Warn("pInstrument is null")
		return
	}
	if pInstrument.ProductClass != '1' {
		return
	}
	s.symbols[pInstrument.InstrumentID] = pInstrument

}
func (s *TdSpi) WaitLogin(ctx context.Context) (err error) {
Out:
	for {
		select {
		case <-ctx.Done():
			return errors.New("deadline")
		default:
			if s.hasLogin {
				break Out
			}
			time.Sleep(time.Millisecond)
		}
	}
	return
}
