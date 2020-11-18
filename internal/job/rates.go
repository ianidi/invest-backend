package job

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/ianidi/exchange-server/graph/model"
	"github.com/ianidi/exchange-server/internal/db"
	"github.com/ianidi/exchange-server/internal/models"
	"github.com/ianidi/exchange-server/internal/redis"
	"github.com/ianidi/exchange-server/internal/trade"
	shopspring "github.com/jackc/pgtype/ext/shopspring-numeric"
	jsoniter "github.com/json-iterator/go"
	"github.com/parnurzeal/gorequest"
	"github.com/shopspring/decimal"
	"github.com/spf13/cast"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

type RatesJob struct {
}

//SleepTime how often to run the job
func (RatesJob) SleepTime() time.Duration {
	return time.Second * 20
}

const (
	APICryptonator = "https://api.cryptonator.com" //APICryptonator represents cryptonator.com API endpoint URL
	APIIEX         = "https://cloud.iexapis.com/stable"
	APIFCS         = "https://fcsapi.com/api-v2"
)

type Rate struct {
	Settings   models.Settings    //System settings
	Asset      models.Asset       //Asset information
	MarketID   int64              //Asset MarketID
	Interval   int64              //How often to update asset rate (seconds)
	Ticker     string             //Asset ticker (for Forex)
	RateString string             //Rate srting from API
	Rate       decimal.Decimal    //Rate decimal
	RateBuy    decimal.Decimal    //Rate buy (rate with spread)
	RateSell   decimal.Decimal    //Rate sell (rate with spread)
	DayAgoRate shopspring.Numeric //Asset rate 24H ago
	Change     decimal.Decimal    //24H change (%)
	Timestamp  int64              //UNIX timestamp
	FcsID      string
}

func (RatesJob) Run() {
	db := db.GetDB()

	var rate Rate
	var err error

	rate.Settings, err = rate.QuerySettings()
	if err != nil {
		return
	}

	rate.Timestamp = time.Now().Unix()

	var job []models.Job

	if err := db.Select(&job, "SELECT * FROM Job WHERE (Timestamp + Interval) < $1 AND Active=$2", rate.Timestamp, true); err != nil {
		return
	}

	for _, jobRow := range job {

		rate.Timestamp = time.Now().Unix()
		var err error

		rate.Interval = jobRow.Interval
		rate.MarketID = jobRow.MarketID

		//All assets except Forex
		if jobRow.Title == "rates_iex" {
			err = rate.IexUpdate()
		}

		if jobRow.Title == "rates" {
			err = rate.FcsUpdate()
		}

		//Update the last completion time if job succeeded
		if err == nil {
			tx := db.MustBegin()
			tx.Exec("UPDATE Job SET Timestamp=$1 WHERE JobID=$2", rate.Timestamp, jobRow.JobID)
			tx.Commit()
		}

	}

}

func (rate Rate) IexUpdate() error {
	db := db.GetDB()

	var err error

	rate.Timestamp = time.Now().Unix()

	var asset []models.Asset

	//Assets are updated each Interval seconds
	if err := db.Select(&asset, "SELECT * FROM Asset WHERE Updated<$1 AND Tradable=$2 AND Active=$3 AND MarketID=$4", rate.Timestamp-rate.Interval, true, true, rate.MarketID); err != nil {
		return err
	}

	for _, assetRow := range asset {

		rate.Asset = assetRow

		rate.RateString, err = rate.QueryRateStringFromAPI()
		if err == nil {
			err = rate.Update()
			if err != nil {
				fmt.Println("DefaultUpdate error", rate.Asset.Ticker, rate.RateString, err)
			}
		}

	}

	return nil
}

func (rate Rate) DefaultUpdateOld() error {
	db := db.GetDB()

	var err error

	rate.Timestamp = time.Now().Unix()

	var asset []models.Asset

	//Assets are updated each Interval seconds
	if err := db.Select(&asset, "SELECT * FROM Asset WHERE Updated<$1 AND Tradable=$2 AND Active=$3 AND MarketID=$4", rate.Timestamp-rate.Interval, true, true, rate.MarketID); err != nil {
		return err
	}

	for _, assetRow := range asset {

		rate.Asset = assetRow

		rate.RateString, err = rate.QueryRateStringFromAPI()
		if err == nil {
			err = rate.Update()
			if err != nil {
				fmt.Println("DefaultUpdate error", rate.Asset.Ticker, rate.RateString, err)
			}
		}

	}

	return nil
}

