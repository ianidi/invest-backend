package manager

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/gin-gonic/gin"
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/ianidi/exchange-server/graph/methods/constants"
	"github.com/ianidi/exchange-server/graph/model"
	"github.com/ianidi/exchange-server/internal/db"
	"github.com/ianidi/exchange-server/internal/jwt"
	"github.com/ianidi/exchange-server/internal/models"
	"github.com/jackc/pgtype"
	"github.com/jmoiron/sqlx"
	"github.com/spf13/cast"
	"github.com/spf13/viper"
)

type Manager struct {
	c         *gin.Context
	Ctx       context.Context
	Timestamp int64 //Current UNIX timestamp
	Member    models.Member
	Offer     models.Offer
	Contract  models.Contract
	Invoice   models.Invoice
}

func (manager *Manager) GinContextFromContext() error {
	ginContext := manager.Ctx.Value("GinContextKey")
	if ginContext == nil {
		err := fmt.Errorf("could not retrieve gin.Context")
		return err
	}

	gc, ok := ginContext.(*gin.Context)
	if !ok {
		err := fmt.Errorf("gin.Context has wrong type")
		return err
	}

	manager.c = gc

	return nil
}

func (manager *Manager) GetMember() error {

	db := db.GetDB()

	var err error

	err = manager.GinContextFromContext()
	if err != nil {
		return err
	}

	h := jwt.Service

	tokenString := jwt.ExtractToken(manager.c.Request)

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

	manager.Member = member

	return nil
}

func (manager *Manager) ValidateCategoryID(CategoryID string) error {
	db := db.GetDB()

	if err := validation.Validate(CategoryID,
		validation.Required,
	); err != nil {
		return errors.New("INVALID_CATEGORY")
	}

	var count int

	if err := db.Get(&count, "SELECT count(*) FROM Category WHERE CategoryID=$1", CategoryID); err != nil {
		if err != sql.ErrNoRows {
			return err
		}
	}

	if count == 0 {
		return errors.New("INVALID_CATEGORY")
	}

	return nil
}

func (manager *Manager) ValidateTitle(Title string) error {
	if err := validation.Validate(Title,
		validation.Length(1, 200),
		validation.Required,
	); err != nil {
		return errors.New("INVALID_TITLE")
	}
	return nil
}

func (manager *Manager) ValidateSubtitle(Subtitle string) error {
	if err := validation.Validate(Subtitle,
		validation.Required,
	); err != nil {
		return errors.New("INVALID_SUBTITLE")
	}
	return nil
}

func (manager *Manager) ValidateDescription(Description string) error {
	if err := validation.Validate(Description,
		validation.Required,
	); err != nil {
		return errors.New("INVALID_DESCRIPTION")
	}
	return nil
}

func (manager *Manager) ValidateCurrencyID(CurrencyID string) error {

	db := db.GetDB()

	if err := validation.Validate(CurrencyID,
		validation.Required,
	); err != nil {
		return errors.New("INVALID_CURRENCY")
	}

	var count int

	if err := db.Get(&count, "SELECT count(*) FROM Currency WHERE CurrencyID=$1", CurrencyID); err != nil {
		if err != sql.ErrNoRows {
			return err
		}
	}

	if count == 0 {
		return errors.New("INVALID_CURRENCY")
	}

	return nil
}

func (manager *Manager) QueryInvestList() ([]*model.Invest, error) {
	db := db.GetDB()

	var err error

	var invest []models.Invest

	if err = db.Select(&invest, "SELECT * FROM Invest ORDER BY InvestID DESC"); err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("NO_INVEST_RECORD")
		}
		return nil, err
	}

	var res []*model.Invest

	for _, row := range invest {

		InvestID := cast.ToInt(row.InvestID)
		CategoryID := cast.ToInt(row.CategoryID)
		Title := row.Title
		Subtitle := row.Subtitle
		Description := row.Description

		resRow := &model.Invest{
			InvestID:    InvestID,
			Title:       &Title,
			CategoryID:  &CategoryID,
			Subtitle:    &Subtitle,
			Description: &Description,
		}

		res = append(res, resRow)
	}

	return res, nil
}

func (manager *Manager) QueryContractList() ([]*model.Contract, error) {
	db := db.GetDB()

	var err error

	var Contract []models.Contract

	if err = db.Select(&Contract, "SELECT * FROM Contract WHERE Template=$1 ORDER BY ContractID DESC", true); err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("NO_CONTRACT_RECORD")
		}
		return nil, err
	}

	var res []*model.Contract

	for _, row := range Contract {

		ContractID := cast.ToInt(row.ContractID)
		OfferID := cast.ToInt(row.OfferID)
		Title := row.Title
		ContentRaw := row.ContentRaw
		Current := row.Current
		Template := row.Template

		resRow := &model.Contract{
			ContractID: ContractID,
			OfferID:    &OfferID,
			Title:      Title,
			ContentRaw: &ContentRaw,
			Current:    &Current,
			Template:   &Template,
		}

		res = append(res, resRow)
	}

	return res, nil
}

func (manager *Manager) EditContract(input model.ManagerEditContractRequest) (*model.Result, error) {
	db := db.GetDB()

	tx := db.MustBegin()
	tx.MustExec("UPDATE Contract SET ContentRaw=$1, Title=$2 WHERE ContractID=$3", input.ContentRaw, input.Title, input.ContractID)
	tx.Commit()

	return &model.Result{
		Status: true,
	}, nil
}

func (manager *Manager) ValidateInvoiceStatus(Status string) error {

	if Status != constants.STATUS_PAID && Status != constants.STATUS_PENDING {
		return errors.New("INVALID_STATUS")
	}

	return nil
}

func (manager *Manager) EditInvoice(input model.ManagerEditInvoiceRequest) (*model.Result, error) {
	db := db.GetDB()

	if input.Status == constants.STATUS_PAID {
		tx := db.MustBegin()
		tx.MustExec("UPDATE Invoice SET Status=$1, DatePaid=CURRENT_TIMESTAMP, TimestampPaid=$2 WHERE InvoiceID=$3", input.Status, manager.Timestamp, input.InvoiceID)
		tx.MustExec("UPDATE Deal SET Status=$1, DatePaid=CURRENT_TIMESTAMP WHERE DealID=$2", input.Status, manager.Invoice.DealID)
		tx.Commit()
	}

	if input.Status == constants.STATUS_PENDING {
		tx := db.MustBegin()
		tx.MustExec("UPDATE Invoice SET Status=$1, DatePaid='infinity'::timestamptz, TimestampPaid=$2 WHERE InvoiceID=$3", input.Status, 0, input.InvoiceID)
		tx.MustExec("UPDATE Deal SET Status=$1, DatePaid='infinity'::timestamptz WHERE DealID=$2", constants.STATUS_SIGNED, manager.Invoice.DealID)
		tx.Commit()
	}

	return &model.Result{
		Status: true,
	}, nil
}

func (manager *Manager) EditInvest(input model.ManagerEditInvestRequest) (*model.Result, error) {
	db := db.GetDB()

	tx := db.MustBegin()
	tx.MustExec("UPDATE Invest SET CategoryID=$1, Title=$2, Subtitle=$3, Description=$4 WHERE InvestID=$5", input.CategoryID, input.Title, input.Subtitle, input.Description, input.InvestID)
	tx.Commit()

	return &model.Result{
		Status: true,
	}, nil
}

func (manager *Manager) EditOffer(input model.ManagerEditOfferRequest) (*model.Result, error) {
	db := db.GetDB()

	tx := db.MustBegin()
	tx.MustExec("UPDATE Offer SET CurrencyID=$1, Title=$2 WHERE OfferID=$3", input.CurrencyID, input.Title, input.OfferID)
	tx.Commit()

	return &model.Result{
		Status: true,
	}, nil
}

func (manager *Manager) EditLead(input model.ManagerEditLeadRequest) (*model.Result, error) {
	db := db.GetDB()

	Birthday := input.Birthday

	if len(*Birthday) == 0 {
		Birthday = nil
	}

	tx := db.MustBegin()
	tx.MustExec("UPDATE Lead SET MemberID=$1, CampaignID=$2, CurrencyID=$3, Email=$4, Phone=$5, FirstName=$6, LastName=$7, Gender=$8, FamilyStatus=$9, MaidenName=$10, Citizenship=$11, Country=$12, City=$13, Zip=$14, Address1=$15, Address2=$16, StreetNumber=$17, StreetName=$18, Birthday=$19, Status=$20 WHERE LeadID=$21", input.MemberID, input.CampaignID, input.CurrencyID, input.Email, input.Phone, input.FirstName, input.LastName, input.Gender, input.FamilyStatus, input.MaidenName, input.Citizenship, input.Country, input.City, input.Zip, input.Address1, input.Address2, input.StreetNumber, input.StreetName, Birthday, input.Status, input.LeadID)
	tx.Commit()

	return &model.Result{
		Status: true,
	}, nil
}

func (manager *Manager) EditComment(input model.ManagerEditCommentRequest) (*model.Result, error) {
	db := db.GetDB()

	tx := db.MustBegin()
	tx.MustExec("UPDATE Comment SET Content=$1, DateEdited=CURRENT_TIMESTAMP, TimestampEdited=$2, LeadID=$3 WHERE CommentID=$4", input.Content, manager.Timestamp, input.CommentID, input.LeadID)
	tx.Commit()

	return &model.Result{
		Status: true,
	}, nil
}

