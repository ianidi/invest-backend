package portal

import (
	"context"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"hash"
	"strings"
	"time"

	"github.com/adrg/postcode"
	"github.com/cbroglie/mustache"
	"github.com/gin-gonic/gin"
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
	"github.com/google/uuid"
	"github.com/ianidi/exchange-server/graph/methods/constants"
	"github.com/ianidi/exchange-server/graph/model"
	"github.com/ianidi/exchange-server/internal/db"
	"github.com/ianidi/exchange-server/internal/jwt"
	"github.com/ianidi/exchange-server/internal/models"
	"github.com/ianidi/exchange-server/internal/utils"
	"github.com/jackc/pgtype"
	"github.com/jmoiron/sqlx"
	"github.com/muesli/crunchy"
	"github.com/nyaruka/phonenumbers"
	"github.com/sethvargo/go-password/password"
	"github.com/spf13/cast"
	"github.com/spf13/viper"
	"golang.org/x/crypto/bcrypt"
)

type Portal struct {
	c              *gin.Context
	Ctx            context.Context
	ProfileHandler *jwt.ProfileHandler
	Timestamp      int64 //Current UNIX timestamp
	Member         models.Member
	Invest         models.Invest
	Contract       models.Contract
	Offer          models.Offer
	Deal           models.Deal
	Invoice        models.Invoice
	Interest       models.Interest
	Settings       models.Settings

	Verify models.Verify

	Action           string
	Method           string
	Phone            string //Phone number
	Email            string //Email address
	EmailHash        string //Email address md5 hash
	Password         string //Current (auth) or new (reset) member password
	PasswordHash     string //Password hash
	Code             string //OTP code
	CodeHash         string //OTP code hash
	ConfirmationLink string //Email confirmation link
	Timeout          int64  //How frequently in seconds OTP codes can be requested

	InvoiceLink string //Invoice link for payment

	FirstName    string
	LastName     string
	Birthday     string
	Citizenship  string
	Gender       string
	FamilyStatus string
	MaidenName   string
	Country      string
	City         string
	Zip          string
	StreetNumber string
	StreetName   string
}

func (portal *Portal) GinContextFromContext() error {
	ginContext := portal.Ctx.Value("GinContextKey")
	if ginContext == nil {
		err := fmt.Errorf("could not retrieve gin.Context")
		return err
	}

	gc, ok := ginContext.(*gin.Context)
	if !ok {
		err := fmt.Errorf("gin.Context has wrong type")
		return err
	}

	portal.c = gc

	return nil
}

func (portal *Portal) GetMember() error {

	db := db.GetDB()

	var err error

	err = portal.GinContextFromContext()
	if err != nil {
		return err
	}

	h := jwt.Service

	tokenString := jwt.ExtractToken(portal.c.Request)

	metadata, err := h.TK.TokenMetadata(tokenString)
	if err != nil {
		return err
	}

	MemberID := metadata.UserId

	var member models.Member

	if err := db.Get(&member, "SELECT * FROM Member WHERE MemberID=$1", MemberID); err != nil {
		if err == sql.ErrNoRows {
			return errors.New("NO_MEMBER_RECORD")
		}
		return err
	}

	portal.Member = member

	return nil
}

func (portal *Portal) ReplyMember() (*model.Member, error) {
	var err error

	if portal.Member.CurrencyID == 0 {
		err = portal.QuerySettings()
		if err != nil {
			return nil, err
		}

		portal.Member.CurrencyID = portal.Settings.DefaultCurrencyID
	}

	MemberID := cast.ToInt(portal.Member.MemberID)
	Email := portal.Member.Email
	ManagerID := cast.ToInt(portal.Member.ManagerID)
	IP := portal.Member.IP
	FirstName := portal.Member.FirstName
	LastName := portal.Member.LastName
	Gender := portal.Member.Gender
	FamilyStatus := portal.Member.FamilyStatus
	MaidenName := portal.Member.MaidenName
	Citizenship := portal.Member.Citizenship
	Country := portal.Member.Country
	City := portal.Member.City
	Zip := portal.Member.Zip
	Address1 := portal.Member.Address1
	Address2 := portal.Member.Address2
	StreetNumber := portal.Member.StreetNumber
	StreetName := portal.Member.StreetName
	Image := viper.GetString("s3_cdn_url") + portal.Member.Image
	Birthday := portal.Member.Birthday.Time.String()
	EmailNotifications := portal.Member.EmailNotifications
	Phone := portal.Member.Phone
	Created := cast.ToInt(portal.Member.Created)
	Role := cast.ToInt(portal.Member.Role)
	CurrencyID := cast.ToInt(portal.Member.CurrencyID)
	USD := portal.Member.USD.Decimal.String()
	EUR := portal.Member.EUR.Decimal.String()
	LeverageAllowed := portal.Member.LeverageAllowed.Decimal.String()
	StopLossAllowed := portal.Member.StopLossAllowed.Decimal.String()
	TakeProfitAllowed := portal.Member.TakeProfitAllowed.Decimal.String()
	Status := portal.Member.Status
	ManagerRole := portal.Member.ManagerRole

	res := &model.Member{
		MemberID:           MemberID,
		Email:              &Email,
		ManagerID:          &ManagerID,
		IP:                 &IP,
		FirstName:          &FirstName,
		LastName:           &LastName,
		Gender:             &Gender,
		FamilyStatus:       &FamilyStatus,
		MaidenName:         &MaidenName,
		Citizenship:        &Citizenship,
		Country:            &Country,
		City:               &City,
		Zip:                &Zip,
		Address1:           &Address1,
		Address2:           &Address2,
		StreetNumber:       &StreetNumber,
		StreetName:         &StreetName,
		Image:              &Image,
		Birthday:           &Birthday,
		EmailNotifications: &EmailNotifications,
		Phone:              &Phone,
		Created:            &Created,
		Role:               &Role,
		CurrencyID:         &CurrencyID,
		Usd:                &USD,
		Eur:                &EUR,
		LeverageAllowed:    &LeverageAllowed,
		StopLossAllowed:    &StopLossAllowed,
		TakeProfitAllowed:  &TakeProfitAllowed,
		Status:             &Status,
		ManagerRole:        &ManagerRole,
	}

	return res, nil
}

func (portal *Portal) MemberPersonalUpdate(input model.MemberPersonalUpdateRequest) error {
	db := db.GetDB()

	tx := db.MustBegin()
	tx.MustExec("UPDATE Member SET FirstName=$1, LastName=$2 WHERE MemberID=$3", input.FirstName, input.LastName, portal.Member.MemberID)
	tx.Commit()

	return nil
}

func (portal *Portal) MemberPhoneUpdate() error {
	db := db.GetDB()

	tx := db.MustBegin()
	tx.MustExec("UPDATE Member SET Phone=$1 WHERE MemberID=$2", portal.Phone, portal.Member.MemberID)
	tx.Commit()

	return nil
}

func (portal *Portal) MemberEmailUpdate() error {
	db := db.GetDB()

	tx := db.MustBegin()
	tx.MustExec("UPDATE Member SET Email=$1 WHERE MemberID=$2", portal.Email, portal.Member.MemberID)
	tx.Commit()

	return nil
}

func (portal *Portal) GetMemberIDFromToken(token string) (string, error) {
	var err error

	err = portal.GinContextFromContext()
	if err != nil {
		return "", err
	}

	err = jwt.TokenValidNoExtraction(token)
	if err != nil {
		return "", err
	}

	h := jwt.Service

	metadata, err := h.TK.TokenMetadata(token)
	if err != nil {
		return "", err
	}
	// userId, err := h.RD.FetchAuth(metadata.TokenUuid)
	// if err != nil {
	// 	return "", err
	// }

	MemberID := metadata.UserId

	portal.c.Set(jwt.MemberKey, MemberID)

	return MemberID, nil
}

