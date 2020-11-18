package operator

import (
	"github.com/jackc/pgtype"
	shopspring "github.com/jackc/pgtype/ext/shopspring-numeric"
)

// Member
type Member struct {
	MemberID           int64
	Email              string
	IP                 string
	FirstName          string
	LastName           string
	FamilyStatus       string
	MaidenName         string
	Citizenship        string
	Country            string
	City               string
	Zip                string
	Address1           string
	Address2           string
	StreetNumber       string
	StreetName         string
	Image              string
	Birthday           pgtype.Timestamptz
	EmailNotifications bool
	Phone              string
	Created            int64
	Role               int
	PasswordHash       string `json:"-"`
	Gender             string
	USD                shopspring.Numeric
	EUR                shopspring.Numeric
	LeverageAllowed    shopspring.Numeric
	StopLossAllowed    shopspring.Numeric //Maximum allowed StopLoss % for this member
	TakeProfitAllowed  shopspring.Numeric //Maximum allowed TakeProfit % for this member
	Status             string
}

type Deposit struct {
	DepositID   int64
	MemberID    int64
	MemberEmail string
	Amount      shopspring.Numeric
	Status      string
}

type Withdrawal struct {
	WithdrawalID int64
	MemberID     int64
	MemberEmail  string
	Amount       shopspring.Numeric
	Status       string
}

//Asset
type Asset struct {
	AssetID         int64
	MarketID        int64
	Ticker          string
	TickerTV        string
	Title           string
	Description     string
	Icon            string
	BuySpread       shopspring.Numeric
	SellSpread      shopspring.Numeric
	RateBuy         shopspring.Numeric
	RateSell        shopspring.Numeric
	Sentiment       int
	SentimentType   string
	Tradable        bool
	Position        int64
	Active          bool
	Rate            shopspring.Numeric
	Change          shopspring.Numeric
	Updated         int64
	Performance     []Rate
	DecimalScale    int32
	Currency        string
	BaseCurrency    string
	PipDecimals     int
	LeverageAllowed shopspring.Numeric
	TVWidget        bool
	FcsID           int64 `json:"-"`
}

//Trade
type Trade struct {
	TradeID        int64 //TradeID of existing order
	MemberID       int64
	AssetID        int64
	Type           string             //l/m limit/market
	Action         string             //b/s buy/sell
	MemberRate     shopspring.Numeric //Member defined asset rate in case of limit order
	MarketRate     shopspring.Numeric //Current market rate of asset
	RateEntry      shopspring.Numeric //Rate for member with buy/sell % fee at order creation time
	RateClosed     shopspring.Numeric //Rate for member with buy/sell % fee at order closure time
	Qty            shopspring.Numeric //How much asset qty / forex lots member wants to purchase
	OnePip         shopspring.Numeric //One forex pip (0.0001 for all Forex lots, 0.01 for JPY)
	PipsRateEntry  shopspring.Numeric //How much pips Forex lot rate had at order creation time
	PipsRateClosed shopspring.Numeric //How much pips Forex lot rate had at order closure time
	PipValue       shopspring.Numeric //How much money does cost 1 forex pip
	ForexAmount    shopspring.Numeric //How much Forex lot member bought / sold (for display in dashboard)
	Leverage       shopspring.Numeric //Leverage (example: 1x, 10x) with which asset is being purchased
	TotalReal      shopspring.Numeric //Real total order market cost (without applying leverage)
	DateOpen       pgtype.Timestamptz
	DateClosed     pgtype.Timestamptz
	Total          shopspring.Numeric //Total order cost for member balance (leverage applied)
	BalanceAsset   shopspring.Numeric //How much asset member has on his balance
	BalanceEntry   shopspring.Numeric //Member USD/EUR balance at order placement time
	BalanceClosed  shopspring.Numeric //Member USD/EUR balance after order is closed
	StopLoss       shopspring.Numeric //Stop loss %
	TakeProfit     shopspring.Numeric //Take profit %
	Profit         shopspring.Numeric //Order total profit that member earned (or lost)
	ProfitAbs      shopspring.Numeric //Absolute (no negative sign) profit value
	ProfitNegative bool               //Order total profit is less than 0
	Gain           shopspring.Numeric //How much % profit asset gained (or lost) since order creation
	ClosedBySystem bool               //Order was closed by system because of stop loss / take profit
	Status         string             //Order status on placement - pending (=> cancelled) => open => closed
	Timestamp      int64              //UNIX timestamp
}

//History
type History struct {
	HistoryID      int64
	MemberID       int64
	AssetID        int64
	TradeID        int64
	Type           string
	Action         string
	Currency       string
	Qty            shopspring.Numeric
	Rate           shopspring.Numeric
	Leverage       shopspring.Numeric
	Profit         shopspring.Numeric
	ProfitAbs      shopspring.Numeric //Absolute (no negative sign) profit value
	ProfitNegative bool               //History record profit is less than 0
	Status         string
	Address        string
	Created        pgtype.Timestamptz
	Timestamp      int64
}

//Rate
type Rate struct {
	RateID    int64
	AssetID   int64
	Rate      float64
	Timestamp int64
	Datetime  pgtype.Timestamptz
}

// News
type News struct {
	NewsID  int64
	Title   string
	Content string
	URL     string
	Date    pgtype.Timestamptz
}
