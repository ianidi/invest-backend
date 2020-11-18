package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/99designs/gqlgen/handler"
	"github.com/google/uuid"
	"github.com/ianidi/exchange-server/graph/generated"
	"github.com/ianidi/exchange-server/graph/methods/constants"
	"github.com/ianidi/exchange-server/graph/methods/manager"
	"github.com/ianidi/exchange-server/graph/methods/portal"
	"github.com/ianidi/exchange-server/graph/model"
	"github.com/ianidi/exchange-server/internal/db"
	"github.com/ianidi/exchange-server/internal/mail"
	"github.com/ianidi/exchange-server/internal/redis"
	"github.com/ianidi/exchange-server/internal/s3"
	"github.com/ianidi/exchange-server/internal/utils"
	"github.com/spf13/cast"
)

func (r *mutationResolver) MemberPersonalUpdate(ctx context.Context, input model.MemberPersonalUpdateRequest) (*model.Result, error) {
	var err error

	portal := portal.Portal{
		Ctx:       ctx,
		Timestamp: time.Now().Unix(),
		Timeout:   60,
	}

	err = portal.GetMember()
	if err != nil {
		return nil, err
	}

	err = portal.ValidateFirstName(input.FirstName)
	if err != nil {
		return nil, err
	}

	err = portal.ValidateLastName(input.LastName)
	if err != nil {
		return nil, err
	}

	err = portal.MemberPersonalUpdate(input)
	if err != nil {
		return nil, err
	}

	return &model.Result{
		Status: true,
	}, nil
}

func (r *mutationResolver) MemberPhoneUpdate(ctx context.Context, input model.MemberPhoneUpdateRequest) (*model.Result, error) {
	var err error

	portal := portal.Portal{
		Ctx:       ctx,
		Timestamp: time.Now().Unix(),
		Timeout:   60,
	}

	err = portal.GetMember()
	if err != nil {
		return nil, err
	}

	err = portal.ParsePhone(input.Phone)
	if err != nil {
		return nil, err
	}

	//TODO:
	// err = portal.CheckPhoneInUse()
	// if err != nil {
	// 	return nil, err
	// }

	err = portal.MemberPhoneUpdate()
	if err != nil {
		return nil, err
	}

	return &model.Result{
		Status: true,
	}, nil
}

func (r *mutationResolver) MemberEmailUpdate(ctx context.Context, input model.MemberEmailUpdateRequest) (*model.Result, error) {
	var err error

	portal := portal.Portal{
		Ctx:       ctx,
		Timestamp: time.Now().Unix(),
		Timeout:   60,
	}

	err = portal.GetMember()
	if err != nil {
		return nil, err
	}

	err = portal.ParseEmail(input.Email)
	if err != nil {
		return nil, err
	}

	//TODO:
	// err = portal.CheckEmailInUse()
	// if err != nil {
	// 	return nil, err
	// }

	err = portal.MemberEmailUpdate()
	if err != nil {
		return nil, err
	}

	return &model.Result{
		Status: true,
	}, nil
}