func (portal *Portal) QueryFAQArray(IDs []pgtype.Int8) ([]*model.Faq, error) {
	if len(IDs) == 0 {
		return nil, nil
	}

	db := db.GetDB()

	var err error

	var res []*model.Faq
	var faq []models.FAQ

	query, args, err := sqlx.In("SELECT * FROM FAQ WHERE FAQID IN (?);", IDs)
	query = db.Rebind(query)

	if err = db.Select(&faq, query, args...); err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("NO_FAQ_RECORD")
		}
		return nil, err
	}

	for _, row := range faq {

		FAQID := cast.ToInt(row.FAQID)
		Question := row.Question
		Answer := row.Answer
		Position := cast.ToInt(row.Position)

		resRow := &model.Faq{
			Faqid:    FAQID,
			Question: &Question,
			Answer:   &Answer,
			Position: &Position,
		}

		res = append(res, resRow)
	}

	return res, nil
}

func (portal *Portal) QueryUploadArray(IDs []pgtype.Int8) ([]*model.Upload, error) {
	if len(IDs) == 0 {
		return nil, nil
	}

	db := db.GetDB()

	var err error

	var res []*model.Upload
	var upload []models.Upload

	query, args, err := sqlx.In("SELECT * FROM Upload WHERE UploadID IN (?);", IDs)

	// sqlx.In returns queries with the `?` bindvar, we can rebind it for our backend
	query = db.Rebind(query)

	if err = db.Select(&upload, query, args...); err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("NO_UPLOAD_RECORD")
		}
		return nil, err
	}

	for _, uploadRow := range upload {

		UploadID := cast.ToInt(uploadRow.UploadID)
		URL := viper.GetString("s3_cdn_url") + uploadRow.Filename
		Filename := uploadRow.Filename
		Category := uploadRow.Category
		Created := cast.ToInt(uploadRow.Created)

		resRow := &model.Upload{
			UploadID: UploadID,
			URL:      &URL,
			Filename: &Filename,
			Category: &Category,
			Created:  &Created,
		}

		res = append(res, resRow)
	}

	return res, nil
}

func (portal *Portal) QueryCategoryList() ([]*model.Category, error) {
	db := db.GetDB()

	var res []*model.Category
	var category []models.Category

	if err := db.Select(&category, "SELECT * FROM Category"); err != nil {
		if err != sql.ErrNoRows {
			return nil, err
		}
	}

	for _, categoryRow := range category {

		resRow := &model.Category{
			CategoryID: cast.ToInt(categoryRow.CategoryID),
			Title:      categoryRow.Title,
		}

		res = append(res, resRow)
	}

	return res, nil
}

func (portal *Portal) QueryDealList() ([]*model.Deal, error) {
	db := db.GetDB()

	var res []*model.Deal
	var Deal []models.Deal

	if err := db.Select(&Deal, "SELECT * FROM Deal WHERE MemberID=$1 AND EXISTS (SELECT * FROM Offer WHERE Offer.OfferID=Deal.OfferID AND (Offer.Status!=$2 AND Offer.Status!=$3)) ORDER BY DealID DESC", portal.Member.MemberID, constants.STATUS_NOACTIVE, constants.STATUS_CANCELLED); err != nil {
		if err != sql.ErrNoRows {
			return nil, err
		}
	}
	// AND (Status=$2 OR Status=$3 OR Status=$4)
	//, constants.STATUS_SIGNED, constants.STATUS_PAID, constants.STATUS_ACTIVE

	for _, row := range Deal {

		DealID := cast.ToInt(row.DealID)
		OfferID := cast.ToInt(row.OfferID)
		ContractID := cast.ToInt(row.ContractID)
		MemberID := cast.ToInt(row.MemberID)
		CurrencyID := cast.ToInt(row.CurrencyID)
		SignatureFilename := row.SignatureFilename
		SignatureURL := viper.GetString("s3_cdn_url") + row.SignatureFilename
		VerificationCode := row.VerificationCode
		DateCreated := row.DateCreated.Time.String()
		DateSigned := row.DateSigned.Time.String()
		DateVerified := row.DateVerified.Time.String()
		DatePaid := row.DatePaid.Time.String()
		DateStart := row.DateStart.Time.String()
		DateEnd := row.DateEnd.Time.String()
		Status := row.Status
		Amount := row.Amount.Decimal.String()
		Duration := row.Duration.Decimal.String()

		resRow := &model.Deal{
			DealID:            DealID,
			OfferID:           &OfferID,
			ContractID:        &ContractID,
			MemberID:          &MemberID,
			CurrencyID:        &CurrencyID,
			SignatureFilename: &SignatureFilename,
			SignatureURL:      &SignatureURL,
			VerificationCode:  &VerificationCode,
			DateCreated:       &DateCreated,
			DateSigned:        &DateSigned,
			DateVerified:      &DateVerified,
			DatePaid:          &DatePaid,
			DateStart:         &DateStart,
			DateEnd:           &DateEnd,
			Status:            &Status,
			Amount:            &Amount,
			Duration:          &Duration,
		}

		res = append(res, resRow)
	}

	return res, nil
}

func (portal *Portal) QueryTXList() ([]*model.Tx, error) {
	db := db.GetDB()

	var res []*model.Tx
	var tx []models.TX

	if err := db.Select(&tx, "SELECT * FROM TX WHERE MemberID=$1 ORDER BY TXID DESC", portal.Member.MemberID); err != nil {
		if err != sql.ErrNoRows {
			return nil, err
		}
	}

	for _, txRow := range tx {

		TXID := cast.ToInt(txRow.TXID)
		MemberID := cast.ToInt(txRow.MemberID)
		Amount := txRow.Amount.Decimal.String()
		AmountNegative := txRow.AmountNegative
		CurrencyID := cast.ToInt(txRow.CurrencyID.Int)
		Status := txRow.Status
		DateCreated := txRow.DateCreated.Time.String()
		DateComplete := txRow.DateComplete.Time.String()
		TimestampCreated := cast.ToInt(txRow.TimestampCreated)
		TimestampComplete := cast.ToInt(txRow.TimestampComplete)

		resRow := &model.Tx{
			Txid:              TXID,
			MemberID:          &MemberID,
			Amount:            &Amount,
			AmountNegative:    &AmountNegative,
			CurrencyID:        &CurrencyID,
			Status:            &Status,
			DateCreated:       &DateCreated,
			DateComplete:      &DateComplete,
			TimestampCreated:  &TimestampCreated,
			TimestampComplete: &TimestampComplete,
		}

		res = append(res, resRow)
	}

	return res, nil
}

func (portal *Portal) QueryBalanceList() ([]*model.Balance, error) {
	db := db.GetDB()

	var res []*model.Balance
	var balance []models.Balance

	if err := db.Select(&balance, "SELECT * FROM Balance WHERE MemberID=$1 ORDER BY BalanceID DESC", portal.Member.MemberID); err != nil {
		if err != sql.ErrNoRows {
			return nil, err
		}
	}

	for _, row := range balance {

		BalanceID := cast.ToInt(row.BalanceID)
		MemberID := cast.ToInt(row.MemberID)
		CurrencyID := cast.ToInt(row.CurrencyID.Int)
		Amount := row.Amount.Decimal.String()
		AmountNegative := row.AmountNegative

		resRow := &model.Balance{
			BalanceID:      BalanceID,
			MemberID:       &MemberID,
			CurrencyID:     &CurrencyID,
			Amount:         &Amount,
			AmountNegative: &AmountNegative,
		}

		res = append(res, resRow)
	}

	return res, nil
}

