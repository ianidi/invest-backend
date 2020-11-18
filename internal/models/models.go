package models

import (
	"github.com/jackc/pgtype"
	shopspring "github.com/jackc/pgtype/ext/shopspring-numeric"
)

// Member
type Member struct {
	MemberID           int64
	Email              string
	ManagerID          int64
	IP                 string `json:"-"`
	FirstName          string
	LastName           string
	Gender             string
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
	Created            int64  `json:"-"`
	Role               int64  `json:"-"`
	PasswordHash       string `json:"-"`
	CurrencyID         int64  `json:"-"`
	USD                shopspring.Numeric
	EUR                shopspring.Numeric
	LeverageAllowed    shopspring.Numeric
	StopLossAllowed    shopspring.Numeric //Maximum allowed StopLoss % for this member
	TakeProfitAllowed  shopspring.Numeric //Maximum allowed TakeProfit % for this member
	Status             string
	ManagerRole        string
}

// Verify - OTP verification
type Verify struct {
	VerifyID  int64
	MemberID  int64
	CodeHash  string //OTP code hash
	Method    string //Phone/email
	Action    string //Auth/reset/confirm
	Email     string //Email being verified
	EmailHash string //Email md5 hash
	Phone     string //Phone being verified
	Status    string //Status (pending, cancelled, success)
	Attempts  int    //Code entry attempts
	IP        string //IP address
	Created   int64  //UNIX timestamp
}

//Swagger error response
type Error struct {
	Status bool   `example:"false"`
	Error  string `example:"ERROR_MESSAGE"`
}

//Swagger success response
type Success struct {
	Status bool `example:"true"`
}

// Settings
type Settings struct {
	SettingsID                 int
	SMTPHost                   string //SMTP host
	SMTPUsername               string //SMTP username
	SMTPPassword               string //SMTP password
	SMTPPort                   int64  //SMTP port
	SMTPFromEmail              string //SMTP sender email
	SMTPFromName               string //SMTP sender name
	PlatformURL                string //Platform URL
	NewsURL                    string //News blog URL
	Title                      string //Platform title
	LeverageAllowedCrypto      shopspring.Numeric
	LeverageAllowedStock       shopspring.Numeric
	LeverageAllowedForex       shopspring.Numeric
	LeverageAllowedCommodities shopspring.Numeric
	LeverageAllowedIndices     shopspring.Numeric
	StopLossProtection         shopspring.Numeric //Default system setting (system max order loss protection)
	TakeProfitProtection       shopspring.Numeric //Default system setting (system max profit loss protection)
	StopLossAllowed            shopspring.Numeric //Default system setting (system max order loss default allowance)
	TakeProfitAllowed          shopspring.Numeric //Default system setting (system max order profit default allowance)
	APIKeyIEX                  string             //IEXCloud API key
	APIKeyFCS                  string             //FCS API key
	DefaultCurrencyID          int64              //Default currency for member balances
}

// News
type News struct {
	NewsID  int64
	Title   string
	Content string
	URL     string
	Date    pgtype.Timestamptz
}

//Wallet
type Wallet struct {
	WalletID int64 `json:"-"`
	MemberID int64 `json:"-"`
	AssetID  int64
	Balance  shopspring.Numeric
}

//History
type History struct {
	HistoryID      int64
	MemberID       int64 `json:"-"`
	AssetID        int64
	TradeID        int64 `json:"-"`
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
	Timestamp      int64 `json:"-"`
}

//Trade
type Trade struct {
	TradeID        int64 //TradeID of existing order
	MemberID       int64 `json:"-"`
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
	PipValue       shopspring.Numeric `json:"-"` //How much money does cost 1 forex pip
	ForexAmount    shopspring.Numeric `json:"-"` //How much Forex lot member bought / sold (for display in dashboard)
	Leverage       shopspring.Numeric //Leverage (example: 1x, 10x) with which asset is being purchased
	TotalReal      shopspring.Numeric `json:"-"` //Real total order market cost (without applying leverage)
	DateOpen       pgtype.Timestamptz
	DateClosed     pgtype.Timestamptz
	Total          shopspring.Numeric //Total order cost for member balance (leverage applied)
	BalanceAsset   shopspring.Numeric `json:"-"` //How much asset member has on his balance
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
	Timestamp      int64              `json:"-"` //UNIX timestamp
}