func (manager *Manager) EditChecklist(input model.ManagerEditChecklistRequest) (*model.Result, error) {
	db := db.GetDB()

	tx := db.MustBegin()
	tx.MustExec("UPDATE Checklist SET Title=$1, Complete=$2, Position=$3, LeadID=$4 WHERE ChecklistID=$5", input.Title, input.Complete, input.Position, input.LeadID, input.ChecklistID)
	tx.Commit()

	return &model.Result{
		Status: true,
	}, nil
}

func (manager *Manager) EditAppointment(input model.ManagerEditAppointmentRequest) (*model.Result, error) {
	db := db.GetDB()

	tx := db.MustBegin()
	tx.MustExec("UPDATE Appointment SET Title=$1, Description=$2, DateDue=$3, Status=$4, LeadID=$5 WHERE AppointmentID=$6", input.Title, input.Description, input.DateDue, input.Status, input.LeadID, input.AppointmentID)
	tx.Commit()

	//, TimestampDue=$4

	return &model.Result{
		Status: true,
	}, nil
}

func (manager *Manager) EditCampaign(input model.ManagerEditCampaignRequest) (*model.Result, error) {
	db := db.GetDB()

	tx := db.MustBegin()
	tx.MustExec("UPDATE Campaign SET Title=$1, Description=$2 WHERE CampaignID=$3", input.Title, input.Description, input.CampaignID)
	tx.Commit()

	return &model.Result{
		Status: true,
	}, nil
}

func (manager *Manager) EditMedia(input model.ManagerEditMediaRequest) (*model.Result, error) {
	db := db.GetDB()

	tx := db.MustBegin()
	tx.MustExec("UPDATE Media SET Title=$1 WHERE MediaID=$2", input.Title, input.MediaID)
	tx.Commit()

	return &model.Result{
		Status: true,
	}, nil
}

func (manager *Manager) RemoveLead(input model.RecordRequest) (*model.Result, error) {
	db := db.GetDB()

	tx := db.MustBegin()
	tx.MustExec("DELETE FROM Lead WHERE LeadID=$1", input.RecordID)
	tx.Commit()

	return &model.Result{
		Status: true,
	}, nil
}

func (manager *Manager) RemoveComment(input model.RecordRequest) (*model.Result, error) {
	db := db.GetDB()

	tx := db.MustBegin()
	tx.MustExec("DELETE FROM Comment WHERE CommentID=$1", input.RecordID)
	tx.Commit()

	return &model.Result{
		Status: true,
	}, nil
}

func (manager *Manager) RemoveChecklist(input model.RecordRequest) (*model.Result, error) {
	db := db.GetDB()

	tx := db.MustBegin()
	tx.MustExec("DELETE FROM Checklist WHERE ChecklistID=$1", input.RecordID)
	tx.Commit()

	return &model.Result{
		Status: true,
	}, nil
}

func (manager *Manager) RemoveAppointment(input model.RecordRequest) (*model.Result, error) {
	db := db.GetDB()

	tx := db.MustBegin()
	tx.MustExec("DELETE FROM Appointment WHERE AppointmentID=$1", input.RecordID)
	tx.Commit()

	return &model.Result{
		Status: true,
	}, nil
}

func (manager *Manager) RemoveCampaign(input model.RecordRequest) (*model.Result, error) {
	db := db.GetDB()

	tx := db.MustBegin()
	tx.MustExec("DELETE FROM Campaign WHERE CampaignID=$1", input.RecordID)
	tx.Commit()

	return &model.Result{
		Status: true,
	}, nil
}

func (manager *Manager) RemoveInterest(input model.RecordRequest) (*model.Result, error) {
	db := db.GetDB()

	tx := db.MustBegin()
	tx.MustExec("DELETE FROM Interest WHERE InterestID=$1", input.RecordID)
	tx.Commit()

	return &model.Result{
		Status: true,
	}, nil
}

func (manager *Manager) RemoveManager(input model.RecordRequest) (*model.Result, error) {
	db := db.GetDB()

	tx := db.MustBegin()
	tx.MustExec("UPDATE Member SET ManagerRole=$1 WHERE MemberID=$2", "", input.RecordID)
	tx.Commit()

	return &model.Result{
		Status: true,
	}, nil
}

func (manager *Manager) RemoveMedia(input model.RecordRequest) (*model.Result, error) {
	db := db.GetDB()

	tx := db.MustBegin()
	tx.MustExec("DELETE FROM Media WHERE MediaID=$1", input.RecordID)
	tx.Commit()

	//TODO: remove file from cloud
	//TODO: remove upload record

	return &model.Result{
		Status: true,
	}, nil
}

func (manager *Manager) DragMedia(input model.DragRequest) (*model.Result, error) {
	db := db.GetDB()

	tx := db.MustBegin()
	for index, RecordID := range input.Position {
		tx.MustExec("UPDATE Media SET Position=$1 WHERE MediaID=$2", index, RecordID)
	}
	tx.Commit()

	return &model.Result{
		Status: true,
	}, nil
}

func (manager *Manager) AssignManager(input model.ManagerAssignManagerRequest) (*model.Result, error) {
	db := db.GetDB()

	tx := db.MustBegin()
	tx.MustExec("UPDATE Member SET ManagerRole=$1 WHERE MemberID=$2", "manager", input.MemberID)
	tx.Commit()

	return &model.Result{
		Status: true,
	}, nil
}

func (manager *Manager) EditBankDetails(input model.ManagerEditBankDetailsRequest) (*model.Result, error) {
	db := db.GetDB()

	tx := db.MustBegin()
	tx.MustExec("UPDATE BankDetails SET Title=$1, BeneficiaryCompany=$2, BeneficiaryFirstName=$3, BeneficiaryLastName=$4, BeneficiaryCountry=$5, BeneficiaryCity=$6, BeneficiaryZip=$7, BeneficiaryAddress=$8, BankName=$9, BankBranch=$10, BankIFSC=$11, BankBranchCountry=$12, BankBranchCity=$13, BankBranchZip=$14, BankBranchAddress=$15, BankAccountNumber=$16, BankAccountType=$17, BankRoutingNumber=$18, BankTransferCaption=$19, BankIBAN=$20, BankSWIFT=$21, BankSWIFTCorrespondent=$22, BankBIC=$23 WHERE BankDetailsID=$24", input.Title, input.BeneficiaryCompany, input.BeneficiaryFirstName, input.BeneficiaryLastName, input.BeneficiaryCountry, input.BeneficiaryCity, input.BeneficiaryZip, input.BeneficiaryAddress, input.BankName, input.BankBranch, input.BankIfsc, input.BankBranchCountry, input.BankBranchCity, input.BankBranchZip, input.BankBranchAddress, input.BankAccountNumber, input.BankAccountType, input.BankRoutingNumber, input.BankTransferCaption, input.BankIban, input.BankSwift, input.BankSWIFTCorrespondent, input.BankBic, input.BankDetailsID)
	tx.Commit()

	return &model.Result{
		Status: true,
	}, nil
}

func (manager *Manager) QueryBankDetailsList() ([]*model.BankDetails, error) {
	db := db.GetDB()

	var err error

	var bankDetails []models.BankDetails

	if err = db.Select(&bankDetails, "SELECT * FROM BankDetails ORDER BY BankDetailsID DESC"); err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("NO_BANK_DETAILS_RECORD")
		}
		return nil, err
	}

	var res []*model.BankDetails

	for _, row := range bankDetails {

		BankDetailsID := cast.ToInt(row.BankDetailsID)
		Title := row.Title
		BeneficiaryCompany := row.BeneficiaryCompany
		BeneficiaryFirstName := row.BeneficiaryFirstName
		BeneficiaryLastName := row.BeneficiaryLastName
		BeneficiaryCountry := row.BeneficiaryCountry
		BeneficiaryCity := row.BeneficiaryCity
		BeneficiaryZip := row.BeneficiaryZip
		BeneficiaryAddress := row.BeneficiaryAddress
		BankName := row.BankName
		BankBranch := row.BankBranch
		BankIFSC := row.BankIFSC
		BankBranchCountry := row.BankBranchCountry
		BankBranchCity := row.BankBranchCity
		BankBranchZip := row.BankBranchZip
		BankBranchAddress := row.BankBranchAddress
		BankAccountNumber := row.BankAccountNumber
		BankAccountType := row.BankAccountType
		BankRoutingNumber := row.BankRoutingNumber
		BankTransferCaption := row.BankTransferCaption
		BankIBAN := row.BankIBAN
		BankSWIFT := row.BankSWIFT
		BankSWIFTCorrespondent := row.BankSWIFTCorrespondent
		BankBIC := row.BankBIC

		resRow := &model.BankDetails{
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

		res = append(res, resRow)
	}

	return res, nil
}

func (manager *Manager) SearchManager() ([]*model.ManagerSearch, error) {
	db := db.GetDB()

	var err error

	var Manager []models.Member

	if err = db.Select(&Manager, "SELECT * FROM Member WHERE (ManagerRole=$1 OR ManagerRole=$2) ORDER BY MemberID DESC", "manager", "admin"); err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("NO_MANAGER_RECORD")
		}
		return nil, err
	}

	var res []*model.ManagerSearch

	for _, row := range Manager {

		ManagerID := cast.ToInt(row.MemberID)
		Title := row.FirstName + " " + row.LastName + ", " + row.ManagerRole + " (ID " + cast.ToString(row.MemberID) + ")"

		resRow := &model.ManagerSearch{
			ManagerID: ManagerID,
			Title:     &Title,
		}

		res = append(res, resRow)
	}

	return res, nil
}