func (portal *Portal) QueryCurrencyList() ([]*model.Currency, error) {
	db := db.GetDB()

	var res []*model.Currency
	var currency []models.Currency

	if err := db.Select(&currency, "SELECT * FROM Currency"); err != nil {
		if err != sql.ErrNoRows {
			return nil, err
		}
	}

	for _, currencyRow := range currency {

		resRow := &model.Currency{
			CurrencyID: cast.ToInt(currencyRow.CurrencyID.Int),
			Title:      currencyRow.Title,
			Symbol:     currencyRow.Symbol,
		}

		res = append(res, resRow)
	}

	return res, nil
}

func (portal *Portal) QueryMediaArray(InvestID int64, Category string) ([]*model.Media, error) {
	db := db.GetDB()

	var err error

	var res []*model.Media
	var Media []models.Media

	if err = db.Select(&Media, "SELECT * FROM Media WHERE InvestID=$1 AND Category=$2 ORDER BY Position ASC", InvestID, Category); err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("NO_MEDIA_RECORD")
		}
		return nil, err
	}

	for _, row := range Media {

		MediaID := cast.ToInt(row.MediaID.Int)
		MemberID := cast.ToInt(row.MemberID.Int)
		Title := row.Title.String
		Position := cast.ToInt(row.Position.Int)
		URL := viper.GetString("s3_cdn_url") + row.Filename.String
		Filename := row.Filename.String
		Category := row.Category.String
		Created := cast.ToInt(row.Created.Int)

		resRow := &model.Media{
			MediaID:  MediaID,
			MemberID: &MemberID,
			Title:    &Title,
			Position: &Position,
			URL:      &URL,
			Filename: &Filename,
			Category: &Category,
			Created:  &Created,
		}

		res = append(res, resRow)
	}

	return res, nil
}

func (portal *Portal) QueryOfferList() ([]*model.Invest, error) {
	db := db.GetDB()

	var err error

	var res []*model.Invest
	var Invest []models.InvestOffer

	if err = db.Select(&Invest, "SELECT Offer.OfferID, Offer.Status, Invest.InvestID, Invest.CategoryID, Offer.CurrencyID, Offer.BankDetailsID, Invest.Title, Invest.Subtitle, Invest.Description FROM Offer LEFT OUTER JOIN Invest ON Invest.InvestID=Offer.InvestID WHERE Offer.MemberID=$1 AND (Offer.Status=$2 OR Offer.Status=$3 OR Offer.Status=$4)", portal.Member.MemberID, constants.STATUS_ACTIVE, constants.STATUS_SIGNED, constants.STATUS_PAID); err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("NO_OFFER_RECORD")
		}
		return nil, err
	}

	for _, row := range Invest {

		InvestID := cast.ToInt(row.InvestID)
		OfferID := cast.ToInt(row.OfferID)
		CategoryID := cast.ToInt(row.CategoryID)
		CurrencyID := cast.ToInt(row.CurrencyID.Int)
		BankDetailsID := cast.ToInt(row.BankDetailsID)
		Status := row.Status
		Title := row.Title
		Subtitle := row.Subtitle
		Description := row.Description

		Photo, _ := portal.QueryMediaArray(row.InvestID, constants.CATEGORY_PHOTO)
		Document, _ := portal.QueryMediaArray(row.InvestID, constants.CATEGORY_DOCUMENT)
		// FAQ, _ := portal.QueryFAQArray(row.FAQ.Elements)

		resRow := &model.Invest{
			InvestID:      InvestID,
			OfferID:       &OfferID,
			Status:        &Status,
			Title:         &Title,
			CategoryID:    &CategoryID,
			CurrencyID:    &CurrencyID,
			BankDetailsID: &BankDetailsID,
			Subtitle:      &Subtitle,
			Description:   &Description,
			Photo:         Photo,
			Document:      Document,
			// Faq: FAQ,
		}

		res = append(res, resRow)
	}

	return res, nil
}

func (portal *Portal) QueryInvestByOfferID(OfferID int) (*model.Invest, error) {
	db := db.GetDB()

	var err error

	var offer models.Offer

	if err = db.Get(&offer, "SELECT * FROM Offer WHERE OfferID=$1", OfferID); err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("NO_OFFER_RECORD")
		}
		return nil, err
	}

	var invest models.Invest

	if err = db.Get(&invest, "SELECT * FROM Invest WHERE InvestID=$1", offer.InvestID); err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("NO_INVEST_RECORD")
		}
		return nil, err
	}

	InvestID := cast.ToInt(invest.InvestID)
	CategoryID := cast.ToInt(invest.CategoryID)
	CurrencyID := cast.ToInt(offer.CurrencyID.Int)
	BankDetailsID := cast.ToInt(offer.BankDetailsID)
	Status := offer.Status
	Title := invest.Title
	Subtitle := invest.Subtitle
	Description := invest.Description

	Photo, _ := portal.QueryMediaArray(invest.InvestID, constants.CATEGORY_PHOTO)
	Document, _ := portal.QueryMediaArray(invest.InvestID, constants.CATEGORY_DOCUMENT)
	// FAQ, _ := portal.QueryFAQArray(row.FAQ.Elements)

	res := &model.Invest{
		InvestID:      InvestID,
		OfferID:       &OfferID,
		CurrencyID:    &CurrencyID,
		BankDetailsID: &BankDetailsID,
		Status:        &Status,
		Title:         &Title,
		CategoryID:    &CategoryID,
		Subtitle:      &Subtitle,
		Description:   &Description,
		Photo:         Photo,
		Document:      Document,
		// Faq: FAQ,
	}

	return res, nil
}

func (portal *Portal) QueryBankDetails(RecordID int) (*model.BankDetails, error) {
	db := db.GetDB()

	var err error

	var bankDetails models.BankDetails

	if err = db.Get(&bankDetails, "SELECT * FROM BankDetails WHERE BankDetailsID=$1", RecordID); err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("NO_BANK_DETAILS_RECORD")
		}
		return nil, err
	}

	BankDetailsID := cast.ToInt(bankDetails.BankDetailsID)
	Title := bankDetails.Title
	BeneficiaryCompany := bankDetails.BeneficiaryCompany
	BeneficiaryFirstName := bankDetails.BeneficiaryFirstName
	BeneficiaryLastName := bankDetails.BeneficiaryLastName
	BeneficiaryCountry := bankDetails.BeneficiaryCountry
	BeneficiaryCity := bankDetails.BeneficiaryCity
	BeneficiaryZip := bankDetails.BeneficiaryZip
	BeneficiaryAddress := bankDetails.BeneficiaryAddress
	BankName := bankDetails.BankName
	BankBranch := bankDetails.BankBranch
	BankIFSC := bankDetails.BankIFSC
	BankBranchCountry := bankDetails.BankBranchCountry
	BankBranchCity := bankDetails.BankBranchCity
	BankBranchZip := bankDetails.BankBranchZip
	BankBranchAddress := bankDetails.BankBranchAddress
	BankAccountNumber := bankDetails.BankAccountNumber
	BankAccountType := bankDetails.BankAccountType
	BankRoutingNumber := bankDetails.BankRoutingNumber
	BankTransferCaption := bankDetails.BankTransferCaption
	BankIBAN := bankDetails.BankIBAN
	BankSWIFT := bankDetails.BankSWIFT
	BankSWIFTCorrespondent := bankDetails.BankSWIFTCorrespondent
	BankBIC := bankDetails.BankBIC

	res := &model.BankDetails{
		BankDetailsID:          &BankDetailsID,
		Title:                  &Title,
		BeneficiaryCompany:     &BeneficiaryCompany,
		BeneficiaryFirstName:   &BeneficiaryFirstName,
		BeneficiaryLastName:    &BeneficiaryLastName,
		BeneficiaryCountry:     &BeneficiaryCountry,
		BeneficiaryCity:        &BeneficiaryCity,
		BeneficiaryZip:         &BeneficiaryZip,
		BeneficiaryAddress:     &BeneficiaryAddress,
		BankName:               &BankName,
		BankBranch:             &BankBranch,
		BankIfsc:               &BankIFSC,
		BankBranchCountry:      &BankBranchCountry,
		BankBranchCity:         &BankBranchCity,
		BankBranchZip:          &BankBranchZip,
		BankBranchAddress:      &BankBranchAddress,
		BankAccountNumber:      &BankAccountNumber,
		BankAccountType:        &BankAccountType,
		BankRoutingNumber:      &BankRoutingNumber,
		BankTransferCaption:    &BankTransferCaption,
		BankIban:               &BankIBAN,
		BankSwift:              &BankSWIFT,
		BankSWIFTCorrespondent: &BankSWIFTCorrespondent,
		BankBic:                &BankBIC,
	}

	return res, nil
}

