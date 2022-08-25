package model

import (
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/copier"
	"github.com/jmoiron/sqlx"
	"github.com/lunny/log"
	"github.com/sirupsen/logrus"
	_ "github.com/taosdata/driver-go/v3/taosSql"
	"github.com/ztrade/ctp"
	_ "modernc.org/sqlite"
)

var (
	matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
	matchAllCap   = regexp.MustCompile("([a-z0-9])([A-Z])")
)

func CamelToSnakeCase(str string) string {
	snake := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}

func init() {
	sqlx.NameMapper = CamelToSnakeCase
}

type DB struct {
	database  string
	db        *sqlx.DB
	writeOnce int
	mutex     sync.Mutex
	closed    bool
	wg        sync.WaitGroup

	chs   map[string]chan *MarketData
	cache int
}

func NewDB(uri string) (db *DB, err error) {
	db = &DB{}
	db.database = "ctp"
	db.writeOnce = 100
	db.chs = make(map[string]chan *MarketData)
	db.cache = 10000
	err = db.connect(uri)
	return
}

func (db *DB) connect(uri string) (err error) {
	db.db, err = sqlx.Connect("taosSql", uri)
	if err != nil {
		return
	}
	_, err = db.db.Exec("create database if not exists ? KEEP 365000 DURATION 10 BUFFER 16 WAL_LEVEL 1;", db.database)
	if err != nil {
		err = fmt.Errorf("init database failed:%w", err)
		return
	}
	_, err = db.db.Exec("use ?", db.database)
	if err != nil {
		err = fmt.Errorf("use database failed:%w", err)
		return
	}

	_, err = db.db.Exec("CREATE STABLE IF NOT EXISTS market (ts timestamp,trading_day binary(20),action_day binary(10), last_price float, pre_settlement_price float, pre_close_price float, pre_open_interest float, open_price float, highest_price float, lowest_price float, volume BIGINT, turnover float, open_interest float, close_price float, settlement_price float, upper_limit_price float, lower_limit_price float, pre_delta float, curr_delta float, bid_price1 float, bid_volume1 BIGINT, ask_price1 float, ask_volume1 BIGINT, bid_price2 float, bid_volume2 BIGINT, ask_price2 float, ask_volume2 BIGINT, bid_price3 float, bid_volume3 BIGINT, ask_price3 float, ask_volume3 BIGINT, bid_price4 float, bid_volume4 BIGINT, ask_price4 float, ask_volume4 BIGINT, bid_price5 float, bid_volume5 BIGINT, ask_price5 float, ask_volume5 BIGINT, average_price float) TAGS(instrument_id BINARY(128), exchange_id BINARY(10), exchange_inst_id BINARY(64))")
	if err != err {
		err = fmt.Errorf("create stable failed:%w", err)
		return
	}

	return
}

func (db *DB) AddMarketData(data *ctp.CThostFtdcDepthMarketDataField) (err error) {
	var mData MarketData
	copier.Copy(&mData, data)
	now := time.Now()
	loc, _ := time.LoadLocation("Asia/Shanghai")
	date := now.Format("20060102")
	timeStr := fmt.Sprintf("%s %s.%03d", date, data.UpdateTime, data.UpdateMillisec)
	tm, err := time.ParseInLocation("20060102 15:04:05.000", timeStr, loc)
	if err != nil {
		log.Errorf("CtpExchange Parse MarketData timestamp %s failed %s", timeStr, err.Error())
	}
	mData.Ts = tm

	ch, ok := db.chs[mData.InstrumentID]
	if !ok {
		// createSQL := fmt.Sprintf(`CREATE TABLE %s using market TAGS("%s", "%s", "%s") if not exists`, data.InstrumentID, data.InstrumentID, data.ExchangeID, data.ExchangeInstID)
		// fmt.Println(createSQL)
		// _, err = db.db.Exec(createSQL)
		// if err != nil {
		// err = fmt.Errorf("create stable failed:%w", err)
		// return
		// }
		ch = make(chan *MarketData, db.cache)
		db.chs[mData.InstrumentID] = ch
		db.wg.Add(1)
		go db.loop(mData.InstrumentID, ch)
	}
	ch <- &mData

	return
}