func (manager *Manager) DealListByOfferID(input model.RecordRequest) ([]*model.Deal, error) {
	db := db.GetDB()

	var err error

	var Deal []models.Deal

	if err = db.Select(&Deal, "SELECT * FROM Deal WHERE OfferID=$1 ORDER BY DealID DESC", input.RecordID); err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("NO_DEAL_RECORD")
		}
		return nil, err
	}

	var res []*model.Deal

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
		DateVerified := row.DateVerified.Time.String()
		DateSigned := row.DateSigned.Time.String()
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
			DateVerified:      &DateVerified,
			DateSigned:        &DateSigned,
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

func (manager *Manager) DealByContractID(input *model.ManagerDealByContractIDRequest) (*model.Deal, error) {
	db := db.GetDB()

	var err error

	var row models.Deal

	if err = db.Get(&row, "SELECT * FROM Deal WHERE OfferID=$1 AND ContractID=$2 ORDER BY DealID DESC", input.OfferID, input.ContractID); err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("NO_DEAL_RECORD")
		}
		return nil, err
	}

	DealID := cast.ToInt(row.DealID)
	OfferID := cast.ToInt(row.OfferID)
	ContractID := cast.ToInt(row.ContractID)
	MemberID := cast.ToInt(row.MemberID)
	CurrencyID := cast.ToInt(row.CurrencyID)
	SignatureFilename := row.SignatureFilename
	SignatureURL := viper.GetString("s3_cdn_url") + row.SignatureFilename
	VerificationCode := row.VerificationCode
	DateCreated := row.DateCreated.Time.String()
	DateVerified := row.DateVerified.Time.String()
	DateSigned := row.DateSigned.Time.String()
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
		DateVerified:      &DateVerified,
		DateSigned:        &DateSigned,
		DatePaid:          &DatePaid,
		DateStart:         &DateStart,
		DateEnd:           &DateEnd,
		Status:            &Status,
		Amount:            &Amount,
		Duration:          &Duration,
	}

	return res, nil
}

func (manager *Manager) QueryContractListByOfferID(RecordID int) ([]*model.Contract, error) {
	db := db.GetDB()

	var err error

	var Contract []models.Contract

	if err = db.Select(&Contract, "SELECT * FROM Contract WHERE OfferID=$1 ORDER BY ContractID DESC", RecordID); err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("NO_CONTRACT_RECORD")
		}
		return nil, err
	}

	var res []*model.Contract

	for _, row := range Contract {

		ContractID := cast.ToInt(row.ContractID)
		OfferID := cast.ToInt(row.OfferID)
		Title := row.Title
		ContentRaw := row.ContentRaw
		Current := row.Current
		Template := row.Template

		resRow := &model.Contract{
			ContractID: ContractID,
			OfferID:    &OfferID,
			Title:      Title,
			ContentRaw: &ContentRaw,
			Current:    &Current,
			Template:   &Template,
		}

		res = append(res, resRow)
	}

	return res, nil
}

func (manager *Manager) QueryCommentListByLeadID(RecordID int) ([]*model.Comment, error) {
	db := db.GetDB()

	var err error

	var comment []models.Comment

	if err = db.Select(&comment, "SELECT * FROM Comment WHERE LeadID=$1 ORDER BY CommentID DESC", RecordID); err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("NO_COMMENT_RECORD")
		}
		return nil, err
	}

	var res []*model.Comment

	for _, row := range comment {

		CommentID := cast.ToInt(row.CommentID.Int)
		MemberID := cast.ToInt(row.MemberID.Int)
		Content := row.Content.String
		DateCreated := row.DateCreated.Time.String()
		TimestampCreated := cast.ToInt(row.TimestampCreated.Int)
		DateEdited := row.DateEdited.Time.String()
		TimestampEdited := cast.ToInt(row.TimestampEdited.Int)
		LeadID := cast.ToInt(row.LeadID.Int)

		resRow := &model.Comment{
			CommentID:        CommentID,
			MemberID:         &MemberID,
			Content:          &Content,
			DateCreated:      &DateCreated,
			TimestampCreated: &TimestampCreated,
			DateEdited:       &DateEdited,
			TimestampEdited:  &TimestampEdited,
			LeadID:           &LeadID,
		}

		res = append(res, resRow)
	}

	return res, nil
}

func (manager *Manager) QueryAppointmentListByLeadID(RecordID int) ([]*model.Appointment, error) {
	db := db.GetDB()

	var err error

	var Appointment []models.Appointment

	if err = db.Select(&Appointment, "SELECT * FROM Appointment WHERE LeadID=$1 ORDER BY AppointmentID DESC", RecordID); err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("NO_APPOINTMENT_RECORD")
		}
		return nil, err
	}

	var res []*model.Appointment

	for _, row := range Appointment {

		AppointmentID := cast.ToInt(row.AppointmentID.Int)
		Type := row.Type.String
		Title := row.Title.String
		Description := row.Description.String
		DateCreated := row.DateCreated.Time.String()
		TimestampCreated := cast.ToInt(row.TimestampCreated.Int)
		DateDue := row.DateDue.Time.String()
		TimestampDue := cast.ToInt(row.TimestampDue.Int)
		Status := row.Status.String
		LeadID := cast.ToInt(row.LeadID.Int)

		resRow := &model.Appointment{
			AppointmentID:    AppointmentID,
			Type:             &Type,
			Title:            &Title,
			Description:      &Description,
			DateCreated:      &DateCreated,
			TimestampCreated: &TimestampCreated,
			DateDue:          &DateDue,
			TimestampDue:     &TimestampDue,
			Status:           &Status,
			LeadID:           &LeadID,
		}

		res = append(res, resRow)
	}

	return res, nil
}

func (manager *Manager) QueryInterestListByOfferID(RecordID int) ([]*model.Interest, error) {
	db := db.GetDB()

	var err error

	var Interest []models.Interest

	if err = db.Select(&Interest, "SELECT * FROM Interest WHERE OfferID=$1 ORDER BY DurationFrom ASC", RecordID); err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("NO_INTEREST_RECORD")
		}
		return nil, err
	}

	var res []*model.Interest

	for _, row := range Interest {

		InterestID := cast.ToInt(row.InterestID)
		OfferID := cast.ToInt(row.OfferID)
		AmountFrom := row.AmountFrom.Decimal.String()
		AmountTo := row.AmountTo.Decimal.String()
		DurationFrom := row.DurationFrom.Decimal.String()
		DurationTo := row.DurationTo.Decimal.String()
		Interest := row.Interest.Decimal.String()

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

func (manager *Manager) QueryBankDetailsByOfferID() (*model.BankDetails, error) {
	db := db.GetDB()

	var err error

	var row models.BankDetails

	if err = db.Get(&row, "SELECT * FROM BankDetails WHERE BankDetailsID=$1", manager.Offer.BankDetailsID); err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("NO_BANK_DETAILS_RECORD")
		}
		return nil, err
	}

	BankDetailsID := cast.ToInt(row.BankDetailsID)
	Title := row.Title
	BeneficiaryCompany := row.BeneficiaryCompany
	BeneficiaryFirstName := row.BeneficiaryFirstName
	BeneficiaryLastName := row.BeneficiaryLastName
	BeneficiaryCountry := row.BeneficiaryCountry
	BeneficiaryCity := row.BeneficiaryCity
	BeneficiaryZip := row.BeneficiaryZip
	BeneficiaryAddress := row.BeneficiaryAddress
	BankName := row.BankName
	BankBranch := row.BankBranch
	BankIFSC := row.BankIFSC
	BankBranchCountry := row.BankBranchCountry
	BankBranchCity := row.BankBranchCity
	BankBranchZip := row.BankBranchZip
	BankBranchAddress := row.BankBranchAddress
	BankAccountNumber := row.BankAccountNumber
	BankAccountType := row.BankAccountType
	BankRoutingNumber := row.BankRoutingNumber
	BankTransferCaption := row.BankTransferCaption
	BankIBAN := row.BankIBAN
	BankSWIFT := row.BankSWIFT
	BankSWIFTCorrespondent := row.BankSWIFTCorrespondent
	BankBIC := row.BankBIC

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

func (manager *Manager) QueryInvestByOfferID() (*model.Invest, error) {
	db := db.GetDB()

	var err error

	var row models.Invest

	if err = db.Get(&row, "SELECT * FROM Invest WHERE InvestID=$1", manager.Offer.InvestID); err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("NO_INVEST_RECORD")
		}
		return nil, err
	}

	InvestID := cast.ToInt(row.InvestID)
	CategoryID := cast.ToInt(row.CategoryID)
	Title := row.Title
	Subtitle := row.Subtitle
	Description := row.Description

	res := &model.Invest{
		InvestID:    InvestID,
		Title:       &Title,
		CategoryID:  &CategoryID,
		Subtitle:    &Subtitle,
		Description: &Description,
	}

	return res, nil
}

func (manager *Manager) QueryMediaByInvestID(input model.ManagerMediaByInvestIDRequest) ([]*model.Media, error) {
	db := db.GetDB()

	var err error

	var res []*model.Media
	var Media []models.Media

	if err = db.Select(&Media, "SELECT * FROM Media WHERE InvestID=$1 AND Category=$2 ORDER BY Position ASC", input.RecordID, input.Category); err != nil {
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

func (manager *Manager) QueryInvoiceByDealID(input *model.RecordRequest) (*model.Invoice, error) {
	db := db.GetDB()

	var err error

	var row models.Invoice

	if err = db.Get(&row, "SELECT * FROM Invoice WHERE DealID=$1", input.RecordID); err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("NO_INVOICE_RECORD")
		}
		return nil, err
	}

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

func (manager *Manager) QueryInvoice(RecordID int) (*model.Invoice, error) {
	db := db.GetDB()

	var err error

	var row models.Invoice

	if err = db.Get(&row, "SELECT * FROM Invoice WHERE InvoiceID=$1", RecordID); err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("NO_INVOICE_RECORD")
		}
		return nil, err
	}

	manager.Invoice = row

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