func (rate Rate) QueryRateStringFromAPI() (string, error) {

	var err error
	var rateString string

	if rate.Asset.MarketID == 1 {
		//Crypto
		rateString, err = rate.QueryCryptonator()
		if err != nil {
			return rateString, err
		}
	} else if rate.Asset.MarketID == 2 || rate.Asset.MarketID == 4 || rate.Asset.MarketID == 6 {
		//Stock
		rateString, err = rate.QueryStockIex()
		if err != nil {
			return rateString, err
		}
	} else if rate.Asset.MarketID == 5 {
		// Commodity
		rateString, err = rate.QueryCommodity()
		if err != nil {
			return rateString, err
		}

	}

	return rateString, nil
}

func (rate Rate) FcsUpdate() error {

	var err error
	var res FCSStockRes

	//Crypto
	if rate.MarketID == 1 {
		res, err = rate.QueryCrypto()
	}

	//Crypto
	if rate.MarketID == 2 || rate.MarketID == 4 || rate.MarketID == 6 {
		res, err = rate.QueryStock()
	}

	//Forex
	if rate.MarketID == 3 {
		res, err = rate.QueryForex()
	}

	//Indices
	if rate.MarketID == 7 {
		res, err = rate.QueryIndices()
	}

	if err != nil {
		return err
	}

	for _, resRow := range res.Response {

		// rate.Asset.Ticker = resRow.Symbol
		rate.FcsID = resRow.ID
		rate.RateString = cast.ToString(resRow.Price)

		// fmt.Println(rate.FcsID, rate.RateString)
		go rate.Update()
	}

	return nil
}

func (rate Rate) Update() error {
	db := db.GetDB()

	var err error

	// fmt.Println(rate.RateString, rate.FcsID, rate.Asset.Ticker)

	rate.Timestamp = time.Now().Unix()

	rate.Rate, err = decimal.NewFromString(rate.RateString)
	if err != nil {

		return err
	}

	//If new rate is 0, don't update this asset
	if rate.Rate.IsZero() {
		return errors.New("RATE_IS_ZERO")
	}

	if rate.FcsID != "" {
		rate.Asset, err = rate.QueryAssetByFcsID()
		if err != nil {

			return err
		}
	} else {
		rate.Asset, err = rate.QueryAssetByTicker()
		if err != nil {

			return err
		}
	}

	//Rate cut down to asset decimals
	rate.Rate, err = decimal.NewFromString(rate.Rate.StringFixed(rate.Asset.DecimalScale))

	var count int

	//Record rate every 3 hours
	if err := db.Get(&count, "SELECT count(*) FROM Rate WHERE AssetID=$1 AND Timestamp>$2", rate.Asset.AssetID, rate.Timestamp-10800); err != nil {
		if err != sql.ErrNoRows {

			return err
		}
	}

	if count == 0 {
		tx := db.MustBegin()
		tx.Exec("INSERT INTO Rate (AssetID, Rate, Timestamp) VALUES ($1, $2, $3)", rate.Asset.AssetID, rate.Rate, rate.Timestamp)
		//Clean up entries older than 24h ago
		//TODO: tx.Exec("DELETE * FROM Rate WHERE AssetID=$1 AND Timestamp<$2", rate.Asset.AssetID, rate.Timestamp-86400)
		tx.Commit()
	}

	if err := db.Get(&rate.DayAgoRate, "SELECT Rate FROM Rate WHERE AssetID=$1 AND Timestamp>$2 ORDER BY RateID ASC LIMIT 1", rate.Asset.AssetID, rate.Timestamp-86400); err != nil {
		if err != sql.ErrNoRows {

			return err
		}
	}

	//If no rate data is found
	if rate.DayAgoRate.Decimal.IsZero() {
		rate.DayAgoRate.Decimal = rate.Rate
	}

	//24h change = (100 * rate / dayAgoRate) - 100
	rate.Change = rate.Rate.Mul(decimal.NewFromInt(100)).Div(rate.DayAgoRate.Decimal).Sub(decimal.NewFromInt(100))

	//Forex spread has different formula (counted in pips e.g. 0.0005)
	if rate.Asset.MarketID == 3 {
		rate.RateBuy = rate.Rate.Add(rate.Asset.BuySpread.Decimal)
		rate.RateSell = rate.Rate.Sub(rate.Asset.SellSpread.Decimal)
	} else {
		rate.RateBuy = rate.Rate.Add(rate.Rate.Mul(rate.Asset.BuySpread.Decimal.Div(decimal.NewFromInt(100))))
		rate.RateSell = rate.Rate.Sub(rate.Rate.Mul(rate.Asset.SellSpread.Decimal.Div(decimal.NewFromInt(100))))
	}

	tx := db.MustBegin()
	tx.MustExec("UPDATE Asset SET Rate=$1, RateBuy=$2, RateSell=$3, Change=$4, Updated=$5 WHERE AssetID=$6", rate.Rate, rate.RateBuy, rate.RateSell, rate.Change.StringFixed(2), rate.Timestamp, rate.Asset.AssetID)
	tx.Commit()

	//Rate didn't change, stop further update
	// if rate.Rate == rate.Asset.Rate.Decimal {
	// 	fmt.Println("no update")
	// 	return nil
	// }

	//ws notify about asset rate update
	ms := model.Info{
		Event:         "rate",
		ID:            int(rate.Asset.AssetID),
		Rate:          rate.Rate.String(),
		RateBuy:       rate.RateBuy.String(),
		RateSell:      rate.RateSell.String(),
		Change:        rate.Change.StringFixed(2),
		Sentiment:     rate.Asset.Sentiment,
		SentimentType: rate.Asset.SentimentType,
	}

	msg, err := json.Marshal(ms)

	if err == nil {
		ch := redis.Channel{
			Name:    "info",
			Message: string(msg),
		}
		ch.PubToChannel()
	}
	//ws

	//Open pending limit orders that meet rate requirements
	rate.OpenPendingLimitOrders()

	//Update price alerts that meet rate requirements
	rate.UpdateAlert()

	//Update profit of each pending or open order with current AssetID and close orders that meet stop loss / take profit requirements
	rate.UpdateOrders()

	return nil

}

