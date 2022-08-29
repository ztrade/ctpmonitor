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
	hasAuth         bool
	hasLogin        bool
	hasSymbols      bool
	symbols         map[string]*ctp.CThostFtdcInstrumentField
	connectCallback func()
	authCallback    func()
}

func NewTdSpi() *TdSpi {
	td := new(TdSpi)
	td.symbols = make(map[string]*ctp.CThostFtdcInstrumentField)
	return td
}
func (s *TdSpi) GetSymbols() (symbols map[string]*ctp.CThostFtdcInstrumentField) {
	return s.symbols
}

func (b *TdSpi) OnFrontConnected() {
	if b.connectCallback != nil {
		b.connectCallback()
	}
	logrus.Info("TdSpi OnFrontConnected")
}
func (b *TdSpi) OnFrontDisconnected(nReason int) {
	logrus.Info("TdSpi OnFrontDisconnected")

}

func (s *TdSpi) WaitSymbols(ctx context.Context) (err error) {
Out:
	for {
		select {
		case _ = <-ctx.Done():
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
	if pRspInfo != nil && pRspInfo.ErrorID != 0 {
		logrus.Error("OnRspAuthenticate error", pRspInfo.ErrorID, pRspInfo.ErrorMsg)
	}
	buf, _ := json.Marshal(pRspAuthenticateField)
	logrus.Info("OnRspAuthenticate", string(buf))
	s.hasAuth = true
	if s.authCallback != nil {
		s.authCallback()
	}
	return
}
func (s *TdSpi) OnRspUserLogin(pRspUserLogin *ctp.CThostFtdcRspUserLoginField, pRspInfo *ctp.CThostFtdcRspInfoField, nRequestID int, bIsLast bool) {
	if pRspInfo != nil && pRspInfo.ErrorID != 0 {
		logrus.Error("OnRspUserLogin error", pRspInfo.ErrorID, pRspInfo.ErrorMsg)
	}
	buf, _ := json.Marshal(pRspUserLogin)
	logrus.Info("OnRspUserLogin", string(buf))
	s.hasLogin = true
	return
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
		logrus.Error("OnRspQryInstrument error", pRspInfo.ErrorID, pRspInfo.ErrorMsg)
	}
	if pInstrument == nil {
		logrus.Warn("pInstrument is null")
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
		case _ = <-ctx.Done():
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