func (manager *Manager) QueryFAQArray(IDs []pgtype.Int8) ([]*model.Faq, error) {
	if len(IDs) == 0 {
		return nil, nil
	}

	db := db.GetDB()

	var err error

	var res []*model.Faq
	var faq []models.FAQ

	query, args, err := sqlx.In("SELECT * FROM FAQ WHERE FAQID IN (?) ORDER BY Position DESC", IDs)
	query = db.Rebind(query)

	if err = db.Select(&faq, query, args...); err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("NO_FAQ_RECORD")
		}
		return nil, err
	}

	for _, faqRow := range faq {

		FAQID := cast.ToInt(faqRow.FAQID)
		Question := faqRow.Question
		Answer := faqRow.Answer
		Position := cast.ToInt(faqRow.Position)

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

func (manager *Manager) QueryUploadArray(IDs []pgtype.Int8) ([]*model.Upload, error) {
	if len(IDs) == 0 {
		return nil, nil
	}

	db := db.GetDB()

	var err error

	var res []*model.Upload
	var upload []models.Upload

	query, args, err := sqlx.In("SELECT * FROM Upload WHERE UploadID IN (?)", IDs)

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
		MemberID := cast.ToInt(uploadRow.MemberID)
		URL := viper.GetString("s3_cdn_url") + uploadRow.Filename
		Filename := uploadRow.Filename
		Category := uploadRow.Category
		Created := cast.ToInt(uploadRow.Created)

		resRow := &model.Upload{
			UploadID: UploadID,
			MemberID: &MemberID,
			URL:      &URL,
			Filename: &Filename,
			Category: &Category,
			Created:  &Created,
		}

		res = append(res, resRow)
	}

	return res, nil
}

func (manager *Manager) QueryMediaArray1(IDs []pgtype.Int8) ([]*model.Media, error) {
	if len(IDs) == 0 {
		return nil, nil
	}

	db := db.GetDB()

	var err error

	var res []*model.Media
	var media []models.Media

	query, args, err := sqlx.In("SELECT Media.MediaID, Media.UploadID, Media.Title, Media.Position, Upload.Filename, Upload.Category, Upload.Created FROM Media LEFT OUTER JOIN Upload ON Media.UploadID=Upload.UploadID WHERE Media.MediaID IN (?)", IDs)

	// err = db.QueryRow("select cardinality($1::text[])", []string{"a", "b", "c"}).Scan(&n)
	// if err != nil {
	// 	return nil, errors.New("NO_MEDIA_RECORD")
	// }

	// sqlx.In returns queries with the `?` bindvar, we can rebind it for our backend
	// query = db.Rebind(query)

	if err = db.Select(&media, query, args...); err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("NO_MEDIA_RECORD")
		}
		return nil, err
	}

	for _, row := range media {

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

func (manager *Manager) QueryMediaArray(InvestID int64, Category string) ([]*model.Media, error) {
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

func (manager *Manager) QueryInvest(RecordID int) (*model.Invest, error) {
	db := db.GetDB()

	var err error

	var row models.Invest

	if err = db.Get(&row, "SELECT * FROM Invest WHERE InvestID=$1", RecordID); err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("NO_INVEST_RECORD")
		}
		return nil, err
	}

	InvestID := cast.ToInt(row.InvestID)
	CategoryID := cast.ToInt(row.CategoryID)
	Title := row.Title
	Subtitle := row.Subtitle
	Description := row.Description

	Photo, _ := manager.QueryMediaArray(row.InvestID, constants.CATEGORY_PHOTO)
	Document, _ := manager.QueryMediaArray(row.InvestID, constants.CATEGORY_DOCUMENT)
	// FAQ, _ := manager.QueryFAQArray(row.FAQ.Elements)

	res := &model.Invest{
		InvestID:    InvestID,
		Title:       &Title,
		CategoryID:  &CategoryID,
		Subtitle:    &Subtitle,
		Description: &Description,
		Photo:       Photo,
		Document:    Document,
		// Faq:         FAQ,
	}

	return res, nil
}

func (manager *Manager) QueryContract(RecordID int) (*model.Contract, error) {
	db := db.GetDB()

	var err error

	var row models.Contract

	if err = db.Get(&row, "SELECT * FROM Contract WHERE ContractID=$1", RecordID); err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("NO_CONTRACT_RECORD")
		}
		return nil, err
	}

	manager.Contract = row

	ContractID := cast.ToInt(row.ContractID)
	OfferID := cast.ToInt(row.OfferID)
	Title := row.Title
	ContentRaw := row.ContentRaw
	Current := row.Current
	Template := row.Template

	res := &model.Contract{
		ContractID: ContractID,
		OfferID:    &OfferID,
		Title:      Title,
		ContentRaw: &ContentRaw,
		Current:    &Current,
		Template:   &Template,
	}

	return res, nil
}

func (manager *Manager) QueryLead(RecordID int) (*model.Lead, error) {
	db := db.GetDB()

	var err error

	var row models.Lead

	if err = db.Get(&row, "SELECT * FROM Lead WHERE LeadID=$1", RecordID); err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("NO_LEAD_RECORD")
		}
		return nil, err
	}

	LeadID := cast.ToInt(row.LeadID.Int)
	ManagerID := cast.ToInt(row.ManagerID.Int)
	MemberID := cast.ToInt(row.MemberID.Int)
	CampaignID := cast.ToInt(row.CampaignID.Int)
	CurrencyID := cast.ToInt(row.CurrencyID.Int)
	Email := row.Email.String
	Phone := row.Phone.String
	IP := row.IP.String
	FirstName := row.FirstName.String
	LastName := row.LastName.String
	Gender := row.Gender.String
	FamilyStatus := row.FamilyStatus.String
	MaidenName := row.MaidenName.String
	Citizenship := row.Citizenship.String
	Country := row.Country.String
	City := row.City.String
	Zip := row.Zip.String
	Address1 := row.Address1.String
	Address2 := row.Address2.String
	StreetNumber := row.StreetNumber.String
	StreetName := row.StreetName.String
	Birthday := row.Birthday.Time.String()
	Status := row.Status.String
	DateCreated := row.DateCreated.Time.String()
	TimestampCreated := cast.ToInt(row.TimestampCreated.Int)

	res := &model.Lead{
		LeadID:           LeadID,
		ManagerID:        &ManagerID,
		MemberID:         &MemberID,
		CampaignID:       &CampaignID,
		CurrencyID:       &CurrencyID,
		Email:            &Email,
		Phone:            &Phone,
		IP:               &IP,
		FirstName:        &FirstName,
		LastName:         &LastName,
		Gender:           &Gender,
		FamilyStatus:     &FamilyStatus,
		MaidenName:       &MaidenName,
		Citizenship:      &Citizenship,
		Country:          &Country,
		City:             &City,
		Zip:              &Zip,
		Address1:         &Address1,
		Address2:         &Address2,
		StreetNumber:     &StreetNumber,
		StreetName:       &StreetName,
		Birthday:         &Birthday,
		Status:           &Status,
		DateCreated:      &DateCreated,
		TimestampCreated: &TimestampCreated,
	}

	return res, nil
}

func (manager *Manager) QueryComment(RecordID int) (*model.Comment, error) {
	db := db.GetDB()

	var err error

	var row models.Comment

	if err = db.Get(&row, "SELECT * FROM Comment WHERE CommentID=$1", RecordID); err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("NO_COMMENT_RECORD")
		}
		return nil, err
	}

	CommentID := cast.ToInt(row.CommentID.Int)
	MemberID := cast.ToInt(row.MemberID.Int)
	Content := row.Content.String
	DateCreated := row.DateCreated.Time.String()
	TimestampCreated := cast.ToInt(row.TimestampCreated.Int)
	DateEdited := row.DateEdited.Time.String()
	TimestampEdited := cast.ToInt(row.TimestampEdited.Int)
	LeadID := cast.ToInt(row.LeadID.Int)

	res := &model.Comment{
		CommentID:        CommentID,
		MemberID:         &MemberID,
		Content:          &Content,
		DateCreated:      &DateCreated,
		TimestampCreated: &TimestampCreated,
		DateEdited:       &DateEdited,
		TimestampEdited:  &TimestampEdited,
		LeadID:           &LeadID,
	}

	return res, nil
}

func (manager *Manager) QueryChecklist(RecordID int) (*model.Checklist, error) {
	db := db.GetDB()

	var err error

	var row models.Checklist

	if err = db.Get(&row, "SELECT * FROM Checklist WHERE ChecklistID=$1", RecordID); err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("NO_CHECKLIST_RECORD")
		}
		return nil, err
	}

	ChecklistID := cast.ToInt(row.ChecklistID.Int)
	Title := row.Title.String
	Complete := row.Complete.Bool
	Position := cast.ToInt(row.Position.Int)
	DateCreated := row.DateCreated.Time.String()
	TimestampCreated := cast.ToInt(row.TimestampCreated.Int)
	LeadID := cast.ToInt(row.LeadID.Int)

	res := &model.Checklist{
		ChecklistID:      ChecklistID,
		Title:            &Title,
		Complete:         &Complete,
		Position:         &Position,
		DateCreated:      &DateCreated,
		TimestampCreated: &TimestampCreated,
		LeadID:           &LeadID,
	}

	return res, nil
}