//Fave
type Fave struct {
	FaveID   int64 `json:"-"`
	MemberID int64 `json:"-"`
	AssetID  int64
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
	BuySpread       shopspring.Numeric `json:"-"`
	SellSpread      shopspring.Numeric `json:"-"`
	RateBuy         shopspring.Numeric
	RateSell        shopspring.Numeric
	Sentiment       int
	SentimentType   string
	Tradable        bool
	Position        int
	Active          bool `json:"-"`
	Rate            shopspring.Numeric
	Change          shopspring.Numeric
	Updated         int64 `json:"-"`
	Performance     []Rate
	DecimalScale    int32
	Currency        string `json:"-"`
	BaseCurrency    string `json:"-"`
	PipDecimals     int    `json:"-"`
	LeverageAllowed shopspring.Numeric
	TVWidget        bool
	FcsID           int64 `json:"-"`
}

//Rate
type Rate struct {
	RateID    int64 `json:"-"`
	AssetID   int64 `json:"-"`
	Rate      float64
	Timestamp int64 `json:"-"`
	Datetime  pgtype.Timestamptz
}

type Alert struct {
	AlertID       int64
	MemberID      int64 `json:"-"`
	AssetID       int64
	Type          string `json:"-"` //h / l. higher / lower
	Price         shopspring.Numeric
	Status        bool   `json:"-"`
	Currency      string `json:"-"`
	Datetime      pgtype.Timestamptz
	TimestampOpen int64 `json:"-"`
}

//Job
type Job struct {
	JobID     int64
	MarketID  int64
	Title     string
	Timestamp int64
	Interval  int64
	Active    bool
}

//Onboarding
type Onboarding struct {
	MemberID int64
	Email    bool
	Contract bool
	Phone    bool
	Password bool
	Complete bool
}

type Invest struct {
	InvestID    int64
	CategoryID  int64
	Title       string
	Subtitle    string
	Description string
}

type InvestOffer struct {
	Invest
	OfferID       int64
	CurrencyID    pgtype.Int8
	BankDetailsID int64
	Status        string
}

// Faq         []*Faq   `json:"FAQ"`
// Photo       []*Photo `json:"Photo"`

type Category struct {
	CategoryID int64
	Title      string
}

type Currency struct {
	CurrencyID pgtype.Int8
	Title      string
	Symbol     string
}

type Offer struct {
	OfferID       int64
	MemberID      int64
	InvestID      int64
	CurrencyID    pgtype.Int8
	BankDetailsID int64
	Title         string
	Status        string
}

type FAQ struct {
	FAQID    int64
	Question string
	Answer   string
	Position int64
}

type Contract struct {
	ContractID int64
	Title      string
	ContentRaw string
	Content    string
	OfferID    int64
	Current    bool
	Template   bool
}

type Interest struct {
	InterestID   int64
	OfferID      int64
	AmountFrom   shopspring.Numeric
	AmountTo     shopspring.Numeric
	DurationFrom shopspring.Numeric
	DurationTo   shopspring.Numeric
	Interest     shopspring.Numeric
}

type Deal struct {
	DealID            int64
	OfferID           int64
	ContractID        int64
	MemberID          int64
	CurrencyID        int64
	SignatureFilename string
	VerificationCode  string
	DateCreated       pgtype.Timestamptz
	DateVerified      pgtype.Timestamptz
	DateSigned        pgtype.Timestamptz
	DatePaid          pgtype.Timestamptz
	DateStart         pgtype.Timestamptz
	DateEnd           pgtype.Timestamptz
	Status            string
	Amount            shopspring.Numeric
	Duration          shopspring.Numeric
}

type Invoice struct {
	InvoiceID        int64
	OfferID          int64
	MemberID         int64
	DealID           int64
	CurrencyID       int64
	Amount           shopspring.Numeric
	Status           string
	DateCreated      pgtype.Timestamptz
	TimestampCreated int64
	DatePaid         pgtype.Timestamptz
	TimestampPaid    int64
}