func (rate Rate) QueryAssetByFcsID() (models.Asset, error) {
	db := db.GetDB()

	var asset models.Asset

	err := db.Get(&asset, "SELECT * FROM Asset WHERE FcsID=$1 AND MarketID=$2", rate.FcsID, rate.MarketID)

	if err != nil {
		if err == sql.ErrNoRows {
			return asset, errors.New("INVALID_ASSET_BY_FCS_ID")
		}
		return asset, err
	}

	return asset, nil
}

func (rate Rate) QueryAssetByTicker() (models.Asset, error) {
	db := db.GetDB()

	var asset models.Asset

	err := db.Get(&asset, "SELECT * FROM Asset WHERE Ticker=$1", rate.Asset.Ticker)

	if err != nil {
		if err == sql.ErrNoRows {
			return asset, errors.New("INVALID_ASSET_BY_TICKER")
		}
		return asset, err
	}

	return asset, nil
}

//Update profit of each pending or open order with current AssetID and close orders that meet stop loss / take profit requirements
func (rate Rate) UpdateOrders() error {
	db := db.GetDB()

	var orders []trade.Order

	//Find pending or open orders with current AssetID
	if err := db.Select(&orders, "SELECT * FROM Trade WHERE (Status=$1 OR Status=$2) AND AssetID=$3", trade.STATUS_PENDING, trade.STATUS_OPEN, rate.Asset.AssetID); err != nil {
		return err
	}

	for _, orderRow := range orders {

		var err error

		order := orderRow
		order.Asset = rate.Asset
		order.Member, err = order.QueryMember()

		if err == nil {
			//CalculateProfit returns order struct with profit calculation and calls a function to record it to database
			order, err = order.CalculateProfit()
			if err == nil {
				//Close order that meets stop loss / take profit requirements
				if order.Status == trade.STATUS_OPEN {
					order.CloseSLTP()
				}
			}
		}

	}

	return nil
}

//Open pending limit orders that meet rate requirements
func (rate Rate) OpenPendingLimitOrders() error {
	db := db.GetDB()

	var orders []models.Trade

	//Find pending limit orders
	if err := db.Select(&orders, "SELECT * FROM Trade WHERE Type=$1 AND Status=$2 AND AssetID=$3 AND ((Action=$4 AND RateEntry>=$5) OR (Action=$6 AND RateEntry<=$5))", trade.ORDER_LIMIT, trade.STATUS_PENDING, rate.Asset.AssetID, trade.ACTION_BUY, rate.Rate, trade.ACTION_SELL); err != nil {
		return err
	}

	for _, orderRow := range orders {
		tx := db.MustBegin()
		if orderRow.Action == trade.ACTION_BUY {
			tx.MustExec("UPDATE Wallet SET Balance=Balance+$1 WHERE MemberID=$2 AND AssetID=$3", orderRow.Qty.Decimal, orderRow.MemberID, orderRow.AssetID)
		}
		tx.MustExec("UPDATE Trade SET Status=$1 WHERE TradeID=$2", trade.STATUS_OPEN, orderRow.TradeID)
		tx.Commit()

		//ws notify
		ms := model.Info{
			MemberID: int(orderRow.MemberID),
			Event:    "trade",
			Value:    "limit",
		}

		msg, err := json.Marshal(ms)

		if err == nil {
			ch := redis.Channel{
				Name:    "info",
				Message: string(msg),
			}
			ch.PubToChannel()
		}
		//ws
	}

	return nil
}