func (portal *Portal) QueryInterestListByOfferID(RecordID int) ([]*model.Interest, error) {
	db := db.GetDB()

	var err error

	var interest []models.Interest

	if err = db.Select(&interest, "SELECT * FROM Interest WHERE OfferID=$1 ORDER BY DurationFrom ASC", RecordID); err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("NO_INTEREST_RECORD")
		}
		return nil, err
	}

	var res []*model.Interest

	for _, interestRow := range interest {

		InterestID := cast.ToInt(interestRow.InterestID)
		OfferID := cast.ToInt(interestRow.OfferID)
		AmountFrom := interestRow.AmountFrom.Decimal.String()
		AmountTo := interestRow.AmountTo.Decimal.String()
		DurationFrom := interestRow.DurationFrom.Decimal.String()
		DurationTo := interestRow.DurationTo.Decimal.String()
		Interest := interestRow.Interest.Decimal.String()

		resRow := &model.Interest{
			InterestID:   InterestID,
			OfferID:      &OfferID,
			AmountFrom:   &AmountFrom,
			AmountTo:     &AmountTo,
			DurationFrom: &DurationFrom,
			DurationTo:   &DurationTo,
			Interest:     &Interest,
		}

		res = append(res, resRow)
	}

	return res, nil
}

func (portal *Portal) QueryDealByOfferID(RecordID int) (*model.Deal, error) {
	db := db.GetDB()

	var err error

	var row models.Deal

	if err = db.Get(&row, "SELECT * FROM Deal WHERE OfferID=$1 AND MemberID=$2 AND (Status=$3 OR Status=$4) ORDER BY DealID DESC", RecordID, portal.Member.MemberID, constants.STATUS_PENDING, constants.STATUS_SIGNED); err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("NO_DEAL_RECORD")
		}
		return nil, err
	}

	portal.Deal = row

	DealID := cast.ToInt(row.DealID)
	OfferID := cast.ToInt(row.OfferID)
	ContractID := cast.ToInt(row.ContractID)
	MemberID := cast.ToInt(row.MemberID)
	CurrencyID := cast.ToInt(row.CurrencyID)
	SignatureFilename := row.SignatureFilename
	SignatureURL := viper.GetString("s3_cdn_url") + row.SignatureFilename
	VerificationCode := row.VerificationCode
	DateCreated := row.DateCreated.Time.String()
	DateSigned := row.DateSigned.Time.String()
	DateVerified := row.DateVerified.Time.String()
	DatePaid := row.DatePaid.Time.String()
	DateStart := row.DateStart.Time.String()
	DateEnd := row.DateEnd.Time.String()
	Status := row.Status
	Amount := row.Amount.Decimal.String()
	Duration := row.Duration.Decimal.String()

	res := &model.Deal{
		DealID:            DealID,
		OfferID:           &OfferID,
		ContractID:        &ContractID,
		MemberID:          &MemberID,
		CurrencyID:        &CurrencyID,
		SignatureFilename: &SignatureFilename,
		SignatureURL:      &SignatureURL,
		VerificationCode:  &VerificationCode,
		DateCreated:       &DateCreated,
		DateSigned:        &DateSigned,
		DateVerified:      &DateVerified,
		DatePaid:          &DatePaid,
		DateStart:         &DateStart,
		DateEnd:           &DateEnd,
		Status:            &Status,
		Amount:            &Amount,
		Duration:          &Duration,
	}

	return res, nil
}

func (portal *Portal) QueryDeal(RecordID int) (*model.Deal, error) {
	db := db.GetDB()

	var err error

	var row models.Deal

	if err = db.Get(&row, "SELECT * FROM Deal WHERE DealID=$1 AND MemberID=$2", RecordID, portal.Member.MemberID); err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("NO_DEAL_RECORD")
		}
		return nil, err
	}

	portal.Deal = row

	DealID := cast.ToInt(row.DealID)
	OfferID := cast.ToInt(row.OfferID)
	ContractID := cast.ToInt(row.ContractID)
	MemberID := cast.ToInt(row.MemberID)
	CurrencyID := cast.ToInt(row.CurrencyID)
	SignatureFilename := row.SignatureFilename
	SignatureURL := viper.GetString("s3_cdn_url") + row.SignatureFilename
	VerificationCode := row.VerificationCode
	DateCreated := row.DateCreated.Time.String()
	DateSigned := row.DateSigned.Time.String()
	DateVerified := row.DateVerified.Time.String()
	DatePaid := row.DatePaid.Time.String()
	DateStart := row.DateStart.Time.String()
	DateEnd := row.DateEnd.Time.String()
	Status := row.Status
	Amount := row.Amount.Decimal.String()
	Duration := row.Duration.Decimal.String()

	res := &model.Deal{
		DealID:            DealID,
		OfferID:           &OfferID,
		ContractID:        &ContractID,
		MemberID:          &MemberID,
		CurrencyID:        &CurrencyID,
		SignatureFilename: &SignatureFilename,
		SignatureURL:      &SignatureURL,
		VerificationCode:  &VerificationCode,
		DateCreated:       &DateCreated,
		DateSigned:        &DateSigned,
		DateVerified:      &DateVerified,
		DatePaid:          &DatePaid,
		DateStart:         &DateStart,
		DateEnd:           &DateEnd,
		Status:            &Status,
		Amount:            &Amount,
		Duration:          &Duration,
	}

	return res, nil
}

func (portal *Portal) QueryInvoiceByDealID(RecordID int) (*model.Invoice, error) {
	db := db.GetDB()

	var err error

	var row models.Invoice

	if err = db.Get(&row, "SELECT * FROM Invoice WHERE DealID=$1 AND MemberID=$2", RecordID, portal.Member.MemberID); err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("NO_INVOICE_RECORD")
		}
		return nil, err
	}

	portal.Invoice = row

	InvoiceID := cast.ToInt(row.InvoiceID)
	OfferID := cast.ToInt(row.OfferID)
	MemberID := cast.ToInt(row.MemberID)
	DealID := cast.ToInt(row.DealID)
	CurrencyID := cast.ToInt(row.CurrencyID)
	Amount := row.Amount.Decimal.String()
	Status := row.Status
	DateCreated := row.DateCreated.Time.String()
	TimestampCreated := cast.ToInt(row.TimestampCreated)
	DatePaid := row.DatePaid.Time.String()
	TimestampPaid := cast.ToInt(row.TimestampPaid)

	res := &model.Invoice{
		InvoiceID:        InvoiceID,
		OfferID:          &OfferID,
		MemberID:         &MemberID,
		DealID:           &DealID,
		CurrencyID:       &CurrencyID,
		Amount:           &Amount,
		Status:           &Status,
		DateCreated:      &DateCreated,
		TimestampCreated: &TimestampCreated,
		DatePaid:         &DatePaid,
		TimestampPaid:    &TimestampPaid,
	}

	return res, nil
}

