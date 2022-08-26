package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jinzhu/copier"
	"github.com/ztrade/base/common"
	"github.com/ztrade/ctpmonitor/config"
	"github.com/ztrade/ctpmonitor/model"
	"github.com/ztrade/ctpmonitor/pb"
	"github.com/ztrade/ctpmonitor/util"
	"github.com/ztrade/trademodel"
)

var (
	i64 int64
)

type CtpService struct {
	pb.UnimplementedCtpServer
	db *model.DB
}

func NewCtpService(cfg *config.Config) (s *CtpService, err error) {
	db, err := model.NewDB(cfg.Taos)
	if err != nil {
		return
	}
	s = new(CtpService)
	s.db = db
	return
}

func (s *CtpService) GetKline(ctx context.Context, req *pb.KlineReq) (resp *pb.KlineResp, err error) {
	tStart := time.UnixMilli(req.Start)
	tEnd := time.UnixMilli(req.End)
	datas, err := s.db.GetDatas(req.Symbol, tStart, tEnd)
	if err != nil {
		return
	}
	fmt.Println(tStart, tEnd, req.Bin)
	resp = &pb.KlineResp{}
	kl := util.NewCTPKline()
	var candle *trademodel.Candle
	var mergedCandle interface{}
	merge := common.NewKlineMergeStr("1m", req.Bin)
	for _, v := range datas {
		candle = kl.Update(&v)
		if candle != nil {
			mergedCandle = merge.Update(candle)
			if mergedCandle != nil {
				temp := new(pb.KlineData)
				copier.Copy(&temp, mergedCandle)
				resp.Data = append(resp.Data, temp)
			}
		}
	}

	return
}

func (s *CtpService) GetTick(ctx context.Context, req *pb.RangeReq) (resp *pb.MarketDatas, err error) {
	tStart := time.UnixMilli(req.Start)
	tEnd := time.UnixMilli(req.End)
	datas, err := s.db.GetDatas(req.Symbol, tStart, tEnd)
	if err != nil {
		return
	}
	n := len(datas)
	resp = &pb.MarketDatas{
		Data: make([]*pb.MarketData, n),
	}

	err = copier.CopyWithOption(&resp.Data, &datas, copier.Option{
		IgnoreEmpty: true,
		DeepCopy:    true,
		Converters: []copier.TypeConverter{
			{
				SrcType: time.Time{},
				DstType: i64,
				Fn: func(src interface{}) (interface{}, error) {
					s, ok := src.(time.Time)
					if !ok {
						return nil, errors.New("src type not matching")
					}
					return s.UnixMilli(), nil
				},
			},
		},
	})

	if err != nil {
		return
	}
	return
}