//Update price alerts that meet rate requirements
func (rate Rate) UpdateAlert() error {
	db := db.GetDB()

	var alert []models.Alert

	//Find alert with this asset
	if err := db.Select(&alert, "SELECT * FROM Alert WHERE AssetID=$1 AND Status=$2 AND ((Price>=$3 AND Type=$4) OR (Price<=$3 AND Type=$5))", rate.Asset.AssetID, false, rate.Rate, "l", "h"); err != nil {
		return err
	}

	for _, alertRow := range alert {
		tx := db.MustBegin()
		tx.MustExec("UPDATE Alert SET Status=$1, Datetime=CURRENT_TIMESTAMP WHERE AlertID=$2", true, alertRow.AlertID)
		tx.Commit()

		//ws notify
		ms := model.Info{
			MemberID: int(alertRow.MemberID),
			Event:    "alert",
		}

		msg, err := json.Marshal(ms)

		if err == nil {
			ch := redis.Channel{
				Name:    "info",
				Message: string(msg),
			}
			ch.PubToChannel()
		}
		//ws
	}

	return nil
}

//QueryCryptonator get cryptocurrency ticker price rate
func (rate Rate) QueryCryptonator() (string, error) {
	request := gorequest.New()

	var res CryptonatorRes

	request.Get(fmt.Sprintf(`%s/api/ticker/%s-usd`, APICryptonator, strings.ToLower(rate.Asset.Ticker))).
		Retry(3, 2*time.Second, http.StatusBadRequest, http.StatusInternalServerError).
		Set("Accept", "application/json").
		Set("Content-Type", "application/json").
		EndStruct(&res)

	if res.Success != true {
		return "", errors.New("CRYPTONATOR_SERVICE_UNAVAILABLE, ticker: " + rate.Asset.Ticker)
	}

	return res.Ticker.Price, nil
}

//QueryStockIex get stock asset rate from iex
func (rate Rate) QueryStockIex() (string, error) {
	request := gorequest.New()

	var res IEXStockRes

	request.Get(fmt.Sprintf(`%s/stock/%s/quote?token=%s`, APIIEX, strings.ToLower(rate.Asset.Ticker), rate.Settings.APIKeyIEX)).
		Retry(3, 2*time.Second, http.StatusBadRequest, http.StatusInternalServerError).
		Set("Accept", "application/json").
		Set("Content-Type", "application/json").
		EndStruct(&res)

	// TODO: Success != true

	return cast.ToString(res.LatestPrice), nil
}

//QueryForex get forex rates from fcs
func (rate Rate) QueryForex() (FCSStockRes, error) {
	request := gorequest.New()

	var res FCSStockRes
	request.Get(fmt.Sprintf(`%s/forex/latest?id=1,13,18,19,20,21,39&access_key=%s`, APIFCS, rate.Settings.APIKeyFCS)).
		Retry(3, 2*time.Second, http.StatusBadRequest, http.StatusInternalServerError).
		Set("Accept", "application/json").
		Set("Content-Type", "application/json").
		EndStruct(&res)

	if res.Status != true {
		return res, errors.New("FCS_QUERY_ERROR")
	}

	return res, nil
}

//QueryCrypto get forex rates from fcs
func (rate Rate) QueryCrypto() (FCSStockRes, error) {
	request := gorequest.New()

	var res FCSStockRes
	request.Get(fmt.Sprintf(`%s/crypto/latest?id=78,79,81,82,2054,2160,2485,3359,3498,3739,4793,5192,7135&access_key=%s`, APIFCS, rate.Settings.APIKeyFCS)).
		Retry(3, 2*time.Second, http.StatusBadRequest, http.StatusInternalServerError).
		Set("Accept", "application/json").
		Set("Content-Type", "application/json").
		EndStruct(&res)

	if res.Status != true {
		return res, errors.New("FCS_QUERY_ERROR")
	}

	return res, nil
}