func (portal *Portal) QueryInvoice(RecordID int) (*model.Invoice, error) {
	db := db.GetDB()

	var err error

	var row models.Invoice

	if err = db.Get(&row, "SELECT * FROM Invoice WHERE InvoiceID=$1 AND MemberID=$2", RecordID, portal.Member.MemberID); err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("NO_INVOICE_RECORD")
		}
		return nil, err
	}

	portal.Invoice = row

	InvoiceID := cast.ToInt(row.InvoiceID)
	OfferID := cast.ToInt(row.OfferID)
	MemberID := cast.ToInt(row.MemberID)
	DealID := cast.ToInt(row.DealID)
	CurrencyID := cast.ToInt(row.CurrencyID)
	Amount := row.Amount.Decimal.String()
	Status := row.Status
	DateCreated := row.DateCreated.Time.String()
	TimestampCreated := cast.ToInt(row.TimestampCreated)
	DatePaid := row.DatePaid.Time.String()
	TimestampPaid := cast.ToInt(row.TimestampPaid)

	res := &model.Invoice{
		InvoiceID:        InvoiceID,
		OfferID:          &OfferID,
		MemberID:         &MemberID,
		DealID:           &DealID,
		CurrencyID:       &CurrencyID,
		Amount:           &Amount,
		Status:           &Status,
		DateCreated:      &DateCreated,
		TimestampCreated: &TimestampCreated,
		DatePaid:         &DatePaid,
		TimestampPaid:    &TimestampPaid,
	}

	return res, nil
}

func (portal *Portal) QueryInterestByDealID() (*model.Interest, error) {
	db := db.GetDB()

	var err error

	var interest models.Interest

	if err = db.Get(&interest, "SELECT * FROM Interest WHERE OfferID=$1 AND (DurationFrom<=$2 AND DurationTo>=$2) AND (AmountFrom<=$3 AND AmountTo>=$3)", portal.Deal.OfferID, portal.Deal.Duration.Decimal, portal.Deal.Amount.Decimal); err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("NO_INTEREST_RECORD")
		}
		return nil, err
	}

	portal.Interest = interest

	InterestID := cast.ToInt(interest.InterestID)
	OfferID := cast.ToInt(interest.OfferID)
	AmountFrom := interest.AmountFrom.Decimal.String()
	AmountTo := interest.AmountTo.Decimal.String()
	DurationFrom := interest.DurationFrom.Decimal.String()
	DurationTo := interest.DurationTo.Decimal.String()
	Interest := interest.Interest.Decimal.String()

	res := &model.Interest{
		InterestID:   InterestID,
		OfferID:      &OfferID,
		AmountFrom:   &AmountFrom,
		AmountTo:     &AmountTo,
		DurationFrom: &DurationFrom,
		DurationTo:   &DurationTo,
		Interest:     &Interest,
	}

	return res, nil
}

func (portal *Portal) QueryInterestByOfferID() (*model.Interest, error) {
	db := db.GetDB()

	var err error

	var interest models.Interest

	if err = db.Get(&interest, "SELECT * FROM Interest WHERE OfferID=$1 ORDER BY Interest DESC", portal.Offer.OfferID); err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("NO_INTEREST_RECORD")
		}
		return nil, err
	}

	portal.Interest = interest

	InterestID := cast.ToInt(interest.InterestID)
	OfferID := cast.ToInt(interest.OfferID)
	AmountFrom := interest.AmountFrom.Decimal.String()
	AmountTo := interest.AmountTo.Decimal.String()
	DurationFrom := interest.DurationFrom.Decimal.String()
	DurationTo := interest.DurationTo.Decimal.String()
	Interest := interest.Interest.Decimal.String()

	res := &model.Interest{
		InterestID:   InterestID,
		OfferID:      &OfferID,
		AmountFrom:   &AmountFrom,
		AmountTo:     &AmountTo,
		DurationFrom: &DurationFrom,
		DurationTo:   &DurationTo,
		Interest:     &Interest,
	}

	return res, nil
}

func (portal *Portal) UpdateDealSignature(filename string) error {
	db := db.GetDB()

	tx := db.MustBegin()
	tx.MustExec("UPDATE Deal SET SignatureFilename=$1, DateSigned=CURRENT_TIMESTAMP WHERE DealID=$2", filename, portal.Deal.DealID)
	tx.Commit()

	return nil
}

//Query offer
func (portal *Portal) QueryOffer(OfferID int) error {
	db := db.GetDB()

	var err error

	var offer models.Offer

	if err = db.Get(&offer, "SELECT * FROM Offer WHERE (Offer.Status=$1 OR Offer.Status=$2 OR Offer.Status=$3) AND OfferID=$4", constants.STATUS_ACTIVE, constants.STATUS_SIGNED, constants.STATUS_PAID, OfferID); err != nil {
		if err == sql.ErrNoRows {
			return errors.New("NO_OFFER_RECORD")
		}
		return err
	}

	portal.Offer = offer

	return nil
}

func (portal *Portal) ProcessSignatureImage(Image string) ([]byte, string, error) {

	encodedImage := strings.Split(Image, ",")

	if encodedImage[0] != "data:image/svg+xml;base64" {
		return nil, "", errors.New("INVALID_IMAGE")
	}

	decodedImage, err := base64.StdEncoding.DecodeString(encodedImage[1])
	if err != nil {
		return nil, "", err
	}

	if !utils.IsSVG(decodedImage) {
		return nil, "", errors.New("INVALID_IMAGE")
	}

	filename := cast.ToString(portal.Member.MemberID) + "_" + uuid.New().String() + ".svg"

	return decodedImage, filename, nil
}

//Query invest
func (portal *Portal) QueryInvest() error {
	db := db.GetDB()

	var err error

	var invest models.Invest

	if err = db.Get(&invest, "SELECT * FROM Invest WHERE InvestID=$1", portal.Offer.InvestID); err != nil {
		if err == sql.ErrNoRows {
			return errors.New("NO_INVEST_RECORD")
		}
		return err
	}

	portal.Invest = invest

	return nil
}

//Query offer contract
func (portal *Portal) QueryCurrentContract() error {
	db := db.GetDB()

	var err error

	var contract models.Contract

	if err = db.Get(&contract, "SELECT * FROM Contract WHERE OfferID=$1 AND Current=$2", portal.Offer.OfferID, true); err != nil {
		if err == sql.ErrNoRows {
			return errors.New("NO_CONTRACT_RECORD")
		}
		return err
	}

	portal.Contract = contract

	return nil
}

//Mustache parse contract content
func (portal *Portal) ParseContractContent() error {
	var err error

	portal.Contract.Content, err = mustache.Render(portal.Contract.ContentRaw, map[string]string{"FirstName": portal.Member.FirstName})
	if err != nil {
		return errors.New("CONTRACT_CONTENT_PARSE_ERROR")
	}

	return nil
}

func (portal *Portal) CreateDeal(Amount string, Duration string) (*model.Result, error) {
	db := db.GetDB()

	tx := db.MustBegin()
	tx.MustExec("UPDATE Deal SET Status=$1 WHERE OfferID=$2 AND MemberID=$3 AND Status=$4", constants.STATUS_CANCELLED, portal.Offer.OfferID, portal.Member.MemberID, constants.STATUS_PENDING)
	tx.MustExec("INSERT INTO Deal (OfferID, ContractID, MemberID, CurrencyID, Status, Amount, Duration) VALUES ($1, $2, $3, $4, $5, $6, $7)", portal.Offer.OfferID, portal.Contract.ContractID, portal.Member.MemberID, portal.Offer.CurrencyID, constants.STATUS_PENDING, Amount, Duration)
	tx.Commit()

	var RecordID int

	if err := db.Get(&RecordID, "SELECT DealID FROM Deal WHERE OfferID=$1 AND ContractID=$2 AND MemberID=$3 AND Status=$4 ORDER BY DealID DESC", portal.Offer.OfferID, portal.Contract.ContractID, portal.Member.MemberID, constants.STATUS_PENDING); err != nil {
		return nil, err
	}

	return &model.Result{
		Status: true,
	}, nil
}