func (manager *Manager) QueryAppointment(RecordID int) (*model.Appointment, error) {
	db := db.GetDB()

	var err error

	var row models.Appointment

	if err = db.Get(&row, "SELECT * FROM Appointment WHERE AppointmentID=$1", RecordID); err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("NO_APPOINTMENT_RECORD")
		}
		return nil, err
	}

	AppointmentID := cast.ToInt(row.AppointmentID.Int)
	Type := row.Type.String
	Title := row.Title.String
	Description := row.Description.String
	DateCreated := row.DateCreated.Time.String()
	TimestampCreated := cast.ToInt(row.TimestampCreated.Int)
	DateDue := row.DateDue.Time.String()
	TimestampDue := cast.ToInt(row.TimestampDue.Int)
	Status := row.Status.String

	res := &model.Appointment{
		AppointmentID:    AppointmentID,
		Type:             &Type,
		Title:            &Title,
		Description:      &Description,
		DateCreated:      &DateCreated,
		TimestampCreated: &TimestampCreated,
		DateDue:          &DateDue,
		TimestampDue:     &TimestampDue,
		Status:           &Status,
	}

	return res, nil
}

func (manager *Manager) QueryCampaign(RecordID int) (*model.Campaign, error) {
	db := db.GetDB()

	var err error

	var row models.Campaign

	if err = db.Get(&row, "SELECT * FROM Campaign WHERE CampaignID=$1", RecordID); err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("NO_CAMPAIGN_RECORD")
		}
		return nil, err
	}

	CampaignID := cast.ToInt(row.CampaignID.Int)
	Title := row.Title.String
	Description := row.Description.String
	DateCreated := row.DateCreated.Time.String()
	TimestampCreated := cast.ToInt(row.TimestampCreated.Int)

	res := &model.Campaign{
		CampaignID:       CampaignID,
		Title:            &Title,
		Description:      &Description,
		DateCreated:      &DateCreated,
		TimestampCreated: &TimestampCreated,
	}

	return res, nil
}

func (manager *Manager) QueryBankDetails(RecordID int) (*model.BankDetails, error) {
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

func (manager *Manager) QueryOffer(RecordID int) (*model.ManagerOffer, error) {
	db := db.GetDB()

	var err error

	var row models.Offer

	if err = db.Get(&row, "SELECT * FROM Offer WHERE OfferID=$1", RecordID); err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("NO_OFFER_RECORD")
		}
		return nil, err
	}

	OfferID := cast.ToInt(row.OfferID)
	MemberID := cast.ToInt(row.MemberID)
	InvestID := cast.ToInt(row.InvestID)
	CurrencyID := cast.ToInt(row.CurrencyID.Int)
	BankDetailsID := cast.ToInt(row.BankDetailsID)
	Title := row.Title
	Status := row.Status

	res := &model.ManagerOffer{
		OfferID:       OfferID,
		MemberID:      &MemberID,
		InvestID:      &InvestID,
		CurrencyID:    &CurrencyID,
		BankDetailsID: &BankDetailsID,
		Title:         &Title,
		Status:        &Status,
	}

	manager.Offer = row

	return res, nil
}

func (manager *Manager) QueryMember(RecordID int) (*model.Member, error) {
	db := db.GetDB()

	var err error

	var member models.Member

	if err = db.Get(&member, "SELECT * FROM Member WHERE MemberID=$1", RecordID); err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("NO_MEMBER_RECORD")
		}
		return nil, err
	}

	MemberID := cast.ToInt(member.MemberID)
	Email := member.Email
	IP := member.IP
	FirstName := member.FirstName
	LastName := member.LastName
	FamilyStatus := member.FamilyStatus
	MaidenName := member.MaidenName
	Citizenship := member.Citizenship
	Country := member.Country
	City := member.City
	Zip := member.Zip
	Address1 := member.Address1
	Address2 := member.Address2
	StreetNumber := member.StreetNumber
	StreetName := member.StreetName
	Image := viper.GetString("s3_cdn_url") + member.Image
	Birthday := member.Birthday.Time.String()
	EmailNotifications := member.EmailNotifications
	Phone := member.Phone
	Created := cast.ToInt(member.Created)
	Role := cast.ToInt(member.Role)
	Gender := member.Gender
	USD := member.USD.Decimal.String()
	EUR := member.EUR.Decimal.String()
	LeverageAllowed := member.LeverageAllowed.Decimal.String()
	StopLossAllowed := member.StopLossAllowed.Decimal.String()
	TakeProfitAllowed := member.TakeProfitAllowed.Decimal.String()
	Status := member.Status
	CurrencyID := cast.ToInt(member.CurrencyID)
	ManagerID := cast.ToInt(member.ManagerID)
	ManagerRole := member.ManagerRole

	res := &model.Member{
		MemberID:           MemberID,
		Email:              &Email,
		IP:                 &IP,
		FirstName:          &FirstName,
		LastName:           &LastName,
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
		Gender:             &Gender,
		Usd:                &USD,
		Eur:                &EUR,
		LeverageAllowed:    &LeverageAllowed,
		StopLossAllowed:    &StopLossAllowed,
		TakeProfitAllowed:  &TakeProfitAllowed,
		Status:             &Status,
		CurrencyID:         &CurrencyID,
		ManagerID:          &ManagerID,
		ManagerRole:        &ManagerRole,
	}

	return res, nil
}