//QueryStock get stock rates from fcs
func (rate Rate) QueryStock() (FCSStockRes, error) {
	request := gorequest.New()

	var res FCSStockRes
	request.Get(fmt.Sprintf(`%s/stock/latest?id=4,5,8,15,20,26,27,31,32,33,35,36,37,38,39,40,41,42,43,44,45,46,47,48,49,50,51,52,53,54,55,56,57,58,59,60,61,62,63,64,66,67,68,69,70,71,72,73,74,74,75,77,78,79,80,81,82,82,83,84,85,86,87,88,89,90,91,92,94,95,96,97,98,99,100,101,102,103,104,105,106,107,108,109,110,111,112,114,114,115,116,117,118,119,120,121,122,123,124,126,168,185,192,193,195,196,202,206,207,214,215,217,219,222,283,359,488,534,583,595,635,646,802,806,862,876,906,927,934,937,944,971,1002,1006,1031,1047,1142,1145,1146,1147,1154,1165,1210,1225,1231,1314,1324,1333,1357,1386,1409,1425,1440,1443,1458,1475,1478,1479,1481,1931,1978,1980,2013,2086,2094,2124,2260,2286,2433,2545,2555,2602,2625,4125,4127,4128,4129,4130,4134,4135,4138,4139,4140,4142,4143,4144,4146,4150,4153,4155,4653,4852,4854,4856,4863,4873,4881,4884,4902,4918,4920,4934,4935,4947,4949,4951,4953,4960,4965,4968,4971,4975,4976,4979,4984,5002,5354,5408,5615,5681,5683,5688,5715,6206,6207,6608,8071,12502,9957697,9957723,9958901,9958946,9958951,9958990,9959014,9961009,9961011,9961013,9961021,9961031,9961060,9961080,9961083,9961086,9961125,9961155,9963725,9963729,9963734&access_key=%s`, APIFCS, rate.Settings.APIKeyFCS)).
		Retry(3, 2*time.Second, http.StatusBadRequest, http.StatusInternalServerError).
		Set("Accept", "application/json").
		Set("Content-Type", "application/json").
		EndStruct(&res)

	if res.Status != true {
		return res, errors.New("FCS_QUERY_ERROR")
	}

	return res, nil
}

//QueryIndices get indices rates from fcs
func (rate Rate) QueryIndices() (FCSStockRes, error) {
	request := gorequest.New()

	var res FCSStockRes
	request.Get(fmt.Sprintf(`%s/stock/indices_latest?id=1,2,4,5,8,9,12,268,1226&access_key=%s`, APIFCS, rate.Settings.APIKeyFCS)).
		Retry(3, 2*time.Second, http.StatusBadRequest, http.StatusInternalServerError).
		Set("Accept", "application/json").
		Set("Content-Type", "application/json").
		EndStruct(&res)

	if res.Status != true {
		return res, errors.New("FCS_QUERY_ERROR")
	}

	return res, nil
}

//QueryForexIEX get forex rates from iex
func (rate Rate) QueryForexIEX() (IEXForexRes, error) {
	request := gorequest.New()

	var res IEXForexRes
	//,USDCNH,USDCZK,USDDKK,USDILS,USDINR,USDMXN,USDNOK,USDPLN,USDSEK,USDSGD,USDTHB,USDZAR,USDTRY,USDHKD
	request.Get(fmt.Sprintf(`%s/fx/latest?symbols=EURUSD,USDJPY,GBPUSD,USDCHF,AUDUSD,USDCAD,NZDUSD,GBPEUR,EURCHF,EURJPY&token=%s`, APIIEX, rate.Settings.APIKeyIEX)).
		Retry(3, 2*time.Second, http.StatusBadRequest, http.StatusInternalServerError).
		Set("Accept", "application/json").
		Set("Content-Type", "application/json").
		EndStruct(&res)

	// TODO: Success != true

	return res, nil
}

//QueryCommodity get commodity rate from iex
func (rate Rate) QueryCommodity() (string, error) {
	request := gorequest.New()

	var res IEXCommodityRes

	request.Get(fmt.Sprintf(`%s/time-series/energy/%s?token=%s`, APIIEX, strings.ToLower(rate.Asset.Ticker), rate.Settings.APIKeyIEX)).
		Retry(3, 2*time.Second, http.StatusBadRequest, http.StatusInternalServerError).
		Set("Accept", "application/json").
		Set("Content-Type", "application/json").
		EndStruct(&res)

		// TODO: Success != true

	var value float64

	for _, resRow := range res {
		value = resRow.Value
	}

	return cast.ToString(value), nil
}

func (rate Rate) QuerySettings() (models.Settings, error) {
	db := db.GetDB()

	var settings models.Settings

	err := db.Get(&settings, "SELECT * FROM Settings WHERE SettingsID=$1", 1)
	if err != nil {
		return settings, err
	}

	return settings, nil
}