//Generate confirmation code
func (portal *Portal) GenerateOTP() error {
	code, err := password.Generate(6, 6, 0, true, false)
	if err != nil {
		return err
	}

	portal.Code = code

	return nil
}

//Hash using bcrypt
func (portal *Portal) HashBcrypt(str string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(str), 12)
	if err != nil {
		return "", err
	}

	return cast.ToString(hash), nil
}

//Hash using md5
func (portal *Portal) HashMD5(str string) (string, error) {
	hash := md5.Sum([]byte(str))
	return hex.EncodeToString(hash[:]), nil
}

//Parse phone number
func (portal *Portal) ParsePhone(phone string) error {

	// phone = strings.Replace(phone, "+", "", -1)

	phoneNumber, err := phonenumbers.Parse(phone, "")
	if err != nil {
		return err
	}

	if phonenumbers.IsValidNumber(phoneNumber) == false {
		return errors.New("INVALID_PHONE")
	}

	portal.Phone = cast.ToString(phoneNumber.GetCountryCode()) + cast.ToString(phoneNumber.GetNationalNumber())

	return nil
}

//Parse email
func (portal *Portal) ParseEmail(email string) error {

	if err := validation.Validate(email,
		validation.Length(0, 250),
		validation.Required,
		is.Email,
	); err != nil {
		return errors.New("INVALID_EMAIL")
	}

	portal.Email = strings.ToLower(email)

	return nil
}

//Query member by email
func (portal *Portal) QueryMemberByEmail() error {
	db := db.GetDB()

	var member models.Member

	if err := db.Get(&member, "SELECT * FROM Member WHERE Email=$1", portal.Email); err != nil {
		if err == sql.ErrNoRows {
			return errors.New("NO_MEMBER_RECORD")
		}
		return err
	}

	portal.Member = member

	return nil
}

//Query member by phone
func (portal *Portal) QueryMemberByPhone() error {
	db := db.GetDB()

	var member models.Member

	if err := db.Get(&member, "SELECT * FROM Member WHERE Phone=$1", portal.Email); err != nil {
		if err == sql.ErrNoRows {
			return errors.New("NO_MEMBER_RECORD")
		}
		return err
	}

	portal.Member = member

	return nil
}

//Query member count by phone
func (portal *Portal) CheckPhoneInUse() error {
	db := db.GetDB()

	var count int

	if err := db.Get(&count, "SELECT count(*) FROM Member WHERE Phone=$1", portal.Phone); err != nil {
		if err != sql.ErrNoRows {
			return err
		}
	}

	if count > 0 {
		return errors.New("PHONE_ALREADY_IN_USE")
	}

	return nil
}

//Query member count by email
func (portal *Portal) CheckEmailInUse() error {
	db := db.GetDB()

	var count int

	if err := db.Get(&count, "SELECT count(*) FROM Member WHERE Email=$1", portal.Email); err != nil {
		if err != sql.ErrNoRows {
			return err
		}
	}

	if count > 0 {
		return errors.New("EMAIL_ALREADY_IN_USE")
	}

	return nil
}

//Find verification code sent in the last timeout value, in case found return time left until new code request
func (portal *Portal) QueryVerifyTimeout() (int64, error) {
	db := db.GetDB()

	var verify models.Verify

	if err := db.Get(&verify, "SELECT * FROM Verify WHERE MemberID=$1 AND Action=$2 AND Method=$3 AND Status=$4", portal.Member.MemberID, portal.Action, portal.Method, constants.STATUS_PENDING); err != nil {
		if err != sql.ErrNoRows {
			return 0, err
		}
	}

	//Not enough time passed after the last code request
	if verify.Created > portal.Timestamp-portal.Timeout {
		return portal.Timeout - portal.Timestamp + verify.Created, errors.New("CODE_TIMEOUT")
	}

	return portal.Timeout, nil

}

func (portal *Portal) CreateVerify() (int, error) {
	db := db.GetDB()

	var err error
	var Timeout int64

	err = portal.GinContextFromContext()
	if err != nil {
		return 0, err
	}

	if portal.Action == constants.ACTION_SIGN_CONTRACT || portal.Action == constants.ACTION_SIGNUP {
		Timeout, err = portal.QueryVerifyTimeout()
		if err != nil {
			return cast.ToInt(Timeout), err
		}
	}

	err = portal.GenerateOTP()
	if err != nil {
		return 0, err
	}

	portal.CodeHash, err = portal.HashBcrypt(portal.Code)
	if err != nil {
		return 0, err
	}

	if portal.Action == constants.ACTION_SIGN_CONTRACT {
		tx := db.MustBegin()
		tx.MustExec("UPDATE Verify SET Status=$1 WHERE MemberID=$2 AND Action=$3 AND Method=$4 AND Status=$5", constants.STATUS_CANCELLED, portal.Member.MemberID, portal.Action, portal.Method, constants.STATUS_PENDING)
		tx.MustExec("INSERT INTO Verify (MemberID, Action, Method, CodeHash, Phone, Status, IP, Created) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)", portal.Member.MemberID, portal.Action, portal.Method, portal.CodeHash, portal.Phone, constants.STATUS_PENDING, portal.c.ClientIP(), portal.Timestamp)
		tx.Commit()

		// err = otp.SendSMS(portal.Phone, fmt.Sprintf("Your code is %s", portal.Code))
		// if err != nil {
		// 	return nil, err
		// }
		fmt.Println(portal.Code)
	}

	portal.EmailHash, err = portal.HashMD5(portal.Email)
	if err != nil {
		return 0, err
	}

	if portal.Action == constants.ACTION_SIGNUP && portal.Method == constants.METHOD_EMAIL {

		tx := db.MustBegin()
		tx.MustExec("UPDATE Verify SET Status=$1 WHERE Action=$2 AND Method=$3 AND Email=$4 AND Status=$5", constants.STATUS_CANCELLED, portal.Action, portal.Method, portal.Email, constants.STATUS_PENDING)
		tx.MustExec("INSERT INTO Verify (MemberID, Action, Method, CodeHash, Email, EmailHash, Status, IP, Created) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)", portal.Member.MemberID, portal.Action, portal.Method, portal.CodeHash, portal.Email, portal.EmailHash, constants.STATUS_PENDING, portal.c.ClientIP(), portal.Timestamp)
		tx.Commit()
	}

	if portal.Action == constants.ACTION_RESET && portal.Method == constants.METHOD_EMAIL {

		tx := db.MustBegin()
		tx.MustExec("UPDATE Verify SET Status=$1 WHERE MemberID=$2 AND Action=$3 AND Method=$4 AND Status=$5", constants.STATUS_CANCELLED, portal.Member.MemberID, constants.ACTION_RESET, constants.METHOD_EMAIL, constants.STATUS_PENDING)
		tx.MustExec("INSERT INTO Verify (MemberID, Action, Method, CodeHash, Email, EmailHash, Status, IP, Created) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)", portal.Member.MemberID, portal.Action, portal.Method, portal.CodeHash, portal.Email, portal.EmailHash, constants.STATUS_PENDING, portal.c.ClientIP(), portal.Timestamp)
		tx.Commit()
	}

	return cast.ToInt(Timeout), nil

}