func (manager *Manager) QueryManager(RecordID int) (*model.Member, error) {
	db := db.GetDB()

	var err error

	var member models.Member

	if err = db.Get(&member, "SELECT * FROM Member WHERE (ManagerRole=$1 OR ManagerRole=$2) AND MemberID=$3", "manager", "admin", RecordID); err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("NO_MANAGER_RECORD")
		}
		return nil, err
	}

	MemberID := cast.ToInt(member.MemberID)
	Email := member.Email
	IP := member.IP
	FirstName := member.FirstName
	LastName := member.LastName
	FamilyStatus := member.FamilyStatus
	MaidenName := member.MaidenName
	Citizenship := member.Citizenship
	Country := member.Country
	City := member.City
	Zip := member.Zip
	Address1 := member.Address1
	Address2 := member.Address2
	StreetNumber := member.StreetNumber
	StreetName := member.StreetName
	Image := viper.GetString("s3_cdn_url") + member.Image
	Birthday := member.Birthday.Time.String()
	EmailNotifications := member.EmailNotifications
	Phone := member.Phone
	Created := cast.ToInt(member.Created)
	Role := cast.ToInt(member.Role)
	Gender := member.Gender
	USD := member.USD.Decimal.String()
	EUR := member.EUR.Decimal.String()
	LeverageAllowed := member.LeverageAllowed.Decimal.String()
	StopLossAllowed := member.StopLossAllowed.Decimal.String()
	TakeProfitAllowed := member.TakeProfitAllowed.Decimal.String()
	Status := member.Status
	ManagerRole := member.ManagerRole

	res := &model.Member{
		MemberID:           MemberID,
		Email:              &Email,
		IP:                 &IP,
		FirstName:          &FirstName,
		LastName:           &LastName,
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
		Gender:             &Gender,
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

func (manager *Manager) QueryCategoryList() ([]*model.Category, error) {
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

func (manager *Manager) QueryCurrencyList() ([]*model.Currency, error) {
	db := db.GetDB()

	var res []*model.Currency
	var currency []models.Currency

	if err := db.Select(&currency, "SELECT * FROM Currency ORDER BY CurrencyID DESC"); err != nil {
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

func (manager *Manager) QueryLeadList() ([]*model.Lead, error) {
	db := db.GetDB()

	var res []*model.Lead
	var lead []models.Lead

	if err := db.Select(&lead, "SELECT * FROM Lead ORDER BY LeadID DESC"); err != nil {
		if err != sql.ErrNoRows {
			return nil, err
		}
	}

	for _, row := range lead {

		LeadID := cast.ToInt(row.LeadID.Int)
		ManagerID := cast.ToInt(row.ManagerID.Int)
		MemberID := cast.ToInt(row.MemberID.Int)
		CampaignID := cast.ToInt(row.CampaignID.Int)
		CurrencyID := cast.ToInt(row.CurrencyID.Int)
		Email := row.Email.String
		Phone := row.Phone.String
		IP := row.IP.String
		FirstName := row.FirstName.String
		LastName := row.LastName.String
		Gender := row.Gender.String
		FamilyStatus := row.FamilyStatus.String
		MaidenName := row.MaidenName.String
		Citizenship := row.Citizenship.String
		Country := row.Country.String
		City := row.City.String
		Zip := row.Zip.String
		Address1 := row.Address1.String
		Address2 := row.Address2.String
		StreetNumber := row.StreetNumber.String
		StreetName := row.StreetName.String
		Birthday := row.Birthday.Time.String()
		Status := row.Status.String
		DateCreated := row.DateCreated.Time.String()
		TimestampCreated := cast.ToInt(row.TimestampCreated.Int)

		resRow := &model.Lead{
			LeadID:           LeadID,
			ManagerID:        &ManagerID,
			MemberID:         &MemberID,
			CampaignID:       &CampaignID,
			CurrencyID:       &CurrencyID,
			Email:            &Email,
			Phone:            &Phone,
			IP:               &IP,
			FirstName:        &FirstName,
			LastName:         &LastName,
			Gender:           &Gender,
			FamilyStatus:     &FamilyStatus,
			MaidenName:       &MaidenName,
			Citizenship:      &Citizenship,
			Country:          &Country,
			City:             &City,
			Zip:              &Zip,
			Address1:         &Address1,
			Address2:         &Address2,
			StreetNumber:     &StreetNumber,
			StreetName:       &StreetName,
			Birthday:         &Birthday,
			Status:           &Status,
			DateCreated:      &DateCreated,
			TimestampCreated: &TimestampCreated,
		}

		res = append(res, resRow)
	}

	return res, nil
}

func (manager *Manager) QueryCommentList() ([]*model.Comment, error) {
	db := db.GetDB()

	var res []*model.Comment
	var comment []models.Comment

	if err := db.Select(&comment, "SELECT * FROM Comment ORDER BY CommentID DESC"); err != nil {
		if err != sql.ErrNoRows {
			return nil, err
		}
	}

	for _, row := range comment {

		CommentID := cast.ToInt(row.CommentID.Int)
		MemberID := cast.ToInt(row.MemberID.Int)
		Content := row.Content.String
		DateCreated := row.DateCreated.Time.String()
		TimestampCreated := cast.ToInt(row.TimestampCreated.Int)
		DateEdited := row.DateEdited.Time.String()
		TimestampEdited := cast.ToInt(row.TimestampEdited.Int)
		LeadID := cast.ToInt(row.LeadID.Int)

		resRow := &model.Comment{
			CommentID:        CommentID,
			MemberID:         &MemberID,
			Content:          &Content,
			DateCreated:      &DateCreated,
			TimestampCreated: &TimestampCreated,
			DateEdited:       &DateEdited,
			TimestampEdited:  &TimestampEdited,
			LeadID:           &LeadID,
		}

		res = append(res, resRow)
	}

	return res, nil
}

func (manager *Manager) QueryChecklistList() ([]*model.Checklist, error) {
	db := db.GetDB()

	var res []*model.Checklist
	var checklist []models.Checklist

	if err := db.Select(&checklist, "SELECT * FROM Checklist ORDER BY ChecklistID DESC"); err != nil {
		if err != sql.ErrNoRows {
			return nil, err
		}
	}

	for _, row := range checklist {

		ChecklistID := cast.ToInt(row.ChecklistID.Int)
		Title := row.Title.String
		Complete := row.Complete.Bool
		Position := cast.ToInt(row.Position.Int)
		DateCreated := row.DateCreated.Time.String()
		TimestampCreated := cast.ToInt(row.TimestampCreated.Int)
		LeadID := cast.ToInt(row.LeadID.Int)

		resRow := &model.Checklist{
			ChecklistID:      ChecklistID,
			Title:            &Title,
			Complete:         &Complete,
			Position:         &Position,
			DateCreated:      &DateCreated,
			TimestampCreated: &TimestampCreated,
			LeadID:           &LeadID,
		}

		res = append(res, resRow)
	}

	return res, nil
}

func (manager *Manager) QueryAppointmentList() ([]*model.Appointment, error) {
	db := db.GetDB()

	var res []*model.Appointment
	var Appointment []models.Appointment

	if err := db.Select(&Appointment, "SELECT * FROM Appointment ORDER BY AppointmentID DESC"); err != nil {
		if err != sql.ErrNoRows {
			return nil, err
		}
	}

	for _, row := range Appointment {

		AppointmentID := cast.ToInt(row.AppointmentID.Int)
		Type := row.Type.String
		Title := row.Title.String
		Description := row.Description.String
		DateCreated := row.DateCreated.Time.String()
		TimestampCreated := cast.ToInt(row.TimestampCreated.Int)
		DateDue := row.DateDue.Time.String()
		TimestampDue := cast.ToInt(row.TimestampDue.Int)
		Status := row.Status.String
		LeadID := cast.ToInt(row.LeadID.Int)

		resRow := &model.Appointment{
			AppointmentID:    AppointmentID,
			Type:             &Type,
			Title:            &Title,
			Description:      &Description,
			DateCreated:      &DateCreated,
			TimestampCreated: &TimestampCreated,
			DateDue:          &DateDue,
			TimestampDue:     &TimestampDue,
			Status:           &Status,
			LeadID:           &LeadID,
		}

		res = append(res, resRow)
	}

	return res, nil
}

func (manager *Manager) QueryCampaignList() ([]*model.Campaign, error) {
	db := db.GetDB()

	var res []*model.Campaign
	var Campaign []models.Campaign

	if err := db.Select(&Campaign, "SELECT * FROM Campaign ORDER BY CampaignID DESC"); err != nil {
		if err != sql.ErrNoRows {
			return nil, err
		}
	}

	for _, row := range Campaign {

		CampaignID := cast.ToInt(row.CampaignID.Int)
		Title := row.Title.String
		Description := row.Description.String
		DateCreated := row.DateCreated.Time.String()
		TimestampCreated := cast.ToInt(row.TimestampCreated.Int)

		resRow := &model.Campaign{
			CampaignID:       CampaignID,
			Title:            &Title,
			Description:      &Description,
			DateCreated:      &DateCreated,
			TimestampCreated: &TimestampCreated,
		}

		res = append(res, resRow)
	}

	return res, nil
}

func (manager *Manager) QueryMemberList() ([]*model.Member, error) {
	db := db.GetDB()

	var res []*model.Member
	var member []models.Member

	if err := db.Select(&member, "SELECT * FROM Member ORDER BY MemberID DESC"); err != nil {
		if err != sql.ErrNoRows {
			return nil, err
		}
	}

	for _, row := range member {

		MemberID := cast.ToInt(row.MemberID)
		Email := row.Email
		IP := row.IP
		FirstName := row.FirstName
		LastName := row.LastName
		FamilyStatus := row.FamilyStatus
		MaidenName := row.MaidenName
		Citizenship := row.Citizenship
		Country := row.Country
		City := row.City
		Zip := row.Zip
		Address1 := row.Address1
		Address2 := row.Address2
		StreetNumber := row.StreetNumber
		StreetName := row.StreetName
		Image := viper.GetString("s3_cdn_url") + row.Image
		Birthday := row.Birthday.Time.String()
		EmailNotifications := row.EmailNotifications
		Phone := row.Phone
		Created := cast.ToInt(row.Created)
		Role := cast.ToInt(row.Role)
		Gender := row.Gender
		USD := row.USD.Decimal.String()
		EUR := row.EUR.Decimal.String()
		LeverageAllowed := row.LeverageAllowed.Decimal.String()
		StopLossAllowed := row.StopLossAllowed.Decimal.String()
		TakeProfitAllowed := row.TakeProfitAllowed.Decimal.String()
		Status := row.Status
		ManagerRole := row.ManagerRole

		resRow := &model.Member{
			MemberID:           MemberID,
			Email:              &Email,
			IP:                 &IP,
			FirstName:          &FirstName,
			LastName:           &LastName,
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
			Gender:             &Gender,
			Usd:                &USD,
			Eur:                &EUR,
			LeverageAllowed:    &LeverageAllowed,
			StopLossAllowed:    &StopLossAllowed,
			TakeProfitAllowed:  &TakeProfitAllowed,
			Status:             &Status,
			ManagerRole:        &ManagerRole,
		}

		res = append(res, resRow)
	}

	return res, nil
}

func (manager *Manager) QueryMemberListNoManager() ([]*model.Member, error) {
	db := db.GetDB()

	var res []*model.Member
	var member []models.Member

	if err := db.Select(&member, "SELECT * FROM Member WHERE ManagerRole=$1 ORDER BY MemberID DESC", ""); err != nil {
		if err != sql.ErrNoRows {
			return nil, err
		}
	}

	for _, row := range member {

		MemberID := cast.ToInt(row.MemberID)
		Email := row.Email
		IP := row.IP
		FirstName := row.FirstName
		LastName := row.LastName
		FamilyStatus := row.FamilyStatus
		MaidenName := row.MaidenName
		Citizenship := row.Citizenship
		Country := row.Country
		City := row.City
		Zip := row.Zip
		Address1 := row.Address1
		Address2 := row.Address2
		StreetNumber := row.StreetNumber
		StreetName := row.StreetName
		Image := viper.GetString("s3_cdn_url") + row.Image
		Birthday := row.Birthday.Time.String()
		EmailNotifications := row.EmailNotifications
		Phone := row.Phone
		Created := cast.ToInt(row.Created)
		Role := cast.ToInt(row.Role)
		Gender := row.Gender
		USD := row.USD.Decimal.String()
		EUR := row.EUR.Decimal.String()
		LeverageAllowed := row.LeverageAllowed.Decimal.String()
		StopLossAllowed := row.StopLossAllowed.Decimal.String()
		TakeProfitAllowed := row.TakeProfitAllowed.Decimal.String()
		Status := row.Status
		ManagerRole := row.ManagerRole

		resRow := &model.Member{
			MemberID:           MemberID,
			Email:              &Email,
			IP:                 &IP,
			FirstName:          &FirstName,
			LastName:           &LastName,
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
			Gender:             &Gender,
			Usd:                &USD,
			Eur:                &EUR,
			LeverageAllowed:    &LeverageAllowed,
			StopLossAllowed:    &StopLossAllowed,
			TakeProfitAllowed:  &TakeProfitAllowed,
			Status:             &Status,
			ManagerRole:        &ManagerRole,
		}

		res = append(res, resRow)
	}

	return res, nil
}

func (manager *Manager) QueryManagerList() ([]*model.Member, error) {
	db := db.GetDB()

	var res []*model.Member
	var Member []models.Member

	if err := db.Select(&Member, "SELECT * FROM Member WHERE (ManagerRole=$1 OR ManagerRole=$2) ORDER BY MemberID DESC", "manager", "admin"); err != nil {
		if err != sql.ErrNoRows {
			return nil, err
		}
	}

	for _, row := range Member {

		MemberID := cast.ToInt(row.MemberID)
		Email := row.Email
		IP := row.IP
		FirstName := row.FirstName
		LastName := row.LastName
		FamilyStatus := row.FamilyStatus
		MaidenName := row.MaidenName
		Citizenship := row.Citizenship
		Country := row.Country
		City := row.City
		Zip := row.Zip
		Address1 := row.Address1
		Address2 := row.Address2
		StreetNumber := row.StreetNumber
		StreetName := row.StreetName
		Image := viper.GetString("s3_cdn_url") + row.Image
		Birthday := row.Birthday.Time.String()
		EmailNotifications := row.EmailNotifications
		Phone := row.Phone
		Created := cast.ToInt(row.Created)
		Role := cast.ToInt(row.Role)
		Gender := row.Gender
		USD := row.USD.Decimal.String()
		EUR := row.EUR.Decimal.String()
		LeverageAllowed := row.LeverageAllowed.Decimal.String()
		StopLossAllowed := row.StopLossAllowed.Decimal.String()
		TakeProfitAllowed := row.TakeProfitAllowed.Decimal.String()
		Status := row.Status
		ManagerRole := row.ManagerRole

		resRow := &model.Member{
			MemberID:           MemberID,
			Email:              &Email,
			IP:                 &IP,
			FirstName:          &FirstName,
			LastName:           &LastName,
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
			Gender:             &Gender,
			Usd:                &USD,
			Eur:                &EUR,
			LeverageAllowed:    &LeverageAllowed,
			StopLossAllowed:    &StopLossAllowed,
			TakeProfitAllowed:  &TakeProfitAllowed,
			Status:             &Status,
			ManagerRole:        &ManagerRole,
		}

		res = append(res, resRow)
	}

	return res, nil
}

func (manager *Manager) QueryOfferList() ([]*model.ManagerOffer, error) {
	db := db.GetDB()

	var err error

	var Offer []models.Offer

	if err = db.Select(&Offer, "SELECT * FROM Offer ORDER BY OfferID DESC"); err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("NO_OFFER_RECORD")
		}
		return nil, err
	}

	var res []*model.ManagerOffer

	for _, row := range Offer {

		OfferID := cast.ToInt(row.OfferID)
		MemberID := cast.ToInt(row.MemberID)
		InvestID := cast.ToInt(row.InvestID)
		CurrencyID := cast.ToInt(row.CurrencyID.Int)
		BankDetailsID := cast.ToInt(row.BankDetailsID)
		Title := row.Title
		Status := row.Status

		resRow := &model.ManagerOffer{
			OfferID:       OfferID,
			MemberID:      &MemberID,
			InvestID:      &InvestID,
			CurrencyID:    &CurrencyID,
			BankDetailsID: &BankDetailsID,
			Title:         &Title,
			Status:        &Status,
		}

		res = append(res, resRow)
	}

	return res, nil
}

func (manager *Manager) CreateBankDetails(input model.ManagerCreateBankDetailsRequest) (*model.CreationResponse, error) {
	db := db.GetDB()

	tx := db.MustBegin()
	tx.MustExec("INSERT INTO BankDetails (Title, BeneficiaryCompany, BeneficiaryFirstName, BeneficiaryLastName, BeneficiaryCountry, BeneficiaryCity, BeneficiaryZip, BeneficiaryAddress, BankName, BankBranch, BankIFSC, BankBranchCountry, BankBranchCity, BankBranchZip, BankBranchAddress, BankAccountNumber, BankAccountType, BankRoutingNumber, BankTransferCaption, BankIBAN, BankSWIFT, BankSWIFTCorrespondent, BankBIC) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23)", input.Title, input.BeneficiaryCompany, input.BeneficiaryFirstName, input.BeneficiaryLastName, input.BeneficiaryCountry, input.BeneficiaryCity, input.BeneficiaryZip, input.BeneficiaryAddress, input.BankName, input.BankBranch, input.BankIfsc, input.BankBranchCountry, input.BankBranchCity, input.BankBranchZip, input.BankBranchAddress, input.BankAccountNumber, input.BankAccountType, input.BankRoutingNumber, input.BankTransferCaption, input.BankIban, input.BankSwift, input.BankSWIFTCorrespondent, input.BankBic)
	tx.Commit()

	var RecordID int

	if err := db.Get(&RecordID, "SELECT BankDetailsID FROM BankDetails WHERE Title=$1 AND BeneficiaryCompany=$2 AND BeneficiaryFirstName=$3 AND BeneficiaryLastName=$4 AND BeneficiaryCountry=$5 AND BeneficiaryCity=$6 AND BeneficiaryZip=$7 AND BeneficiaryAddress=$8 AND BankName=$9 AND BankBranch=$10 AND BankIFSC=$11 AND BankBranchCountry=$12 AND BankBranchCity=$13 AND BankBranchZip=$14 AND BankBranchAddress=$15 AND BankAccountNumber=$16 AND BankAccountType=$17 AND BankRoutingNumber=$18 AND BankTransferCaption=$19 AND BankIBAN=$20 AND BankSWIFT=$21 AND BankSWIFTCorrespondent=$22 AND BankBIC=$23 ORDER BY BankDetailsID DESC", input.Title, input.BeneficiaryCompany, input.BeneficiaryFirstName, input.BeneficiaryLastName, input.BeneficiaryCountry, input.BeneficiaryCity, input.BeneficiaryZip, input.BeneficiaryAddress, input.BankName, input.BankBranch, input.BankIfsc, input.BankBranchCountry, input.BankBranchCity, input.BankBranchZip, input.BankBranchAddress, input.BankAccountNumber, input.BankAccountType, input.BankRoutingNumber, input.BankTransferCaption, input.BankIban, input.BankSwift, input.BankSWIFTCorrespondent, input.BankBic); err != nil {
		return nil, err
	}

	return &model.CreationResponse{
		RecordID: RecordID,
	}, nil
}

func (manager *Manager) CreateInvest(input model.ManagerCreateInvestRequest) (*model.CreationResponse, error) {
	db := db.GetDB()

	tx := db.MustBegin()
	tx.MustExec("INSERT INTO Invest (CategoryID, Title, Subtitle, Description) VALUES ($1, $2, $3, $4)", input.CategoryID, input.Title, input.Subtitle, input.Description)
	tx.Commit()

	var RecordID int

	if err := db.Get(&RecordID, "SELECT InvestID FROM Invest WHERE CategoryID=$1 AND Title=$2 AND Subtitle=$3 AND Description=$4 ORDER BY InvestID DESC", input.CategoryID, input.Title, input.Subtitle, input.Description); err != nil {
		return nil, err
	}

	return &model.CreationResponse{
		RecordID: RecordID,
	}, nil
}

func (manager *Manager) CreateContract(input model.ManagerCreateContractRequest) (*model.CreationResponse, error) {
	db := db.GetDB()

	tx := db.MustBegin()
	tx.MustExec("INSERT INTO Contract (Title, ContentRaw, Template) VALUES ($1, $2, $3)", input.Title, input.ContentRaw, true)
	tx.Commit()

	var RecordID int

	if err := db.Get(&RecordID, "SELECT ContractID FROM Contract WHERE Title=$1 AND ContentRaw=$2 AND Template=$3 ORDER BY ContractID DESC", input.Title, input.ContentRaw, true); err != nil {
		return nil, err
	}

	return &model.CreationResponse{
		RecordID: RecordID,
	}, nil
}

func (manager *Manager) CreateLead(input model.ManagerCreateLeadRequest) (*model.CreationResponse, error) {
	db := db.GetDB()

	Birthday := input.Birthday

	if len(*Birthday) == 0 {
		Birthday = nil
	}

	tx := db.MustBegin()
	tx.MustExec("INSERT INTO Lead (MemberID, CampaignID, CurrencyID, Email, Phone, FirstName, LastName, Gender, FamilyStatus, MaidenName, Citizenship, Country, City, Zip, Address1, Address2, StreetNumber, StreetName, Birthday, Status, TimestampCreated) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21)", input.MemberID, input.CampaignID, input.CurrencyID, input.Email, input.Phone, input.FirstName, input.LastName, input.Gender, input.FamilyStatus, input.MaidenName, input.Citizenship, input.Country, input.City, input.Zip, input.Address1, input.Address2, input.StreetNumber, input.StreetName, Birthday, input.Status, manager.Timestamp)
	tx.Commit()

	var RecordID int

	if err := db.Get(&RecordID, "SELECT LeadID FROM Lead WHERE Email=$1 AND Phone=$2 AND FirstName=$3 AND LastName=$4 AND TimestampCreated=$5 ORDER BY LeadID DESC", input.Email, input.Phone, input.FirstName, input.LastName, manager.Timestamp); err != nil {
		return nil, err
	}

	return &model.CreationResponse{
		RecordID: RecordID,
	}, nil
}

func (manager *Manager) CreateComment(input model.ManagerCreateCommentRequest) (*model.CreationResponse, error) {
	db := db.GetDB()

	tx := db.MustBegin()
	tx.MustExec("INSERT INTO Comment (MemberID, Content, TimestampCreated, LeadID) VALUES ($1, $2, $3, $4)", manager.Member.MemberID, input.Content, manager.Timestamp, input.LeadID)
	tx.Commit()

	var RecordID int

	if err := db.Get(&RecordID, "SELECT CommentID FROM Comment WHERE MemberID=$1 AND Content=$2 AND TimestampCreated=$3 AND LeadID=$4 ORDER BY CommentID DESC", manager.Member.MemberID, input.Content, manager.Timestamp, input.LeadID); err != nil {
		return nil, err
	}

	return &model.CreationResponse{
		RecordID: RecordID,
	}, nil
}

func (manager *Manager) CreateChecklist(input model.ManagerCreateChecklistRequest) (*model.CreationResponse, error) {
	db := db.GetDB()

	tx := db.MustBegin()
	tx.MustExec("INSERT INTO Checklist (Title, Complete, Position, TimestampCreated, LeadID) VALUES ($1, $2, $3, $4, $5)", input.Title, input.Complete, input.Position, manager.Timestamp, input.LeadID)
	tx.Commit()

	var RecordID int

	if err := db.Get(&RecordID, "SELECT ChecklistID FROM Checklist WHERE Title=$1 AND Complete=$2 AND Position=$3 AND TimestampCreated=$4 AND LeadID=$5 ORDER BY ChecklistID DESC", input.Title, input.Complete, input.Position, manager.Timestamp, input.LeadID); err != nil {
		return nil, err
	}

	return &model.CreationResponse{
		RecordID: RecordID,
	}, nil
}

func (manager *Manager) CreateAppointment(input model.ManagerCreateAppointmentRequest) (*model.CreationResponse, error) {
	db := db.GetDB()

	TimestampDue := 0

	tx := db.MustBegin()
	tx.MustExec("INSERT INTO Appointment (Type, Title, Description, TimestampCreated, DateDue, TimestampDue, Status, LeadID) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)", input.Type, input.Title, input.Description, manager.Timestamp, input.DateDue, TimestampDue, input.Status, input.LeadID)
	tx.Commit()

	var RecordID int

	if err := db.Get(&RecordID, "SELECT AppointmentID FROM Appointment WHERE TimestampCreated=$1 AND LeadID=$2 ORDER BY AppointmentID DESC", manager.Timestamp, input.LeadID); err != nil {
		return nil, err
	}
	//Type=$1 AND Title=$2 AND Description=$3 AND TimestampCreated=$4 AND LeadID=$5
	//input.Type, input.Title, input.Description, manager.Timestamp, input.LeadID

	return &model.CreationResponse{
		RecordID: RecordID,
	}, nil
}

func (manager *Manager) CreateCampaign(input model.ManagerCreateCampaignRequest) (*model.CreationResponse, error) {
	db := db.GetDB()

	tx := db.MustBegin()
	tx.MustExec("INSERT INTO Campaign (Title, Description, TimestampCreated) VALUES ($1, $2, $3)", input.Title, input.Description, manager.Timestamp)
	tx.Commit()

	var RecordID int

	if err := db.Get(&RecordID, "SELECT CampaignID FROM Campaign WHERE Title=$1 AND Description=$2 AND TimestampCreated=$3 ORDER BY CampaignID DESC", input.Title, input.Description, manager.Timestamp); err != nil {
		return nil, err
	}

	return &model.CreationResponse{
		RecordID: RecordID,
	}, nil
}

func (manager *Manager) CreateOffer(input model.ManagerCreateOfferRequest) (*model.CreationResponse, error) {
	db := db.GetDB()

	tx := db.MustBegin()
	tx.MustExec("INSERT INTO Offer (InvestID, MemberID, CurrencyID, BankDetailsID, Title, Status) VALUES ($1, $2, $3, $4, $5, $6)", input.InvestID, input.MemberID, input.CurrencyID, input.BankDetailsID, input.Title, constants.STATUS_DRAFT)
	tx.Commit()

	var RecordID int

	if err := db.Get(&RecordID, "SELECT OfferID FROM Offer WHERE InvestID=$1 AND MemberID=$2 AND CurrencyID=$3 AND BankDetailsID=$4 AND Title=$5 AND Status=$6 ORDER BY OfferID DESC", input.InvestID, input.MemberID, input.CurrencyID, input.BankDetailsID, input.Title, constants.STATUS_DRAFT); err != nil {
		return nil, err
	}

	return &model.CreationResponse{
		RecordID: RecordID,
	}, nil
}

func (manager *Manager) CreateInterest(input model.ManagerCreateInterestRequest) (*model.CreationResponse, error) {
	db := db.GetDB()

	tx := db.MustBegin()
	tx.MustExec("INSERT INTO Interest (OfferID, AmountFrom, AmountTo, DurationFrom, DurationTo, Interest) VALUES ($1, $2, $3, $4, $5, $6)", input.OfferID, input.AmountFrom, input.AmountTo, input.DurationFrom, input.DurationTo, input.Interest)
	tx.Commit()

	var RecordID int

	if err := db.Get(&RecordID, "SELECT InterestID FROM Interest WHERE OfferID=$1 AND AmountFrom=$2 AND AmountTo=$3 AND DurationFrom=$4 AND DurationTo=$5 AND Interest=$6 ORDER BY InterestID DESC", input.OfferID, input.AmountFrom, input.AmountTo, input.DurationFrom, input.DurationTo, input.Interest); err != nil {
		return nil, err
	}

	return &model.CreationResponse{
		RecordID: RecordID,
	}, nil
}

//Duplicate the contract template and assign duplicated contract to offer
func (manager *Manager) DuplicateAndAssignContractToOffer(input model.ManagerDuplicateAndAssignContractToOfferRequest) (*model.Result, error) {
	db := db.GetDB()

	tx := db.MustBegin()
	tx.MustExec("UPDATE Contract SET Current=$1 WHERE OfferID=$2", false, input.OfferID)
	tx.MustExec("INSERT INTO Contract (ContentRaw, OfferID, Current, Title, Template) VALUES ($1, $2, $3, $4, $5)", manager.Contract.ContentRaw, input.OfferID, true, manager.Contract.Title, false)
	tx.Commit()

	return &model.Result{
		Status: true,
	}, nil
}

func (manager *Manager) AssignInvestToOffer(input model.ManagerAssignInvestToOfferRequest) (*model.Result, error) {
	db := db.GetDB()

	tx := db.MustBegin()
	tx.MustExec("UPDATE Offer SET InvestID=$1 WHERE OfferID=$2", input.InvestID, input.OfferID)
	tx.Commit()

	return &model.Result{
		Status: true,
	}, nil
}

func (manager *Manager) AssignBankDetailsToOffer(input model.ManagerAssignBankDetailsToOfferRequest) (*model.Result, error) {
	db := db.GetDB()

	tx := db.MustBegin()
	tx.MustExec("UPDATE Offer SET BankDetailsID=$1 WHERE OfferID=$2", input.BankDetailsID, input.OfferID)
	tx.Commit()

	return &model.Result{
		Status: true,
	}, nil
}

func (manager *Manager) AssignLeadToManager(input model.ManagerAssignLeadToManagerRequest) (*model.Result, error) {
	db := db.GetDB()

	tx := db.MustBegin()
	tx.MustExec("UPDATE Lead SET ManagerID=$1 WHERE LeadID=$2", input.ManagerID, input.LeadID)
	tx.Commit()

	return &model.Result{
		Status: true,
	}, nil
}

func (manager *Manager) DeactivateOffer(input model.RecordRequest) (*model.Result, error) {
	db := db.GetDB()

	tx := db.MustBegin()
	tx.MustExec("UPDATE Offer SET Status=$1 WHERE OfferID=$2", constants.STATUS_NOACTIVE, input.RecordID)
	tx.Commit()

	return &model.Result{
		Status: true,
	}, nil
}

//TODO:
func (manager *Manager) UpdateOfferStatus() error {
	var err error
	var Status string

	interestCount, err := manager.InterestCountByOfferID(manager.Offer.OfferID)
	if err != nil {
		return err
	}

	contractCount, err := manager.ContractCountByOfferID(manager.Offer.OfferID) //current
	if err != nil {
		return err
	}

	if interestCount > 0 && contractCount > 0 {
		Status = constants.STATUS_ACTIVE
	} else {
		Status = constants.STATUS_DRAFT
	}

	db := db.GetDB()
	tx := db.MustBegin()
	tx.MustExec("UPDATE Offer SET Status=$1 WHERE OfferID=$2", Status, manager.Offer.OfferID)
	tx.Commit()

	return nil
}

func (manager *Manager) InterestCountByOfferID(OfferID int64) (int, error) {
	db := db.GetDB()

	var count int

	if err := db.Get(&count, "SELECT count(*) FROM Interest WHERE OfferID=$1", OfferID); err != nil {
		if err != sql.ErrNoRows {
			return 0, err
		}
	}

	return count, nil
}

func (manager *Manager) ContractCountByOfferID(OfferID int64) (int, error) {
	db := db.GetDB()

	var count int

	if err := db.Get(&count, "SELECT count(*) FROM Contract WHERE Current=$1 AND OfferID=$2", true, OfferID); err != nil {
		if err != sql.ErrNoRows {
			return 0, err
		}
	}

	return count, nil
}

func (manager *Manager) CancelOffer(input model.RecordRequest) (*model.Result, error) {
	db := db.GetDB()

	tx := db.MustBegin()
	tx.MustExec("UPDATE Offer SET Status=$1 WHERE OfferID=$2", constants.STATUS_CANCELLED, input.RecordID)
	tx.Commit()

	return &model.Result{
		Status: true,
	}, nil
}