func (db *DB) writeOnceToDB(tbl string, datas []*MarketData) (err error) {
	if len(datas) == 0 {
		return
	}
	logrus.Info("writeOnce:", tbl, len(datas))
	sql := fmt.Sprintf("insert into %s.%s using %s.market TAGS('%s', '%s', '%s') VALUES(:ts, :trading_day, :action_day, :last_price, :pre_settlement_price, :pre_close_price, :pre_open_interest, :open_price, :highest_price, :lowest_price, :volume, :turnover, :open_interest, :close_price, :settlement_price, :upper_limit_price, :lower_limit_price, :pre_delta, :curr_delta, :bid_price1, :bid_volume1, :ask_price1, :ask_volume1, :bid_price2, :bid_volume2, :ask_price2, :ask_volume2, :bid_price3, :bid_volume3, :ask_price3, :ask_volume3, :bid_price4, :bid_volume4, :ask_price4, :ask_volume4, :bid_price5, :bid_volume5, :ask_price5, :ask_volume5, :average_price)", db.database, tbl, db.database, datas[0].InstrumentID, datas[0].ExchangeID, datas[0].ExchangeInstID)
	// fmt.Println(sql)
	// sql := fmt.Sprintf("insert into %s using market TAGS(:InstrumentID, :ExchangeID, :ExchangeInstID) (ts, trading_day, action_day, last_price, pre_settlement_price, pre_close_price, pre_open_interest, open_price, highest_price, lowest_price, volume, turnover, open_interest, close_price, settlement_price, upper_limit_price, lower_limit_price, pre_delta, curr_delta, bid_price1, bid_volume1, ask_price1, ask_volume1, bid_price2, bid_volume2, ask_price2, ask_volume2, bid_price3, bid_volume3, ask_price3, ask_volume3, bid_price4, bid_volume4, ask_price4, ask_volume4, bid_price5, bid_volume5, ask_price5, ask_volume5, average_price) VALUES(:ts, :TradingDay, :ActionDay, :LastPrice, :PreSettlementPrice, :PreClosePrice, :PreOpenInterest, :OpenPrice, :HighestPrice, :LowestPrice, :Volume, :Turnover, :OpenInterest, :ClosePrice, :SettlementPrice, :UpperLimitPrice, :LowerLimitPrice, :PreDelta, :CurrDelta, :BidPrice1, :BidVolume1, :AskPrice1, :AskVolume1, :BidPrice2, :BidVolume2, :AskPrice2, :AskVolume2, :BidPrice3, :BidVolume3, :AskPrice3, :AskVolume3, :BidPrice4, :BidVolume4, :AskPrice4, :AskVolume4, :BidPrice5, :BidVolume5, :AskPrice5, :AskVolume5, :AveragePrice)", tbl)
	// _, err = db.db.Exec("use ?", db.database)
	// if err != nil {
	// 	logrus.Error("use database failed:", err.Error())
	// 	return
	// }
	_, err = db.db.NamedExec(sql, datas)
	return
}

func (db *DB) loop(tbl string, ch chan *MarketData) {
	defer db.wg.Done()
	var err error
	batch := make([]*MarketData, db.writeOnce)
	i := 0
	for v := range ch {
		batch[i] = v
		i++
		if i >= len(batch) {
			err = db.writeOnceToDB(tbl, batch)
			if err != nil {
				logrus.Errorf("db write %s failed: %s", tbl, err.Error())
			}
			i = 0
		}
	}
	if i > 0 {
		err = db.writeOnceToDB(tbl, batch[0:i])
		if err != nil {
			logrus.Errorf("finish db write %s failed: %s", tbl, err.Error())
		}
	}
	logrus.Infof("%s loop finished", tbl)
}

func (db *DB) GetDatas(symbol string, start, end time.Time) (datas []MarketData, err error) {
	rows, err := db.db.Queryx(fmt.Sprintf(`SELECT * FROM %s.market WHERE instrument_id='?' and ts>=? and ts <=?`, db.database), symbol, start, end)
	if err != nil {
		return
	}
	for rows.Next() {
		var data MarketData
		err = rows.StructScan(&data)
		if err != nil {
			return
		}
		datas = append(datas, data)
	}
	return
}

func (db *DB) Close() error {
	logrus.Info("begin to close")
	db.mutex.Lock()
	for _, v := range db.chs {
		close(v)
	}
	db.closed = true
	db.mutex.Unlock()
	db.wg.Wait()
	return db.db.Close()
}