type Upload struct {
	UploadID int64
	MemberID int64
	Filename string
	Category string
	Created  int64
}

type Media struct {
	MediaID  pgtype.Int8
	MemberID pgtype.Int8
	InvestID pgtype.Int8
	Title    pgtype.Varchar
	Position pgtype.Int8
	Filename pgtype.Varchar
	Category pgtype.Varchar
	Created  pgtype.Int8
}

type TX struct {
	TXID              int64
	MemberID          int64
	Amount            shopspring.Numeric
	AmountNegative    bool
	CurrencyID        pgtype.Int8
	Status            string
	DateCreated       pgtype.Timestamptz
	DateComplete      pgtype.Timestamptz
	TimestampCreated  int64
	TimestampComplete int64
}

type Balance struct {
	BalanceID      int64
	MemberID       int64
	CurrencyID     pgtype.Int8
	Amount         shopspring.Numeric
	AmountNegative bool
}

type BankDetails struct {
	BankDetailsID          int64
	Title                  string
	BeneficiaryCompany     string
	BeneficiaryFirstName   string
	BeneficiaryLastName    string
	BeneficiaryCountry     string
	BeneficiaryCity        string
	BeneficiaryZip         string
	BeneficiaryAddress     string
	BankName               string
	BankBranch             string
	BankIFSC               string
	BankBranchCountry      string
	BankBranchCity         string
	BankBranchZip          string
	BankBranchAddress      string
	BankAccountNumber      string
	BankAccountType        string
	BankRoutingNumber      string
	BankTransferCaption    string
	BankIBAN               string
	BankSWIFT              string
	BankSWIFTCorrespondent string
	BankBIC                string
}

// Lead
type Lead struct {
	LeadID           pgtype.Int8
	ManagerID        pgtype.Int8
	MemberID         pgtype.Int8
	CampaignID       pgtype.Int8
	CurrencyID       pgtype.Int8
	Email            pgtype.Varchar
	Phone            pgtype.Varchar
	IP               pgtype.Varchar
	FirstName        pgtype.Varchar
	LastName         pgtype.Varchar
	Gender           pgtype.Varchar
	FamilyStatus     pgtype.Varchar
	MaidenName       pgtype.Varchar
	Citizenship      pgtype.Varchar
	Country          pgtype.Varchar
	City             pgtype.Varchar
	Zip              pgtype.Varchar
	Address1         pgtype.Varchar
	Address2         pgtype.Varchar
	StreetNumber     pgtype.Varchar
	StreetName       pgtype.Varchar
	Birthday         pgtype.Timestamptz
	Status           pgtype.Varchar
	DateCreated      pgtype.Timestamptz
	TimestampCreated pgtype.Int8
}

type Comment struct {
	CommentID        pgtype.Int8
	MemberID         pgtype.Int8
	Content          pgtype.Text
	DateCreated      pgtype.Timestamptz
	TimestampCreated pgtype.Int8
	DateEdited       pgtype.Timestamptz
	TimestampEdited  pgtype.Int8
	LeadID           pgtype.Int8
}

type Checklist struct {
	ChecklistID      pgtype.Int8
	Title            pgtype.Varchar
	Complete         pgtype.Bool
	Position         pgtype.Int8
	DateCreated      pgtype.Timestamptz
	TimestampCreated pgtype.Int8
	LeadID           pgtype.Int8
}

type Appointment struct {
	AppointmentID    pgtype.Int8
	Type             pgtype.Varchar
	Title            pgtype.Varchar
	Description      pgtype.Text
	DateCreated      pgtype.Timestamptz
	TimestampCreated pgtype.Int8
	DateDue          pgtype.Timestamptz
	TimestampDue     pgtype.Int8
	Status           pgtype.Varchar
	LeadID           pgtype.Int8
}

type Campaign struct {
	CampaignID       pgtype.Int8
	Title            pgtype.Varchar
	Description      pgtype.Text
	DateCreated      pgtype.Timestamptz
	TimestampCreated pgtype.Int8
}

type Manager struct {
	ManagerID pgtype.Int8
	MemberID  pgtype.Varchar
	Role      pgtype.Text
}