func (r *mutationResolver) ManagerCreateBankDetails(ctx context.Context, input model.ManagerCreateBankDetailsRequest) (*model.CreationResponse, error) {
	var err error

	manager := manager.Manager{
		Ctx:       ctx,
		Timestamp: time.Now().Unix(),
	}

	res, err := manager.CreateBankDetails(input)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (r *mutationResolver) ManagerCreateInvest(ctx context.Context, input model.ManagerCreateInvestRequest) (*model.CreationResponse, error) {
	var err error

	manager := manager.Manager{
		Ctx:       ctx,
		Timestamp: time.Now().Unix(),
	}

	res, err := manager.CreateInvest(input)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (r *mutationResolver) ManagerCreateOffer(ctx context.Context, input model.ManagerCreateOfferRequest) (*model.CreationResponse, error) {
	var err error

	manager := manager.Manager{
		Ctx:       ctx,
		Timestamp: time.Now().Unix(),
	}

	res, err := manager.CreateOffer(input)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (r *mutationResolver) ManagerCreateInterest(ctx context.Context, input model.ManagerCreateInterestRequest) (*model.CreationResponse, error) {
	var err error

	manager := manager.Manager{
		Ctx:       ctx,
		Timestamp: time.Now().Unix(),
	}

	_, err = manager.QueryOffer(input.OfferID)
	if err != nil {
		return nil, err
	}

	res, err := manager.CreateInterest(input)
	if err != nil {
		return nil, err
	}

	err = manager.UpdateOfferStatus()
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (r *mutationResolver) ManagerCreateCategory(ctx context.Context, input model.ManagerCreateCategoryRequest) (*model.CreationResponse, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) ManagerCreateContract(ctx context.Context, input model.ManagerCreateContractRequest) (*model.CreationResponse, error) {
	var err error

	manager := manager.Manager{
		Ctx:       ctx,
		Timestamp: time.Now().Unix(),
	}

	res, err := manager.CreateContract(input)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (r *mutationResolver) ManagerCreateCurrency(ctx context.Context, input model.ManagerCreateCurrencyRequest) (*model.CreationResponse, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) ManagerCreateLead(ctx context.Context, input model.ManagerCreateLeadRequest) (*model.CreationResponse, error) {
	var err error

	manager := manager.Manager{
		Ctx:       ctx,
		Timestamp: time.Now().Unix(),
	}

	res, err := manager.CreateLead(input)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (r *mutationResolver) ManagerCreateComment(ctx context.Context, input model.ManagerCreateCommentRequest) (*model.CreationResponse, error) {
	var err error

	manager := manager.Manager{
		Ctx:       ctx,
		Timestamp: time.Now().Unix(),
	}

	err = manager.GetMember()
	if err != nil {
		return nil, err
	}

	res, err := manager.CreateComment(input)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (r *mutationResolver) ManagerCreateChecklist(ctx context.Context, input model.ManagerCreateChecklistRequest) (*model.CreationResponse, error) {
	var err error

	manager := manager.Manager{
		Ctx:       ctx,
		Timestamp: time.Now().Unix(),
	}

	res, err := manager.CreateChecklist(input)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (r *mutationResolver) ManagerCreateAppointment(ctx context.Context, input model.ManagerCreateAppointmentRequest) (*model.CreationResponse, error) {
	var err error

	manager := manager.Manager{
		Ctx:       ctx,
		Timestamp: time.Now().Unix(),
	}

	res, err := manager.CreateAppointment(input)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (r *mutationResolver) ManagerCreateCampaign(ctx context.Context, input model.ManagerCreateCampaignRequest) (*model.CreationResponse, error) {
	var err error

	manager := manager.Manager{
		Ctx:       ctx,
		Timestamp: time.Now().Unix(),
	}

	res, err := manager.CreateCampaign(input)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (r *mutationResolver) ManagerAssignMemberToOffer(ctx context.Context, input model.ManagerAssignMemberToOfferRequest) (*model.Result, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) ManagerAssignInvestToOffer(ctx context.Context, input model.ManagerAssignInvestToOfferRequest) (*model.Result, error) {
	var err error

	manager := manager.Manager{
		Ctx:       ctx,
		Timestamp: time.Now().Unix(),
	}

	//TODO: check that offer status is draft

	res, err := manager.AssignInvestToOffer(input)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (r *mutationResolver) ManagerDuplicateAndAssignContractToOffer(ctx context.Context, input model.ManagerDuplicateAndAssignContractToOfferRequest) (*model.Result, error) {
	var err error

	manager := manager.Manager{
		Ctx:       ctx,
		Timestamp: time.Now().Unix(),
	}

	_, err = manager.QueryOffer(input.OfferID)
	if err != nil {
		return nil, err
	}

	_, err = manager.QueryContract(input.ContractID)
	if err != nil {
		return nil, err
	}

	res, err := manager.DuplicateAndAssignContractToOffer(input)
	if err != nil {
		return nil, err
	}

	err = manager.UpdateOfferStatus()
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (r *mutationResolver) ManagerAssignBankDetailsToOffer(ctx context.Context, input model.ManagerAssignBankDetailsToOfferRequest) (*model.Result, error) {
	var err error

	manager := manager.Manager{
		Ctx:       ctx,
		Timestamp: time.Now().Unix(),
	}

	res, err := manager.AssignBankDetailsToOffer(input)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (r *mutationResolver) ManagerAssignLeadToManager(ctx context.Context, input model.ManagerAssignLeadToManagerRequest) (*model.Result, error) {
	var err error

	manager := manager.Manager{
		Ctx:       ctx,
		Timestamp: time.Now().Unix(),
	}

	res, err := manager.AssignLeadToManager(input)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (r *mutationResolver) ManagerDuplicateInvest(ctx context.Context, input model.ManagerDuplicateRequest) (*model.CreationResponse, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) ManagerDuplicateOffer(ctx context.Context, input model.ManagerDuplicateRequest) (*model.CreationResponse, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) ManagerDuplicateContract(ctx context.Context, input model.ManagerDuplicateRequest) (*model.CreationResponse, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) ManagerDeactivateOffer(ctx context.Context, input model.RecordRequest) (*model.Result, error) {
	var err error

	manager := manager.Manager{
		Ctx:       ctx,
		Timestamp: time.Now().Unix(),
	}

	res, err := manager.DeactivateOffer(input)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (r *mutationResolver) ManagerActivateOffer(ctx context.Context, input model.RecordRequest) (*model.Result, error) {
	var err error

	manager := manager.Manager{
		Ctx:       ctx,
		Timestamp: time.Now().Unix(),
	}

	_, err = manager.QueryOffer(input.RecordID)
	if err != nil {
		return nil, err
	}

	err = manager.UpdateOfferStatus()
	if err != nil {
		return nil, err
	}

	return &model.Result{
		Status: true,
	}, nil
}

func (r *mutationResolver) ManagerCancelOffer(ctx context.Context, input model.RecordRequest) (*model.Result, error) {
	var err error

	manager := manager.Manager{
		Ctx:       ctx,
		Timestamp: time.Now().Unix(),
	}

	res, err := manager.CancelOffer(input)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (r *mutationResolver) ManagerEditBankDetails(ctx context.Context, input model.ManagerEditBankDetailsRequest) (*model.Result, error) {
	var err error

	manager := manager.Manager{
		Ctx:       ctx,
		Timestamp: time.Now().Unix(),
	}

	res, err := manager.EditBankDetails(input)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (r *mutationResolver) ManagerEditInvest(ctx context.Context, input model.ManagerEditInvestRequest) (*model.Result, error) {
	var err error

	manager := manager.Manager{
		Ctx:       ctx,
		Timestamp: time.Now().Unix(),
	}

	res, err := manager.EditInvest(input)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (r *mutationResolver) ManagerEditOffer(ctx context.Context, input model.ManagerEditOfferRequest) (*model.Result, error) {
	var err error

	manager := manager.Manager{
		Ctx:       ctx,
		Timestamp: time.Now().Unix(),
	}

	res, err := manager.EditOffer(input)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (r *mutationResolver) ManagerEditCategory(ctx context.Context, input model.ManagerEditCategoryRequest) (*model.Result, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) ManagerEditContract(ctx context.Context, input model.ManagerEditContractRequest) (*model.Result, error) {
	var err error

	manager := manager.Manager{
		Ctx:       ctx,
		Timestamp: time.Now().Unix(),
	}

	res, err := manager.EditContract(input)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (r *mutationResolver) ManagerEditCurrency(ctx context.Context, input model.ManagerEditCurrencyRequest) (*model.Result, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) ManagerEditInvoice(ctx context.Context, input model.ManagerEditInvoiceRequest) (*model.Result, error) {
	var err error

	manager := manager.Manager{
		Ctx:       ctx,
		Timestamp: time.Now().Unix(),
	}

	err = manager.ValidateInvoiceStatus(input.Status)
	if err != nil {
		return nil, err
	}

	_, err = manager.QueryInvoice(input.InvoiceID)
	if err != nil {
		return nil, err
	}

	res, err := manager.EditInvoice(input)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (r *mutationResolver) ManagerEditLead(ctx context.Context, input model.ManagerEditLeadRequest) (*model.Result, error) {
	var err error

	manager := manager.Manager{
		Ctx:       ctx,
		Timestamp: time.Now().Unix(),
	}

	res, err := manager.EditLead(input)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (r *mutationResolver) ManagerEditComment(ctx context.Context, input model.ManagerEditCommentRequest) (*model.Result, error) {
	var err error

	manager := manager.Manager{
		Ctx:       ctx,
		Timestamp: time.Now().Unix(),
	}

	res, err := manager.EditComment(input)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (r *mutationResolver) ManagerEditChecklist(ctx context.Context, input model.ManagerEditChecklistRequest) (*model.Result, error) {
	var err error

	manager := manager.Manager{
		Ctx:       ctx,
		Timestamp: time.Now().Unix(),
	}

	res, err := manager.EditChecklist(input)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (r *mutationResolver) ManagerEditAppointment(ctx context.Context, input model.ManagerEditAppointmentRequest) (*model.Result, error) {
	var err error

	manager := manager.Manager{
		Ctx:       ctx,
		Timestamp: time.Now().Unix(),
	}

	res, err := manager.EditAppointment(input)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (r *mutationResolver) ManagerEditCampaign(ctx context.Context, input model.ManagerEditCampaignRequest) (*model.Result, error) {
	var err error

	manager := manager.Manager{
		Ctx:       ctx,
		Timestamp: time.Now().Unix(),
	}

	res, err := manager.EditCampaign(input)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (r *mutationResolver) ManagerEditMedia(ctx context.Context, input model.ManagerEditMediaRequest) (*model.Result, error) {
	var err error

	manager := manager.Manager{
		Ctx:       ctx,
		Timestamp: time.Now().Unix(),
	}

	res, err := manager.EditMedia(input)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (r *mutationResolver) ManagerRemoveLead(ctx context.Context, input model.RecordRequest) (*model.Result, error) {
	var err error

	manager := manager.Manager{
		Ctx:       ctx,
		Timestamp: time.Now().Unix(),
	}

	res, err := manager.RemoveLead(input)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (r *mutationResolver) ManagerRemoveComment(ctx context.Context, input model.RecordRequest) (*model.Result, error) {
	var err error

	manager := manager.Manager{
		Ctx:       ctx,
		Timestamp: time.Now().Unix(),
	}

	res, err := manager.RemoveComment(input)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (r *mutationResolver) ManagerRemoveChecklist(ctx context.Context, input model.RecordRequest) (*model.Result, error) {
	var err error

	manager := manager.Manager{
		Ctx:       ctx,
		Timestamp: time.Now().Unix(),
	}

	res, err := manager.RemoveChecklist(input)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (r *mutationResolver) ManagerRemoveAppointment(ctx context.Context, input model.RecordRequest) (*model.Result, error) {
	var err error

	manager := manager.Manager{
		Ctx:       ctx,
		Timestamp: time.Now().Unix(),
	}

	res, err := manager.RemoveAppointment(input)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (r *mutationResolver) ManagerRemoveCampaign(ctx context.Context, input model.RecordRequest) (*model.Result, error) {
	var err error

	manager := manager.Manager{
		Ctx:       ctx,
		Timestamp: time.Now().Unix(),
	}

	res, err := manager.RemoveCampaign(input)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (r *mutationResolver) ManagerRemoveInterest(ctx context.Context, input model.RecordRequest) (*model.Result, error) {
	var err error

	manager := manager.Manager{
		Ctx:       ctx,
		Timestamp: time.Now().Unix(),
	}

	res, err := manager.RemoveInterest(input)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (r *mutationResolver) ManagerRemoveManager(ctx context.Context, input model.RecordRequest) (*model.Result, error) {
	var err error

	manager := manager.Manager{
		Ctx:       ctx,
		Timestamp: time.Now().Unix(),
	}

	res, err := manager.RemoveManager(input)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (r *mutationResolver) ManagerRemoveMedia(ctx context.Context, input model.RecordRequest) (*model.Result, error) {
	var err error

	manager := manager.Manager{
		Ctx:       ctx,
		Timestamp: time.Now().Unix(),
	}

	res, err := manager.RemoveMedia(input)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (r *mutationResolver) ManagerDragMedia(ctx context.Context, input model.DragRequest) (*model.Result, error) {
	var err error

	manager := manager.Manager{
		Ctx:       ctx,
		Timestamp: time.Now().Unix(),
	}

	res, err := manager.DragMedia(input)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (r *mutationResolver) ManagerAssignManager(ctx context.Context, input model.ManagerAssignManagerRequest) (*model.Result, error) {
	var err error

	manager := manager.Manager{
		Ctx:       ctx,
		Timestamp: time.Now().Unix(),
	}

	res, err := manager.AssignManager(input)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (r *mutationResolver) ValidateField(ctx context.Context, input model.ValidateFieldRequest) (*model.ValudationStatus, error) {
	message := ""
	status := true
	var err error

	manager := manager.Manager{
		Ctx:       ctx,
		Timestamp: time.Now().Unix(),
	}

	if *input.Type == "manager" {
		if *input.Section == "invest" {
			if *input.Path == "create" {
				if input.Field == "CategoryID" {
					err = manager.ValidateCategoryID(input.Value)
					if err != nil {
						status = false
						message = err.Error()
					}
				}
				if input.Field == "Title" {
					err = manager.ValidateTitle(input.Value)
					if err != nil {
						status = false
						message = err.Error()
					}
				}
				if input.Field == "Subtitle" {
					err = manager.ValidateSubtitle(input.Value)
					if err != nil {
						status = false
						message = err.Error()
					}
				}
				if input.Field == "Description" {
					err = manager.ValidateDescription(input.Value)
					if err != nil {
						status = false
						message = err.Error()
					}
				}
				if input.Field == "CurrencyID" {
					err = manager.ValidateCurrencyID(input.Value)
					if err != nil {
						status = false
						message = err.Error()
					}
				}
			}
		}
	}

	portal := portal.Portal{
		Ctx: ctx,
	}

	if input.Field == "FirstName" {
		err = portal.ValidateFirstName(input.Value)
		if err != nil {
			status = false
			message = err.Error()
		}
	}

	if input.Field == "LastName" {
		err = portal.ValidateLastName(input.Value)
		if err != nil {
			status = false
			message = err.Error()
		}
	}

	if input.Field == "Birthday" {
		err = portal.ValidateBirthday(input.Value)
		if err != nil {
			status = false
			message = err.Error()
		}
	}

	if input.Field == "Citizenship" {
		err = portal.ValidateCitizenship(input.Value)
		if err != nil {
			status = false
			message = err.Error()
		}
	}

	if input.Field == "MaidenName" {
		err = portal.ValidateMaidenName(input.Value)
		if err != nil {
			status = false
			message = err.Error()
		}
	}

	if input.Field == "Email" {
		err = portal.ParseEmail(input.Value)
		if err != nil {
			status = true
			message = err.Error()
		}

		if err == nil {
			err = portal.CheckEmailInUse()
			if err != nil {
				status = false
				message = err.Error()
			}
		}
	}

	if input.Field == "Phone" {
		err = portal.ParsePhone(input.Value)
		if err != nil {
			status = true
			message = err.Error()
		}

		if err == nil {
			err = portal.CheckPhoneInUse()
			if err != nil {
				status = false
				message = err.Error()
			}
		}
	}

	if input.Field == "Gender" {
		err = portal.ValidateGender(input.Value)
		if err != nil {
			status = false
			message = err.Error()
		}
	}

	if input.Field == "FamilyStatus" {
		err = portal.ValidateFamilyStatus(input.Value)
		if err != nil {
			status = false
			message = err.Error()
		}
	}

	if input.Field == "Country" {
		err = portal.ValidateCountry(input.Value)
		if err != nil {
			status = false
			message = err.Error()
		}
	}

	if input.Field == "City" {
		err = portal.ValidateCity(input.Value)
		if err != nil {
			status = false
			message = err.Error()
		}
	}

	if input.Field == "Zip" {
		err = portal.ValidateZip(input.Value)
		if err != nil {
			status = false
			message = err.Error()
		}
	}
	if input.Field == "StreetNumber" {
		err = portal.ValidateStreetNumber(input.Value)
		if err != nil {
			status = false
			message = err.Error()
		}
	}
	if input.Field == "StreetName" {
		err = portal.ValidateStreetName(input.Value)
		if err != nil {
			status = false
			message = err.Error()
		}
	}

	return &model.ValudationStatus{
		Status:  status,
		Message: message,
	}, nil
}

func (r *mutationResolver) OfferInitDeal(ctx context.Context, input model.OfferInitDealRequest) (*model.Result, error) {
	var err error

	portal := portal.Portal{
		Ctx: ctx,
	}

	err = portal.GetMember()
	if err != nil {
		return nil, err
	}

	err = portal.QueryOffer(input.OfferID)
	if err != nil {
		return nil, err
	}

	err = portal.QueryCurrentContract()
	if err != nil {
		return nil, err
	}

	res, err := portal.CreateDeal(input.Amount, input.Duration)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (r *mutationResolver) OfferGeneratePdf(ctx context.Context, input model.RecordRequest) (*model.OfferPdf, error) {
	var err error

	portal := portal.Portal{
		Ctx: ctx,
	}

	err = portal.GetMember()
	if err != nil {
		return nil, err
	}

	err = portal.QueryOffer(input.RecordID)
	if err != nil {
		return nil, err
	}

	err = portal.QueryCurrentContract()
	if err != nil {
		return nil, err
	}

	err = portal.ParseContractContent()
	if err != nil {
		return nil, err
	}

	pdf, err := utils.GeneratePDF(portal.Contract.Content)
	if err != nil {
		return nil, err
	}

	filename := cast.ToString(portal.Member.MemberID) + "_" + uuid.New().String() + ".pdf"

	url, err := s3.Upload(filename, pdf)
	if err != nil {
		return nil, errors.New("UPLOAD_ERROR")
	}

	//TODO: remove previous version of document

	return &model.OfferPdf{
		URL: url,
	}, nil
}

func (r *mutationResolver) OfferSign(ctx context.Context, input model.OfferSignRequest) (*model.OfferSignResult, error) {
	var err error

	portal := portal.Portal{
		Ctx:       ctx,
		Timestamp: time.Now().Unix(),
		Timeout:   60,
		Action:    constants.ACTION_SIGN_CONTRACT,
		Method:    constants.METHOD_PHONE,
	}

	err = portal.GetMember()
	if err != nil {
		return nil, err
	}

	err = portal.QueryOffer(input.OfferID)
	if err != nil {
		return nil, err
	}

	err = portal.QueryCurrentContract()
	if err != nil {
		return nil, err
	}

	_, err = portal.QueryDealByOfferID(input.OfferID)
	if err != nil {
		return nil, err
	}

	decodedImage, filename, err := portal.ProcessSignatureImage(input.Image)
	if err != nil {
		return nil, err
	}

	_, err = s3.Upload(filename, decodedImage)
	if err != nil {
		return nil, errors.New("UPLOAD_ERROR")
	}

	err = portal.UpdateDealSignature(filename)
	if err != nil {
		return nil, err
	}

	Timeout, err := portal.CreateVerify()

	return &model.OfferSignResult{
		Status:  true,
		Timeout: &Timeout,
	}, nil
}

func (r *mutationResolver) CancelDeal(ctx context.Context, input model.RecordRequest) (*model.Result, error) {
	var err error

	portal := portal.Portal{
		Ctx:       ctx,
		Timestamp: time.Now().Unix(),
	}

	err = portal.GetMember()
	if err != nil {
		return nil, err
	}

	_, err = portal.QueryDeal(input.RecordID)
	if err != nil {
		return nil, err
	}

	err = portal.CancelDeal(input.RecordID)
	if err != nil {
		return nil, err
	}

	err = portal.CancelInvoiceByDealID(input.RecordID)
	if err != nil {
		return nil, err
	}

	return &model.Result{
		Status: true,
	}, nil
}

func (r *mutationResolver) RemoveDeal(ctx context.Context, input model.RecordRequest) (*model.Result, error) {
	var err error

	portal := portal.Portal{
		Ctx:       ctx,
		Timestamp: time.Now().Unix(),
	}

	err = portal.GetMember()
	if err != nil {
		return nil, err
	}

	_, err = portal.QueryDeal(input.RecordID)
	if err != nil {
		return nil, err
	}

	err = portal.RemoveDeal(input.RecordID)
	if err != nil {
		return nil, err
	}

	return &model.Result{
		Status: true,
	}, nil
}

func (r *mutationResolver) InvoiceSendToEmail(ctx context.Context, input model.InvoiceSendToEmailRequest) (*model.Result, error) {
	var err error

	portal := portal.Portal{
		Ctx:       ctx,
		Timestamp: time.Now().Unix(),
		Timeout:   60,
		Action:    constants.ACTION_SIGN_CONTRACT,
		Method:    constants.METHOD_PHONE,
	}

	err = portal.ParseEmail(input.Email)
	if err != nil {
		return nil, err
	}

	err = portal.GetMember()
	if err != nil {
		return nil, err
	}

	_, err = portal.QueryInvoice(input.InvoiceID)
	if err != nil {
		return nil, err
	}

	OfferID := cast.ToInt(portal.Invoice.OfferID)

	err = portal.QueryOffer(OfferID)
	if err != nil {
		return nil, err
	}

	BankDetailsID := cast.ToInt(portal.Offer.BankDetailsID)

	_, err = portal.QueryBankDetails(BankDetailsID)
	if err != nil {
		return nil, err
	}

	err = portal.GenerateInvoiceLink()
	if err != nil {
		return nil, err
	}

	mail.SendMail(portal.Member.Email, "Invoice #"+cast.ToString(portal.Invoice.InvoiceID), "Dear customer,", "Please pay your invoice by clicking on the button below.", "View invoice", portal.InvoiceLink, "Thank you for using our platform.")

	return &model.Result{
		Status: true,
	}, nil
}

func (r *mutationResolver) SignIn(ctx context.Context, input model.SignInRequest) (*model.SignInResponse, error) {
	var err error

	portal := portal.Portal{
		Ctx:            ctx,
		ProfileHandler: r.ProfileHandler,
		Timestamp:      time.Now().Unix(),
		Timeout:        60,
		Action:         constants.ACTION_SIGNIN,
		Method:         constants.METHOD_EMAIL,
	}

	err = portal.ParseEmail(input.Email)
	if err != nil {
		return nil, err
	}

	err = portal.ValidatePassword(input.Password, true)
	if err != nil {
		return nil, err
	}

	err = portal.QueryMemberByEmail()
	if err != nil {
		return nil, err
	}

	err = portal.CheckPassword()
	if err != nil {
		return nil, err
	}

	err = portal.CheckMemberStatusIsActive()
	if err != nil {
		Message := err.Error()

		return &model.SignInResponse{
			Status:  false,
			Message: &Message,
		}, nil
	}

	err = portal.CheckMemberStatusIsDisabled()
	if err != nil {
		return nil, err
	}

	Token, err := portal.GenerateAuthorizationToken(portal.Member.MemberID)
	if err != nil {
		return nil, err
	}

	return &model.SignInResponse{
		Status: true,
		Token:  &Token,
	}, nil
}

func (r *mutationResolver) SignUp(ctx context.Context, input model.SignUpRequest) (*model.CreationResponse, error) {
	var err error

	portal := portal.Portal{
		Ctx:       ctx,
		Timestamp: time.Now().Unix(),
		Timeout:   60,
		Action:    constants.ACTION_SIGNUP,
		Method:    constants.METHOD_EMAIL,
	}

	err = portal.ParseEmail(input.Email)
	if err != nil {
		return nil, err
	}

	err = portal.CheckEmailInUse()
	if err != nil {
		return nil, err
	}

	err = portal.ParsePhone(input.Phone)
	if err != nil {
		return nil, err
	}

	err = portal.CheckPhoneInUse()
	if err != nil {
		return nil, err
	}

	err = portal.ValidatePassword(input.Password, true)
	if err != nil {
		return nil, err
	}

	err = portal.ValidateFirstName(input.FirstName)
	if err != nil {
		return nil, err
	}

	err = portal.ValidateLastName(input.LastName)
	if err != nil {
		return nil, err
	}

	err = portal.ValidateBirthday(input.Birthday)
	if err != nil {
		return nil, err
	}

	err = portal.ValidateCitizenship(input.Citizenship)
	if err != nil {
		return nil, err
	}

	err = portal.ValidateGender(input.Gender)
	if err != nil {
		return nil, err
	}

	err = portal.ValidateFamilyStatus(input.FamilyStatus)
	if err != nil {
		return nil, err
	}

	err = portal.ValidateMaidenName(input.MaidenName)
	if err != nil {
		return nil, err
	}

	err = portal.ValidateCountry(input.Country)
	if err != nil {
		return nil, err
	}

	err = portal.ValidateCity(input.City)
	if err != nil {
		return nil, err
	}

	err = portal.ValidateZip(input.Zip)
	if err != nil {
		return nil, err
	}

	err = portal.ValidateStreetNumber(input.StreetNumber)
	if err != nil {
		return nil, err
	}

	err = portal.ValidateStreetName(input.StreetName)
	if err != nil {
		return nil, err
	}

	portal.PasswordHash, err = portal.HashBcrypt(input.Password)
	if err != nil {
		return nil, err
	}

	// queryleadbyemail
	// manager
	// currencyid

	// portal.Settings.DefaultCurrencyID

	err = portal.CreateMember()
	if err != nil {
		return nil, err
	}

	_, err = portal.CreateVerify()
	if err != nil {
		return nil, err
	}

	err = portal.GenerateConfirmationLink()
	if err != nil {
		return nil, err
	}

	mail.SendMail(portal.Email, "Account activation", "Dear customer,", "Please activate your account by clicking on the button below.", "Activate account", portal.ConfirmationLink, "Thank you for joining our platform.")

	return &model.CreationResponse{
		RecordID: cast.ToInt(portal.Member.MemberID),
	}, nil
}

func (r *mutationResolver) Reset(ctx context.Context, input model.ResetRequest) (*model.Result, error) {
	var err error

	portal := portal.Portal{
		Ctx:       ctx,
		Timestamp: time.Now().Unix(),
		Timeout:   60,
		Action:    constants.ACTION_RESET,
		Method:    constants.METHOD_EMAIL,
	}

	err = portal.ParseEmail(input.Email)
	if err != nil {
		return nil, err
	}

	err = portal.QueryMemberByEmail()
	if err != nil {
		return nil, err
	}

	err = portal.CheckMemberStatusIsActive()
	if err != nil {
		return nil, err
	}

	_, err = portal.CreateVerify()
	if err != nil {
		return nil, err
	}

	err = portal.GenerateConfirmationLink()
	if err != nil {
		return nil, err
	}

	mail.SendMail(portal.Email, "Password reset", "Dear customer,", "Please reset your password by clicking on the button below.", "Reset password", portal.ConfirmationLink, "If you have received a password reset email without requesting one, we recommend that you take steps to secure your account.")

	return &model.Result{
		Status: true,
	}, nil
}

func (r *mutationResolver) ResetComplete(ctx context.Context, input model.ResetCompleteRequest) (*model.VerifyResponse, error) {
	var err error

	portal := portal.Portal{
		Ctx:            ctx,
		ProfileHandler: r.ProfileHandler,
		Timestamp:      time.Now().Unix(),
		Timeout:        60,
	}

	err = portal.ValidateResetAction(input.Action)
	if err != nil {
		return nil, err
	}

	err = portal.ValidateResetMethod(input.Method)
	if err != nil {
		return nil, err
	}

	err = portal.QueryVerifyByHash(input.Hash)
	if err != nil {
		return nil, err
	}

	err = portal.ValidateOTPCode(input.Code)
	if err != nil {
		return nil, err
	}

	err = portal.ValidatePassword(input.Password, true)
	if err != nil {
		return nil, err
	}

	portal.PasswordHash, err = portal.HashBcrypt(portal.Password)
	if err != nil {
		return nil, err
	}

	err = portal.VerifyConfirm()
	if err != nil {
		return nil, err
	}

	Token, err := portal.GenerateAuthorizationToken(portal.Verify.MemberID)
	if err != nil {
		return nil, err
	}

	return &model.VerifyResponse{
		Status: true,
		Token:  &Token,
	}, nil
}

func (r *mutationResolver) PhoneVerify(ctx context.Context, input model.PhoneVerifyRequest) (*model.PhoneVerifyResponse, error) {
	var err error

	portal := portal.Portal{
		Ctx:       ctx,
		Timestamp: time.Now().Unix(),
		Timeout:   60,
		Action:    constants.ACTION_SIGN_CONTRACT,
		Method:    constants.METHOD_PHONE,
	}

	err = portal.GetMember()
	if err != nil {
		return nil, err
	}

	// err = portal.ParsePhone(offer.Member.Phone)
	// if err != nil {
	// 	return nil, err
	// }

	// err = portal.CheckPhoneInUse()
	// if err != nil {
	// 	return nil, err
	// }

	Timeout, err := portal.CreateVerify()

	return &model.PhoneVerifyResponse{
		Timeout: Timeout,
	}, nil
}

func (r *mutationResolver) Verify(ctx context.Context, input model.VerifyRequest) (*model.VerifyResponse, error) {
	var err error

	portal := portal.Portal{
		Ctx:            ctx,
		ProfileHandler: r.ProfileHandler,
		Timestamp:      time.Now().Unix(),
		Timeout:        60,
	}

	err = portal.ValidateVerifyAction(input.Action)
	if err != nil {
		return nil, err
	}

	err = portal.ValidateVerifyMethod(input.Method)
	if err != nil {
		return nil, err
	}

	err = portal.QueryVerifyByHash(input.Hash)
	if err != nil {
		return nil, err
	}

	err = portal.ValidateOTPCode(input.Code)
	if err != nil {
		return nil, err
	}

	if portal.Action == constants.ACTION_RESET {
		return &model.VerifyResponse{
			Status: true,
		}, nil
	}

	err = portal.VerifyConfirm()
	if err != nil {
		return nil, err
	}

	Token, err := portal.GenerateAuthorizationToken(portal.Verify.MemberID)
	if err != nil {
		return nil, err
	}

	return &model.VerifyResponse{
		Status: true,
		Token:  &Token,
	}, nil
}

func (r *mutationResolver) VerifyResend(ctx context.Context, input model.VerifyResendRequest) (*model.VerifyResendResponse, error) {
	var err error

	portal := portal.Portal{
		Ctx:       ctx,
		Timestamp: time.Now().Unix(),
		Timeout:   60,
		Action:    constants.ACTION_SIGNUP,
		Method:    constants.METHOD_EMAIL,
	}

	err = portal.ParseEmail(input.Email)
	if err != nil {
		return nil, err
	}

	err = portal.QueryMemberByEmail()
	if err != nil {
		return nil, err
	}

	err = portal.CheckMemberStatusIsNoActive()
	if err != nil {
		return nil, err
	}

	Timeout, err := portal.CreateVerify()
	if err != nil {
		return &model.VerifyResendResponse{
			Status:  false,
			Timeout: &Timeout,
		}, nil
	}

	err = portal.GenerateConfirmationLink()
	if err != nil {
		return nil, err
	}

	mail.SendMail(portal.Email, "Account activation", "Dear customer,", "Please activate your account by clicking on the button below.", "Activate account", portal.ConfirmationLink, "Thank you for joining our platform.")

	return &model.VerifyResendResponse{
		Status:  true,
		Timeout: &Timeout,
	}, nil
}

func (r *mutationResolver) OfferPhoneVerify(ctx context.Context, input model.OfferPhoneVerifyRequest) (*model.Result, error) {
	var err error

	portal := portal.Portal{
		Ctx:       ctx,
		Timestamp: time.Now().Unix(),
		Timeout:   60,
		Action:    constants.ACTION_SIGN_CONTRACT,
		Method:    constants.METHOD_PHONE,
	}

	err = portal.GetMember()
	if err != nil {
		return nil, err
	}

	err = portal.QueryOffer(input.OfferID)
	if err != nil {
		return nil, err
	}

	_, err = portal.QueryDealByOfferID(input.OfferID)
	if err != nil {
		return nil, err
	}

	err = portal.QueryVerifyByOfferID(input.OfferID)
	if err != nil {
		return nil, err
	}

	err = portal.ValidateOTPCode(input.Code)
	if err != nil {
		return nil, err
	}

	err = portal.OfferPhoneVerify(input.Code)
	if err != nil {
		return nil, err
	}

	err = portal.CreateInvoice()
	if err != nil {
		return nil, err
	}

	err = portal.GenerateInvoiceLink()
	if err != nil {
		return nil, err
	}

	mail.SendMail(portal.Member.Email, "Invoice #"+cast.ToString(portal.Invoice.InvoiceID), "Dear customer,", "Please pay your invoice by clicking on the button below.", "View invoice", portal.InvoiceLink, "Thank you for using our platform.")

	return &model.Result{
		Status: true,
	}, nil
}

func (r *mutationResolver) OfferPhoneVerifyResend(ctx context.Context, input model.RecordRequest) (*model.PhoneVerifyResponse, error) {
	var err error

	portal := portal.Portal{
		Ctx:       ctx,
		Timestamp: time.Now().Unix(),
		Timeout:   60,
		Action:    constants.ACTION_SIGN_CONTRACT,
		Method:    constants.METHOD_PHONE,
	}

	err = portal.GetMember()
	if err != nil {
		return nil, err
	}

	Timeout, err := portal.CreateVerify()

	return &model.PhoneVerifyResponse{
		Timeout: Timeout,
	}, nil
}

func (r *queryResolver) Invest(ctx context.Context, input model.RecordRequest) (*model.Invest, error) {
	panic(fmt.Errorf("not implemented"))
	return nil, nil
}

func (r *queryResolver) InvestByOfferID(ctx context.Context, input *model.RecordRequest) (*model.Invest, error) {
	var err error

	portal := portal.Portal{
		Ctx: ctx,
	}

	res, err := portal.QueryInvestByOfferID(input.RecordID)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (r *queryResolver) InterestListByOfferID(ctx context.Context, input *model.RecordRequest) ([]*model.Interest, error) {
	var err error

	portal := portal.Portal{
		Ctx: ctx,
	}

	res, err := portal.QueryInterestListByOfferID(input.RecordID)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (r *queryResolver) ManagerDealByContractID(ctx context.Context, input *model.ManagerDealByContractIDRequest) (*model.Deal, error) {
	var err error

	manager := manager.Manager{
		Ctx:       ctx,
		Timestamp: time.Now().Unix(),
	}

	res, err := manager.DealByContractID(input)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (r *queryResolver) Offer(ctx context.Context, input model.RecordRequest) (*model.Offer, error) {
	var err error

	portal := portal.Portal{
		Ctx: ctx,
	}

	err = portal.GetMember()
	if err != nil {
		return nil, err
	}

	err = portal.QueryOffer(input.RecordID)
	if err != nil {
		return nil, err
	}

	err = portal.QueryCurrentContract()
	if err != nil {
		return nil, err
	}

	err = portal.ParseContractContent()
	if err != nil {
		return nil, err
	}

	return &model.Offer{
		Title:       &portal.Invest.Title,
		Subtitle:    &portal.Invest.Subtitle,
		Description: &portal.Invest.Description,
		Contract: &model.Contract{
			Content: &portal.Contract.Content,
		},
	}, nil
}

func (r *queryResolver) ContractByOfferID(ctx context.Context, input model.RecordRequest) (*model.Contract, error) {
	var err error

	portal := portal.Portal{
		Ctx: ctx,
	}

	err = portal.GetMember()
	if err != nil {
		return nil, err
	}

	err = portal.QueryOffer(input.RecordID)
	if err != nil {
		return nil, err
	}

	err = portal.QueryCurrentContract()
	if err != nil {
		return nil, err
	}

	err = portal.ParseContractContent()
	if err != nil {
		return nil, err
	}

	return &model.Contract{
		Content: &portal.Contract.Content,
	}, nil
}

func (r *queryResolver) DealByOfferID(ctx context.Context, input model.RecordRequest) (*model.Deal, error) {
	var err error

	portal := portal.Portal{
		Ctx: ctx,
	}

	err = portal.GetMember()
	if err != nil {
		return nil, err
	}

	//TODO: check if there is a signed, paid or active deal

	res, err := portal.QueryDealByOfferID(input.RecordID)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (r *queryResolver) InterestByDealID(ctx context.Context, input *model.RecordRequest) (*model.Interest, error) {
	var err error

	portal := portal.Portal{
		Ctx: ctx,
	}

	err = portal.GetMember()
	if err != nil {
		return nil, err
	}

	_, err = portal.QueryDeal(input.RecordID)
	if err != nil {
		return nil, err
	}

	res, err := portal.QueryInterestByDealID()
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (r *queryResolver) InterestByOfferID(ctx context.Context, input *model.RecordRequest) (*model.Interest, error) {
	var err error

	portal := portal.Portal{
		Ctx: ctx,
	}

	err = portal.GetMember()
	if err != nil {
		return nil, err
	}

	err = portal.QueryOffer(input.RecordID)
	if err != nil {
		return nil, err
	}

	res, err := portal.QueryInterestByOfferID()
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (r *queryResolver) Invoice(ctx context.Context, input model.RecordRequest) (*model.Invoice, error) {
	var err error

	portal := portal.Portal{
		Ctx: ctx,
	}

	err = portal.GetMember()
	if err != nil {
		return nil, err
	}

	res, err := portal.QueryInvoice(input.RecordID)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (r *queryResolver) InvoiceByDealID(ctx context.Context, input *model.RecordRequest) (*model.Invoice, error) {
	var err error

	portal := portal.Portal{
		Ctx: ctx,
	}

	err = portal.GetMember()
	if err != nil {
		return nil, err
	}

	res, err := portal.QueryInvoiceByDealID(input.RecordID)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (r *queryResolver) BankDetailsByInvoiceID(ctx context.Context, input *model.RecordRequest) (*model.BankDetails, error) {
	var err error

	portal := portal.Portal{
		Ctx: ctx,
	}

	err = portal.GetMember()
	if err != nil {
		return nil, err
	}

	_, err = portal.QueryInvoice(input.RecordID)
	if err != nil {
		return nil, err
	}

	OfferID := cast.ToInt(portal.Invoice.OfferID)

	err = portal.QueryOffer(OfferID)
	if err != nil {
		return nil, err
	}

	BankDetailsID := cast.ToInt(portal.Offer.BankDetailsID)

	res, err := portal.QueryBankDetails(BankDetailsID)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (r *queryResolver) OfferByInvoiceID(ctx context.Context, input model.RecordRequest) (*model.Invest, error) {
	var err error

	portal := portal.Portal{
		Ctx: ctx,
	}

	err = portal.GetMember()
	if err != nil {
		return nil, err
	}

	_, err = portal.QueryInvoice(input.RecordID)
	if err != nil {
		return nil, err
	}

	OfferID := cast.ToInt(portal.Invoice.OfferID)

	res, err := portal.QueryInvestByOfferID(OfferID)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (r *queryResolver) ManagerBankDetails(ctx context.Context, input model.RecordRequest) (*model.BankDetails, error) {
	var err error

	manager := manager.Manager{
		Ctx:       ctx,
		Timestamp: time.Now().Unix(),
	}

	res, err := manager.QueryBankDetails(input.RecordID)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (r *queryResolver) ManagerBankDetailsByOfferID(ctx context.Context, input *model.RecordRequest) (*model.BankDetails, error) {
	var err error

	manager := manager.Manager{
		Ctx:       ctx,
		Timestamp: time.Now().Unix(),
	}

	_, err = manager.QueryOffer(input.RecordID)
	if err != nil {
		return nil, err
	}

	res, err := manager.QueryBankDetailsByOfferID()
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (r *queryResolver) ManagerInvest(ctx context.Context, input model.RecordRequest) (*model.Invest, error) {
	var err error

	manager := manager.Manager{
		Ctx:       ctx,
		Timestamp: time.Now().Unix(),
	}

	invest, err := manager.QueryInvest(input.RecordID)
	if err != nil {
		return nil, err
	}

	return invest, nil
}

func (r *queryResolver) ManagerMediaByInvestID(ctx context.Context, input model.ManagerMediaByInvestIDRequest) ([]*model.Media, error) {
	var err error

	manager := manager.Manager{
		Ctx:       ctx,
		Timestamp: time.Now().Unix(),
	}

	_, err = manager.QueryInvest(input.RecordID)
	if err != nil {
		return nil, err
	}

	res, err := manager.QueryMediaByInvestID(input)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (r *queryResolver) ManagerInvestByOfferID(ctx context.Context, input *model.RecordRequest) (*model.Invest, error) {
	var err error

	manager := manager.Manager{
		Ctx:       ctx,
		Timestamp: time.Now().Unix(),
	}

	_, err = manager.QueryOffer(input.RecordID)
	if err != nil {
		return nil, err
	}

	res, err := manager.QueryInvestByOfferID()
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (r *queryResolver) ManagerOffer(ctx context.Context, input model.RecordRequest) (*model.ManagerOffer, error) {
	var err error

	manager := manager.Manager{
		Ctx:       ctx,
		Timestamp: time.Now().Unix(),
	}

	res, err := manager.QueryOffer(input.RecordID)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (r *queryResolver) ManagerDeal(ctx context.Context, input model.RecordRequest) (*model.Deal, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) ManagerInvoiceByDealID(ctx context.Context, input *model.RecordRequest) (*model.Invoice, error) {
	var err error

	manager := manager.Manager{
		Ctx:       ctx,
		Timestamp: time.Now().Unix(),
	}

	res, err := manager.QueryInvoiceByDealID(input)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (r *queryResolver) ManagerContract(ctx context.Context, input model.RecordRequest) (*model.Contract, error) {
	var err error

	manager := manager.Manager{
		Ctx:       ctx,
		Timestamp: time.Now().Unix(),
	}

	res, err := manager.QueryContract(input.RecordID)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (r *queryResolver) ManagerCategory(ctx context.Context, input model.RecordRequest) (*model.Category, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) ManagerCurrency(ctx context.Context, input model.RecordRequest) (*model.Currency, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) ManagerMember(ctx context.Context, input model.RecordRequest) (*model.Member, error) {
	var err error

	manager := manager.Manager{
		Ctx:       ctx,
		Timestamp: time.Now().Unix(),
	}

	res, err := manager.QueryMember(input.RecordID)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (r *queryResolver) ManagerMemberByOfferID(ctx context.Context, input model.RecordRequest) (*model.Member, error) {
	var err error

	manager := manager.Manager{
		Ctx:       ctx,
		Timestamp: time.Now().Unix(),
	}

	_, err = manager.QueryOffer(input.RecordID)
	if err != nil {
		return nil, err
	}

	res, err := manager.QueryMember(cast.ToInt(manager.Offer.MemberID))
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (r *queryResolver) ManagerManager(ctx context.Context, input model.RecordRequest) (*model.Member, error) {
	var err error

	manager := manager.Manager{
		Ctx:       ctx,
		Timestamp: time.Now().Unix(),
	}

	res, err := manager.QueryManager(input.RecordID)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (r *queryResolver) ManagerLead(ctx context.Context, input model.RecordRequest) (*model.Lead, error) {
	var err error

	manager := manager.Manager{
		Ctx:       ctx,
		Timestamp: time.Now().Unix(),
	}

	res, err := manager.QueryLead(input.RecordID)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (r *queryResolver) ManagerComment(ctx context.Context, input model.RecordRequest) (*model.Comment, error) {
	var err error

	manager := manager.Manager{
		Ctx:       ctx,
		Timestamp: time.Now().Unix(),
	}

	res, err := manager.QueryComment(input.RecordID)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (r *queryResolver) ManagerChecklist(ctx context.Context, input model.RecordRequest) (*model.Checklist, error) {
	var err error

	manager := manager.Manager{
		Ctx:       ctx,
		Timestamp: time.Now().Unix(),
	}

	res, err := manager.QueryChecklist(input.RecordID)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (r *queryResolver) ManagerAppointment(ctx context.Context, input model.RecordRequest) (*model.Appointment, error) {
	var err error

	manager := manager.Manager{
		Ctx:       ctx,
		Timestamp: time.Now().Unix(),
	}

	res, err := manager.QueryAppointment(input.RecordID)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (r *queryResolver) ManagerCampaign(ctx context.Context, input model.RecordRequest) (*model.Campaign, error) {
	var err error

	manager := manager.Manager{
		Ctx:       ctx,
		Timestamp: time.Now().Unix(),
	}

	res, err := manager.QueryCampaign(input.RecordID)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (r *queryResolver) OfferList(ctx context.Context) ([]*model.Invest, error) {
	var err error

	portal := portal.Portal{
		Ctx: ctx,
	}

	err = portal.GetMember()
	if err != nil {
		return nil, err
	}

	res, err := portal.QueryOfferList()
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (r *queryResolver) DealList(ctx context.Context) ([]*model.Deal, error) {
	var err error

	portal := portal.Portal{
		Ctx: ctx,
	}

	err = portal.GetMember()
	if err != nil {
		return nil, err
	}

	res, err := portal.QueryDealList()
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (r *queryResolver) ContractList(ctx context.Context) ([]*model.Contract, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) CategoryList(ctx context.Context) ([]*model.Category, error) {
	var err error

	portal := portal.Portal{
		Ctx: ctx,
	}

	res, err := portal.QueryCategoryList()
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (r *queryResolver) CurrencyList(ctx context.Context) ([]*model.Currency, error) {
	var err error

	portal := portal.Portal{
		Ctx: ctx,
	}

	res, err := portal.QueryCurrencyList()
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (r *queryResolver) BalanceList(ctx context.Context) ([]*model.Balance, error) {
	var err error

	portal := portal.Portal{
		Ctx: ctx,
	}

	err = portal.GetMember()
	if err != nil {
		return nil, err
	}

	res, err := portal.QueryBalanceList()
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (r *queryResolver) TXList(ctx context.Context) ([]*model.Tx, error) {
	var err error

	portal := portal.Portal{
		Ctx: ctx,
	}

	err = portal.GetMember()
	if err != nil {
		return nil, err
	}

	res, err := portal.QueryTXList()
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (r *queryResolver) ManagerBankDetailsList(ctx context.Context, input *model.ListRequest) ([]*model.BankDetails, error) {
	var err error

	manager := manager.Manager{
		Ctx:       ctx,
		Timestamp: time.Now().Unix(),
	}

	res, err := manager.QueryBankDetailsList() // input.Query
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (r *queryResolver) ManagerInvestList(ctx context.Context, input *model.ListRequest) ([]*model.Invest, error) {
	var err error

	manager := manager.Manager{
		Ctx:       ctx,
		Timestamp: time.Now().Unix(),
	}

	res, err := manager.QueryInvestList()
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (r *queryResolver) ManagerOfferList(ctx context.Context, input *model.ListRequest) ([]*model.ManagerOffer, error) {
	var err error

	manager := manager.Manager{
		Ctx:       ctx,
		Timestamp: time.Now().Unix(),
	}

	res, err := manager.QueryOfferList()
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (r *queryResolver) ManagerDealList(ctx context.Context, input *model.ListRequest) ([]*model.Deal, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) ManagerDealListByOfferID(ctx context.Context, input model.RecordRequest) ([]*model.Deal, error) {
	var err error

	manager := manager.Manager{
		Ctx:       ctx,
		Timestamp: time.Now().Unix(),
	}

	res, err := manager.DealListByOfferID(input)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (r *queryResolver) ManagerContractList(ctx context.Context, input *model.ListRequest) ([]*model.Contract, error) {
	var err error

	manager := manager.Manager{
		Ctx:       ctx,
		Timestamp: time.Now().Unix(),
	}

	res, err := manager.QueryContractList()
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (r *queryResolver) ManagerCategoryList(ctx context.Context, input *model.ListRequest) ([]*model.Category, error) {
	var err error

	manager := manager.Manager{
		Ctx:       ctx,
		Timestamp: time.Now().Unix(),
	}

	res, err := manager.QueryCategoryList()
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (r *queryResolver) ManagerCurrencyList(ctx context.Context, input *model.ListRequest) ([]*model.Currency, error) {
	var err error

	manager := manager.Manager{
		Ctx:       ctx,
		Timestamp: time.Now().Unix(),
	}

	res, err := manager.QueryCurrencyList()
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (r *queryResolver) ManagerLeadList(ctx context.Context, input *model.ListRequest) ([]*model.Lead, error) {
	var err error

	manager := manager.Manager{
		Ctx:       ctx,
		Timestamp: time.Now().Unix(),
	}

	res, err := manager.QueryLeadList()
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (r *queryResolver) ManagerCommentList(ctx context.Context, input *model.ListRequest) ([]*model.Comment, error) {
	var err error

	manager := manager.Manager{
		Ctx:       ctx,
		Timestamp: time.Now().Unix(),
	}

	res, err := manager.QueryCommentList()
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (r *queryResolver) ManagerCommentListByLeadID(ctx context.Context, input *model.RecordRequest) ([]*model.Comment, error) {
	var err error

	manager := manager.Manager{
		Ctx:       ctx,
		Timestamp: time.Now().Unix(),
	}

	res, err := manager.QueryCommentListByLeadID(input.RecordID)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (r *queryResolver) ManagerChecklistList(ctx context.Context, input *model.ListRequest) ([]*model.Checklist, error) {
	var err error

	manager := manager.Manager{
		Ctx:       ctx,
		Timestamp: time.Now().Unix(),
	}

	res, err := manager.QueryChecklistList()
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (r *queryResolver) ManagerAppointmentList(ctx context.Context, input *model.ListRequest) ([]*model.Appointment, error) {
	var err error

	manager := manager.Manager{
		Ctx:       ctx,
		Timestamp: time.Now().Unix(),
	}

	res, err := manager.QueryAppointmentList()
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (r *queryResolver) ManagerAppointmentListByLeadID(ctx context.Context, input *model.RecordRequest) ([]*model.Appointment, error) {
	var err error

	manager := manager.Manager{
		Ctx:       ctx,
		Timestamp: time.Now().Unix(),
	}

	res, err := manager.QueryAppointmentListByLeadID(input.RecordID)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (r *queryResolver) ManagerCampaignList(ctx context.Context, input *model.ListRequest) ([]*model.Campaign, error) {
	var err error

	manager := manager.Manager{
		Ctx:       ctx,
		Timestamp: time.Now().Unix(),
	}

	res, err := manager.QueryCampaignList()
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (r *queryResolver) ManagerManagerList(ctx context.Context, input *model.ListRequest) ([]*model.Member, error) {
	var err error

	manager := manager.Manager{
		Ctx:       ctx,
		Timestamp: time.Now().Unix(),
	}

	res, err := manager.QueryManagerList()
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (r *queryResolver) ManagerContractListByOfferID(ctx context.Context, input *model.RecordRequest) ([]*model.Contract, error) {
	var err error

	manager := manager.Manager{
		Ctx:       ctx,
		Timestamp: time.Now().Unix(),
	}

	res, err := manager.QueryContractListByOfferID(input.RecordID)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (r *queryResolver) ManagerInterestListByOfferID(ctx context.Context, input *model.RecordRequest) ([]*model.Interest, error) {
	var err error

	manager := manager.Manager{
		Ctx:       ctx,
		Timestamp: time.Now().Unix(),
	}

	res, err := manager.QueryInterestListByOfferID(input.RecordID)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (r *queryResolver) ManagerSearchMember(ctx context.Context, input *model.SearchRequest) ([]*model.Member, error) {
	var err error

	manager := manager.Manager{
		Ctx:       ctx,
		Timestamp: time.Now().Unix(),
	}

	res, err := manager.QueryMemberList() // input.Query
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (r *queryResolver) ManagerSearchMemberNoManager(ctx context.Context, input *model.SearchRequest) ([]*model.Member, error) {
	var err error

	manager := manager.Manager{
		Ctx:       ctx,
		Timestamp: time.Now().Unix(),
	}

	res, err := manager.QueryMemberListNoManager() // input.Query
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (r *queryResolver) ManagerSearchInvest(ctx context.Context, input *model.SearchRequest) ([]*model.Invest, error) {
	var err error

	manager := manager.Manager{
		Ctx:       ctx,
		Timestamp: time.Now().Unix(),
	}

	res, err := manager.QueryInvestList() // input.Query
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (r *queryResolver) ManagerSearchContract(ctx context.Context, input *model.SearchRequest) ([]*model.Contract, error) {
	var err error

	manager := manager.Manager{
		Ctx:       ctx,
		Timestamp: time.Now().Unix(),
	}

	res, err := manager.QueryContractList() // input.Query
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (r *queryResolver) ManagerSearchCurrency(ctx context.Context, input *model.SearchRequest) ([]*model.Currency, error) {
	var err error

	manager := manager.Manager{
		Ctx:       ctx,
		Timestamp: time.Now().Unix(),
	}

	res, err := manager.QueryCurrencyList()
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (r *queryResolver) ManagerSearchBankDetails(ctx context.Context, input *model.SearchRequest) ([]*model.BankDetails, error) {
	var err error

	manager := manager.Manager{
		Ctx:       ctx,
		Timestamp: time.Now().Unix(),
	}

	res, err := manager.QueryBankDetailsList() // input.Query
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (r *queryResolver) ManagerSearchManager(ctx context.Context, input *model.SearchRequest) ([]*model.ManagerSearch, error) {
	var err error

	manager := manager.Manager{
		Ctx:       ctx,
		Timestamp: time.Now().Unix(),
	}

	res, err := manager.SearchManager() // input.Query
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (r *queryResolver) Member(ctx context.Context) (*model.Member, error) {
	var err error

	portal := portal.Portal{
		Ctx: ctx,
	}

	err = portal.GetMember()
	if err != nil {
		return nil, err
	}

	res, err := portal.ReplyMember()
	if err != nil {
		return nil, err
	}

	return res, err
}

func (r *queryResolver) Alert(ctx context.Context) ([]*model.Alert, error) {
	db := db.GetDB()

	var err error

	portal := portal.Portal{
		Ctx: ctx,
	}

	err = portal.GetMember()
	if err != nil {
		return nil, err
	}

	var alert []*model.Alert

	err = db.Select(&alert, "SELECT AlertID, AssetID, Price, TimestampOpen FROM Alert WHERE MemberID=$1 AND Status=$2 ORDER BY AlertID DESC", portal.Member.MemberID, true)

	if err != nil {
		if err != sql.ErrNoRows {
			return nil, err
		}
	}

	// var alerts []*model.Alert

	// alert := &model.Alert{
	// 	AlertID: "1",
	// 	AssetID: "1",
	// 	Price:   "1",
	// 	Datetime: &model.Datetime{
	// 		Time:             "a",
	// 		Status:           1,
	// 		InfinityModifier: 1,
	// 	},
	// }
	// alerts = append(alert, alerts)
	return alert, nil
}

func (r *subscriptionResolver) NewInfo(ctx context.Context) (<-chan *model.Info, error) {
	var err error

	portal := portal.Portal{
		Ctx: ctx,
	}

	wsauth := handler.GetInitPayload(ctx).GetString("Authorization")

	MemberID, err := portal.GetMemberIDFromToken(wsauth)
	if err != nil {
		return nil, err
	}

	info := make(chan *model.Info, 1)

	//radix
	ch := redis.Channel{
		Name: "info",
	}

	go func() {
		ch.MsgCh, err = ch.GetInfoChannel()
		if err == nil {
			for msg := range ch.MsgCh {

				var infoMsg *model.Info

				err := json.Unmarshal(msg.Message, &infoMsg)
				if err == nil {
					if ((infoMsg.Event == "trade" || infoMsg.Event == "alert") && MemberID == cast.ToString(infoMsg.MemberID)) || infoMsg.Event == "rate" {
						info <- infoMsg
					}
				}
			}
		}
	}()
	//radix

	go func() {
		<-ctx.Done()
	}()

	return info, nil
}

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

// Subscription returns generated.SubscriptionResolver implementation.
func (r *Resolver) Subscription() generated.SubscriptionResolver { return &subscriptionResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
type subscriptionResolver struct{ *Resolver }

// !!! WARNING !!!
// The code below was going to be deleted when updating resolvers. It has been copied here so you have
// one last chance to move it out of harms way if you want. There are two reasons this happens:
//  - When renaming or deleting a resolver the old code will be put in here. You can safely delete
//    it when you're done.
//  - You have helper methods in this file. Move them out to keep these resolver files clean.
func (r *queryResolver) ManagerDragPhoto(ctx context.Context, input *model.DragRequest) (*model.Result, error) {
	panic(fmt.Errorf("not implemented"))
}