//Validate password
func (portal *Portal) ValidatePassword(password string, validateStrength bool) error {

	if err := validation.Validate(password,
		validation.Required,
	); err != nil {
		return errors.New("INVALID_PASSWORD")
	}

	//TODO: remove validateStrength
	if validateStrength {
		validator := crunchy.NewValidatorWithOpts(crunchy.Options{
			// MinLength is the minimum length required for a valid password
			// (must be >= 1, default is 8)
			MinLength: 6,

			// MinDiff is the minimum amount of unique characters required for a valid password
			// (must be >= 1, default is 5)
			MinDiff: 5,

			// Hashers will be used to find hashed passwords in dictionaries
			Hashers: []hash.Hash{md5.New(), sha1.New(), sha256.New(), sha512.New()},

			// Check haveibeenpwned.com database
			CheckHIBP: true,
		})

		if err := validator.Check(password); err != nil {
			return errors.New("WEAK_PASSWORD")
		}
	}

	if strings.ToLower(password) == portal.Email {
		return errors.New("EMAIL_PASSWORD_MATCH")
	}

	portal.Password = password

	return nil
}

//Query verify by md5 hash (email)
func (portal *Portal) QueryVerifyByHash(Hash string) error {
	db := db.GetDB()

	var verify models.Verify

	//Find verify request
	if err := db.Get(&verify, "SELECT * FROM Verify WHERE EmailHash=$1 AND Action=$2 AND Method=$3 AND Status=$4", Hash, portal.Action, portal.Method, constants.STATUS_PENDING); err != nil {
		if err == sql.ErrNoRows {
			return errors.New("INVALID_LINK")
		}
		return err
	}

	//Request is expired (day)
	if verify.Created < portal.Timestamp-86400 {
		return errors.New("EXPIRED_REQUEST")
	}

	portal.Verify = verify

	return nil
}

func (portal *Portal) ValidateVerifyAction(Action string) error {

	if Action != constants.ACTION_SIGNUP && Action != constants.ACTION_RESET {
		return errors.New("INVALID_ACTION")
	}

	portal.Action = Action

	return nil
}

func (portal *Portal) ValidateResetAction(Action string) error {

	if Action != constants.ACTION_RESET {
		return errors.New("INVALID_ACTION")
	}

	portal.Action = Action

	return nil
}

func (portal *Portal) ValidateVerifyMethod(Method string) error {

	if Method != constants.METHOD_EMAIL { //Method != constants.METHOD_PHONE &&
		return errors.New("INVALID_METHOD")
	}

	portal.Method = Method

	return nil
}

func (portal *Portal) ValidateResetMethod(Method string) error {

	if Method != constants.METHOD_EMAIL {
		return errors.New("INVALID_METHOD")
	}

	portal.Method = Method

	return nil
}

//Query verify by md5 hash (email)
func (portal *Portal) QueryVerifyByOfferID(OfferID int) error {
	db := db.GetDB()

	var deal models.Deal

	//Find deal record
	if err := db.Get(&deal, "SELECT * FROM Deal WHERE OfferID=$1 AND MemberID=$2 AND Status=$3", OfferID, portal.Member.MemberID, constants.STATUS_PENDING); err != nil {
		if err == sql.ErrNoRows {
			return errors.New("NO_DEAL_RECORD")
		}
		return err
	}

	portal.Deal = deal

	var verify models.Verify

	//Find verify request
	if err := db.Get(&verify, "SELECT * FROM Verify WHERE MemberID=$1 AND Action=$2 AND Method=$3 AND Status=$4", portal.Member.MemberID, portal.Action, portal.Method, constants.STATUS_PENDING); err != nil {
		if err == sql.ErrNoRows {
			return errors.New("NO_VERIFY_RECORD")
		}
		return err
	}

	//Request is expired (day)
	if verify.Created < portal.Timestamp-86400 {
		return errors.New("EXPIRED_REQUEST")
	}

	portal.Verify = verify
	return nil
}

//Validate OTP code
func (portal *Portal) ValidateOTPCode(code string) error {
	db := db.GetDB()

	if err := bcrypt.CompareHashAndPassword([]byte(portal.Verify.CodeHash), []byte(code)); err != nil {

		tx := db.MustBegin()
		tx.MustExec("UPDATE Verify SET Attempts=Attempts+$1 WHERE VerifyID=$2", 1, portal.Verify.VerifyID)
		tx.Commit()

		//Verify attempts exceeded
		if portal.Verify.Attempts >= 3 {

			tx := db.MustBegin()
			tx.MustExec("UPDATE Verify SET Status=$1 WHERE VerifyID=$2", constants.STATUS_FAIL, portal.Verify.VerifyID)
			tx.Commit()

			return errors.New("TOO_MANY_ATTEMPTS")
		}

		return errors.New("INVALID_CODE")
	}

	return nil
}

func (portal *Portal) CheckMemberStatusIsNoActive() error {

	if portal.Member.Status != constants.STATUS_NOACTIVE {
		return errors.New("INVALID_ACCOUNT_STATUS")
	}

	return nil
}

func (portal *Portal) CheckMemberStatusIsActive() error {

	if portal.Member.Status == constants.STATUS_NOACTIVE {
		return errors.New("RESEND_EMAIL_CONFIRMATION")
	}

	return nil
}

func (portal *Portal) CheckMemberStatusIsDisabled() error {

	if portal.Member.Status == constants.STATUS_DISABLED {
		return errors.New("ACCOUNT_DISABLED")
	}

	return nil
}

//Validate password
func (portal *Portal) CheckPassword() error {

	if err := bcrypt.CompareHashAndPassword([]byte(portal.Member.PasswordHash), []byte(portal.Password)); err != nil {

		//TODO: bruteforce protection

		return errors.New("INVALID_PASSWORD")
	}

	return nil
}

//Query setttings
func (portal *Portal) QuerySettings() error {
	db := db.GetDB()

	var settings models.Settings

	err := db.Get(&settings, "SELECT * FROM Settings WHERE SettingsID=$1", 1)
	if err != nil {
		return err
	}

	portal.Settings = settings

	return nil
}

//Generate email confirmation link
func (portal *Portal) GenerateConfirmationLink() error {
	var err error

	err = portal.QuerySettings()
	if err != nil {
		return err
	}

	portal.ConfirmationLink = portal.Settings.PlatformURL + "verify/" + portal.Action + "/" + portal.Method + "/" + portal.EmailHash + "/" + portal.Code

	return nil
}

//Generate invoice link
func (portal *Portal) GenerateInvoiceLink() error {
	var err error

	err = portal.QuerySettings()
	if err != nil {
		return err
	}

	portal.InvoiceLink = portal.Settings.PlatformURL + "invoice/" + cast.ToString(portal.Invoice.InvoiceID)

	return nil
}

func (portal *Portal) ValidateFirstName(firstName string) error {
	if err := validation.Validate(firstName,
		validation.Length(1, 30),
		validation.Required,
	); err != nil {
		return errors.New("INVALID_FIRST_NAME")
	}

	portal.FirstName = firstName

	return nil
}

func (portal *Portal) ValidateLastName(lastName string) error {
	if err := validation.Validate(lastName,
		validation.Length(1, 30),
		validation.Required,
	); err != nil {
		return errors.New("INVALID_LAST_NAME")
	}

	portal.LastName = lastName

	return nil
}

func (portal *Portal) ValidateBirthday(birthday string) error {

	//Year not less 1900 & not bigger than current. Date format YYYY-MM-DD
	if err := validation.Validate(birthday,
		validation.Date("2006-01-02").Min(time.Date(1900, 1, 1, 0, 0, 0, 0, time.UTC)).Max(time.Date(time.Now().Year()+1, 1, 1, 0, 0, 0, 0, time.UTC)),
		validation.Required,
	); err != nil {
		return errors.New("INVALID_BIRTHDAY")
	}

	portal.Birthday = birthday

	return nil
}

