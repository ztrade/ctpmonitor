package monitor

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/ztrade/ctp"
	"github.com/ztrade/ctpmonitor/config"
)

type TdSpi struct {
	ctp.CThostFtdcTraderSpiBase
	hasSymbols   atomic.Bool
	symbols      map[string]*ctp.CThostFtdcInstrumentField
	symbolsCache map[string]*ctp.CThostFtdcInstrumentField
	l            *logrus.Entry

	api *ctp.CThostFtdcTraderApi
	cfg *config.Config
}

func NewTdSpi(cfg *config.Config) *TdSpi {
	td := new(TdSpi)
	td.cfg = cfg
	td.symbols = make(map[string]*ctp.CThostFtdcInstrumentField)
	td.symbolsCache = make(map[string]*ctp.CThostFtdcInstrumentField)
	td.l = logrus.WithField("module", "tdSpi")
	return td
}

func (s *TdSpi) Connect(ctx context.Context) (err error) {
	s.api = ctp.TdCreateFtdcTraderApi("./td/")
	s.api.RegisterSpi(s)
	s.api.RegisterFront(fmt.Sprintf("tcp://%s", s.cfg.TdServer))
	s.api.Init()
	err = s.WaitSymbols(ctx)
	return
}

func (s *TdSpi) GetSymbols() (symbols map[string]*ctp.CThostFtdcInstrumentField) {
	return s.symbols
}

func (s *TdSpi) OnFrontConnected() {
	s.symbols = make(map[string]*ctp.CThostFtdcInstrumentField)
	s.symbolsCache = make(map[string]*ctp.CThostFtdcInstrumentField)
	n := s.api.ReqAuthenticate(&ctp.CThostFtdcReqAuthenticateField{BrokerID: s.cfg.BrokerID, UserID: s.cfg.User, UserProductInfo: "", AuthCode: s.cfg.AuthCode, AppID: s.cfg.AppID}, 0)
	s.l.Info("TdSpi OnFrontConnected, ReqAuthenticate:", n)
}
func (s *TdSpi) OnFrontDisconnected(nReason int) {
	s.hasSymbols.Store(false)
	s.l.Info("TdSpi OnFrontDisconnected")
}

func (s *TdSpi) WaitSymbols(ctx context.Context) (err error) {
Out:
	for {
		select {
		case <-ctx.Done():
			return errors.New("deadline")
		default:
			if s.hasSymbols.Load() {
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
		s.l.Errorf("OnRspAuthenticate error %d %s, retry after 10s", pRspInfo.ErrorID, pRspInfo.ErrorMsg)
		go func() {
			time.Sleep(time.Second * 10)
			n := s.api.ReqAuthenticate(&ctp.CThostFtdcReqAuthenticateField{BrokerID: s.cfg.BrokerID, UserID: s.cfg.User, UserProductInfo: "", AuthCode: s.cfg.AuthCode, AppID: s.cfg.AppID}, 0)
			s.l.Info("OnRspAuthenticate fail, do ReqAuthenticate:", n)
		}()

		return
	}
	n := s.api.ReqUserLogin(&ctp.CThostFtdcReqUserLoginField{UserID: s.cfg.User, BrokerID: s.cfg.BrokerID, Password: s.cfg.Password}, 0)
	s.l.Info("TdSpi OnRspAuthenticate, ReqUserLogin:", n)
}

func (s *TdSpi) OnRspUserLogin(pRspUserLogin *ctp.CThostFtdcRspUserLoginField, pRspInfo *ctp.CThostFtdcRspInfoField, nRequestID int, bIsLast bool) {
	buf, _ := json.Marshal(pRspUserLogin)
	s.l.Info("OnRspUserLogin", string(buf))
	if pRspInfo != nil && pRspInfo.ErrorID != 0 {
		s.l.Errorf("OnRspUserLogin error: %d %s, retry after 10s", pRspInfo.ErrorID, pRspInfo.ErrorMsg)
		go func() {
			time.Sleep(10 * time.Second)
			n := s.api.ReqUserLogin(&ctp.CThostFtdcReqUserLoginField{UserID: s.cfg.User, BrokerID: s.cfg.BrokerID, Password: s.cfg.Password}, 0)
			s.l.Info("OnRspUserLogin fail, do ReqUserLogin:", n)
		}()

		return
	}
	go func() {
		n := s.api.ReqQryInstrument(&ctp.CThostFtdcQryInstrumentField{}, 1)
		s.l.Info("TdSpi OnRspUserLogin, ReqQryInstrument:", n)
	}()
}

func (s *TdSpi) OnRtnInstrumentStatus(pInstrumentStatus *ctp.CThostFtdcInstrumentStatusField) {
	// buf, _ := json.Marshal(pInstrumentStatus)
	// fmt.Println("OnRtnInstrumentStatus:", string(buf))
}

func (s *TdSpi) OnRspQryInstrument(pInstrument *ctp.CThostFtdcInstrumentField, pRspInfo *ctp.CThostFtdcRspInfoField, nRequestID int, bIsLast bool) {
	defer func() {
		if bIsLast {
			if len(s.symbolsCache) != 0 {
				s.symbols = s.symbolsCache
				s.hasSymbols.Store(true)
			} else {
				s.l.Errorf("OnRspQryInstrument return empty, retry after 10s")
				go func() {
					time.Sleep(10 * time.Second)
					n := s.api.ReqQryInstrument(&ctp.CThostFtdcQryInstrumentField{}, 1)
					s.l.Info("TdSpi OnRspQryInstrument no symbols, ReqQryInstrument:", n)
				}()
			}
		}
	}()
	s.l.Info("OnRspQryInstrument:", pInstrument)
	if pRspInfo != nil && pRspInfo.ErrorID != 0 {
		s.l.Error("OnRspQryInstrument error", pRspInfo.ErrorID, pRspInfo.ErrorMsg)
	}
	if pInstrument == nil {
		s.l.Warn("pInstrument is null")
		return
	}
	if s.cfg.Filter != "" && !strings.Contains(s.cfg.Filter, string(pInstrument.ProductClass)) {
		s.l.Infof("ignore symbol by filter: %s,%s, filter: %s", pInstrument.InstrumentID, pInstrument.ProductClass, s.cfg.Filter)
		return
	}
	s.symbolsCache[pInstrument.InstrumentID] = pInstrument

}

func (s *TdSpi) Close() {
	api := s.api
	s.api = nil
	if api != nil {
		api.Release()
		s.l.Info("release tradeApi success")
	}
}
