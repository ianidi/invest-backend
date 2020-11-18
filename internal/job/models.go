package job

type CryptonatorRes struct {
	Ticker struct {
		Base   string `json:"base"`
		Target string `json:"target"`
		Price  string `json:"price"`
		Volume string `json:"volume"`
		Change string `json:"change"`
	} `json:"ticker"`
	Timestamp int    `json:"timestamp"`
	Success   bool   `json:"success"`
	Error     string `json:"error"`
}

type IEXStockRes struct {
	Symbol                 string      `json:"symbol"`
	CompanyName            string      `json:"companyName"`
	PrimaryExchange        string      `json:"primaryExchange"`
	CalculationPrice       string      `json:"calculationPrice"`
	Open                   interface{} `json:"open"`
	OpenTime               interface{} `json:"openTime"`
	OpenSource             string      `json:"openSource"`
	Close                  interface{} `json:"close"`
	CloseTime              interface{} `json:"closeTime"`
	CloseSource            string      `json:"closeSource"`
	High                   interface{} `json:"high"`
	HighTime               int64       `json:"highTime"`
	HighSource             string      `json:"highSource"`
	Low                    interface{} `json:"low"`
	LowTime                int64       `json:"lowTime"`
	LowSource              string      `json:"lowSource"`
	LatestPrice            float64     `json:"latestPrice"`
	LatestSource           string      `json:"latestSource"`
	LatestTime             string      `json:"latestTime"`
	LatestUpdate           int64       `json:"latestUpdate"`
	LatestVolume           interface{} `json:"latestVolume"`
	IexRealtimePrice       int         `json:"iexRealtimePrice"`
	IexRealtimeSize        int         `json:"iexRealtimeSize"`
	IexLastUpdated         int         `json:"iexLastUpdated"`
	DelayedPrice           interface{} `json:"delayedPrice"`
	DelayedPriceTime       interface{} `json:"delayedPriceTime"`
	OddLotDelayedPrice     interface{} `json:"oddLotDelayedPrice"`
	OddLotDelayedPriceTime interface{} `json:"oddLotDelayedPriceTime"`
	ExtendedPrice          interface{} `json:"extendedPrice"`
	ExtendedChange         interface{} `json:"extendedChange"`
	ExtendedChangePercent  interface{} `json:"extendedChangePercent"`
	ExtendedPriceTime      interface{} `json:"extendedPriceTime"`
	PreviousClose          float64     `json:"previousClose"`
	PreviousVolume         int         `json:"previousVolume"`
	Change                 float64     `json:"change"`
	ChangePercent          float64     `json:"changePercent"`
	Volume                 interface{} `json:"volume"`
	IexMarketPercent       interface{} `json:"iexMarketPercent"`
	IexVolume              int         `json:"iexVolume"`
	AvgTotalVolume         int         `json:"avgTotalVolume"`
	IexBidPrice            int         `json:"iexBidPrice"`
	IexBidSize             int         `json:"iexBidSize"`
	IexAskPrice            int         `json:"iexAskPrice"`
	IexAskSize             int         `json:"iexAskSize"`
	IexOpen                interface{} `json:"iexOpen"`
	IexOpenTime            interface{} `json:"iexOpenTime"`
	IexClose               float64     `json:"iexClose"`
	IexCloseTime           int64       `json:"iexCloseTime"`
	MarketCap              int64       `json:"marketCap"`
	PeRatio                float64     `json:"peRatio"`
	Week52High             float64     `json:"week52High"`
	Week52Low              float64     `json:"week52Low"`
	YtdChange              float64     `json:"ytdChange"`
	LastTradeTime          int64       `json:"lastTradeTime"`
	IsUSMarketOpen         bool        `json:"isUSMarketOpen"`
}

type IEXForexRes []struct {
	Symbol    string  `json:"symbol"`
	Rate      float64 `json:"rate"`
	Timestamp int64   `json:"timestamp"`
	IsDerived bool    `json:"isDerived"`
}

type IEXCommodityRes []struct {
	Value   float64 `json:"value"`
	ID      string  `json:"id"`
	Source  string  `json:"source"`
	Key     string  `json:"key"`
	Subkey  string  `json:"subkey"`
	Date    int64   `json:"date"`
	Updated int64   `json:"updated"`
}

type FCSStockRes struct {
	Status   bool   `json:"status"`
	Code     int    `json:"code"`
	Msg      string `json:"msg"`
	Response []struct {
		ID          string `json:"id"`
		Price       string `json:"price"`
		Change      string `json:"change"`
		ChgPer      string `json:"chg_per"`
		LastChanged string `json:"last_changed"`
		Symbol      string `json:"symbol"`
	} `json:"response"`
	Info struct {
		ServerTime  string `json:"server_time"`
		CreditCount int    `json:"credit_count"`
	} `json:"info"`
}

type FCSIndicesRes struct {
	Status   bool   `json:"status"`
	Code     int    `json:"code"`
	Msg      string `json:"msg"`
	Response []struct {
		Price      string `json:"price"`
		High       string `json:"high"`
		Low        string `json:"low"`
		Chg        string `json:"chg"`
		ChgPercent string `json:"chg_percent"`
		DateTime   string `json:"dateTime"`
		ID         string `json:"id"`
		Name       string `json:"name"`
	} `json:"response"`
	Info struct {
		ServerTime  string `json:"server_time"`
		CreditCount int    `json:"credit_count"`
	} `json:"info"`
}