func (portal *Portal) IsValidCountryCode(countryCode string) bool {
	switch countryCode {
	case
		"GB",
		"US":
		return true
	}
	return false
}

func (portal *Portal) ValidateCitizenship(countryCode string) error {
	if err := validation.Validate(countryCode,
		validation.Length(2, 2),
		validation.Required,
	); err != nil {
		return errors.New("INVALID_CITIZENSHIP")
	}

	portal.Citizenship = countryCode

	return nil
}

func (portal *Portal) ValidateCountry(countryCode string) error {
	if err := validation.Validate(countryCode,
		validation.Length(2, 2),
		validation.Required,
	); err != nil {
		return errors.New("INVALID_COUNTRY")
	}

	portal.Country = countryCode

	return nil
}

func (portal *Portal) ValidateCity(city string) error {
	if err := validation.Validate(city,
		validation.Length(0, 50),
		validation.Required,
	); err != nil {
		return errors.New("INVALID_CITY")
	}

	portal.City = city

	return nil
}

func (portal *Portal) ValidateZip(zip string) error {
	if err := validation.Validate(zip,
		validation.Length(0, 15),
		validation.Required,
	); err != nil {
		return errors.New("INVALID_ZIP")
	}

	if err := postcode.Validate(zip); err != nil {
		return errors.New("INVALID_ZIP")
	}

	portal.Zip = zip

	return nil
}

func (portal *Portal) ValidateStreetNumber(streetNumber string) error {
	if err := validation.Validate(streetNumber,
		validation.Length(0, 25),
		validation.Required,
	); err != nil {
		return errors.New("INVALID_STREET_NUMBER")
	}

	portal.StreetNumber = streetNumber

	return nil
}

func (portal *Portal) ValidateStreetName(streetName string) error {
	if err := validation.Validate(streetName,
		validation.Length(0, 50),
		validation.Required,
	); err != nil {
		return errors.New("INVALID_STREET_NAME")
	}

	portal.StreetName = streetName

	return nil
}

func (portal *Portal) ValidateGender(gender string) error {
	if gender != "m" && gender != "f" {
		return errors.New("INVALID_GENDER")
	}

	portal.Gender = gender

	return nil
}

func (portal *Portal) ValidateFamilyStatus(familyStatus string) error {
	if familyStatus != "no" && familyStatus != "m" && familyStatus != "d" && familyStatus != "w" {
		return errors.New("INVALID_FAMILY_STATUS")
	}

	portal.FamilyStatus = familyStatus

	return nil
}

func (portal *Portal) ValidateMaidenName(maidenName string) error {

	if maidenName == "" {
		return nil
	}

	if portal.Gender == "m" || portal.FamilyStatus == "no" {
		return errors.New("CANNOT_SET_MAIDEN_NAME")
	}

	if err := validation.Validate(maidenName,
		validation.Length(1, 30),
		validation.Required,
	); err != nil {
		return errors.New("INVALID_MAIDEN_NAME")
	}

	portal.MaidenName = maidenName

	return nil
}

func (portal *Portal) CreateMember() error {

	db := db.GetDB()

	var err error

	err = portal.GinContextFromContext()
	if err != nil {
		return err
	}

	tx := db.MustBegin()
	tx.MustExec("INSERT INTO Member (FirstName, LastName, Birthday, Citizenship, Gender, FamilyStatus, MaidenName, Phone, Email, PasswordHash, Country, City, Zip, StreetNumber, StreetName, IP, Created, Status) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)", portal.FirstName, portal.LastName, portal.Birthday, portal.Citizenship, portal.Gender, portal.FamilyStatus, portal.MaidenName, portal.Phone, portal.Email, portal.PasswordHash, portal.Country, portal.City, portal.Zip, portal.StreetNumber, portal.StreetName, portal.c.ClientIP(), portal.Timestamp, constants.STATUS_NOACTIVE)
	tx.Commit()

	//Get new member record
	err = portal.QueryMemberByEmail()
	if err != nil {
		return err
	}

	return nil
}

func (portal *Portal) GenerateAuthorizationToken(MemberID int64) (string, error) {
	var err error

	//Create new pairs of refresh and access tokens
	ts, err := portal.ProfileHandler.TK.CreateToken(cast.ToString(MemberID))
	if err != nil {
		return "", err
	}

	//Save the tokens metadata to redis
	// err = portal.ProfileHandler.RD.CreateAuth(cast.ToString(MemberID), ts)
	// if err != nil {
	// 	return "", err
	// }

	return ts.AccessToken, nil
}

func (portal *Portal) VerifyConfirm() error {
	db := db.GetDB()

	//Reset member password
	if portal.Action == constants.ACTION_RESET {
		tx := db.MustBegin()
		tx.MustExec("UPDATE Member SET PasswordHash=$1 WHERE MemberID=$2", portal.PasswordHash, portal.Verify.MemberID)
		tx.MustExec("UPDATE Verify SET Status=$1 WHERE VerifyID=$2", constants.STATUS_SUCCESS, portal.Verify.VerifyID)
		tx.Commit()
	}

	//Confirm member email after visiting sign up confirmation link
	if portal.Action == constants.ACTION_SIGNUP {
		tx := db.MustBegin()
		tx.MustExec("UPDATE Member SET Status=$1 WHERE MemberID=$2", constants.STATUS_ACTIVE, portal.Verify.MemberID)
		tx.MustExec("UPDATE Verify SET Status=$1 WHERE VerifyID=$2", constants.STATUS_SUCCESS, portal.Verify.VerifyID)
		tx.Commit()
	}

	return nil
}

func (portal *Portal) OfferPhoneVerify(Code string) error {
	db := db.GetDB()

	tx := db.MustBegin()
	tx.MustExec("UPDATE Deal SET VerificationCode=$1, Status=$2, DateVerified=CURRENT_TIMESTAMP WHERE DealID=$3", Code, constants.STATUS_SIGNED, portal.Deal.DealID)
	tx.MustExec("UPDATE Verify SET Status=$1 WHERE VerifyID=$2", constants.STATUS_SUCCESS, portal.Verify.VerifyID)
	tx.Commit()

	return nil
}

func (portal *Portal) CreateInvoice() error {
	db := db.GetDB()

	tx := db.MustBegin()
	tx.MustExec("INSERT INTO Invoice (OfferID, MemberID, DealID, CurrencyID, Amount, Status, TimestampCreated) VALUES ($1, $2, $3, $4, $5, $6, $7)", portal.Offer.OfferID, portal.Member.MemberID, portal.Deal.DealID, portal.Offer.CurrencyID, portal.Deal.Amount.Decimal, constants.STATUS_PENDING, portal.Timestamp)
	tx.Commit()

	return nil
}

func (portal *Portal) CancelDeal(RecordID int) error {
	db := db.GetDB()

	tx := db.MustBegin()
	tx.MustExec("UPDATE Deal SET Status=$1 WHERE Status=$2 AND DealID=$3", constants.STATUS_CANCELLED, constants.STATUS_SIGNED, RecordID)
	tx.Commit()

	return nil
}

func (portal *Portal) RemoveDeal(RecordID int) error {
	db := db.GetDB()

	tx := db.MustBegin()
	tx.MustExec("UPDATE Deal SET Status=$1 WHERE Status=$2 AND DealID=$3", constants.STATUS_REMOVED, constants.STATUS_CANCELLED, RecordID)
	tx.Commit()

	return nil
}

func (portal *Portal) CancelInvoiceByDealID(RecordID int) error {
	db := db.GetDB()

	tx := db.MustBegin()
	tx.MustExec("UPDATE Invoice SET Status=$1 WHERE Status=$2 AND DealID=$3", constants.STATUS_CANCELLED, constants.STATUS_PENDING, RecordID)
	tx.Commit()

	return nil
}
