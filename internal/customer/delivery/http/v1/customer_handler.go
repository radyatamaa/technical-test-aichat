package v1

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	beego "github.com/beego/beego/v2/server/web"
	"github.com/radyatamaa/technical-test-aichat/internal"
	"github.com/radyatamaa/technical-test-aichat/internal/domain"
	"github.com/radyatamaa/technical-test-aichat/pkg/response"
	"github.com/radyatamaa/technical-test-aichat/pkg/zaplogger"
	"gorm.io/gorm"
)

type CustomerHandler struct {
	ZapLogger zaplogger.Logger
	internal.BaseController
	response.ApiResponse
	CustomerUsecase domain.CustomerUseCase
}

func NewCustomerHandler(customerUsecase domain.CustomerUseCase, zapLogger zaplogger.Logger) {
	pHandler := &CustomerHandler{
		ZapLogger:       zapLogger,
		CustomerUsecase: customerUsecase,
	}
	beego.Router("/api/v1/verify-photo/:id", pHandler, "post:VerifyPhoto")
	beego.Router("/api/v1/link-voucher/:id", pHandler, "get:GetLinkVoucher")
}

func (h *CustomerHandler) Prepare() {
	// check user access when needed
	h.SetLangVersion()
}

// VerifyPhoto
// @Title VerifyPhoto
// @Tags Customer
// @Summary VerifyPhoto
// @Produce json
// @Param Accept-Language header string false "lang"
// @Success 200 {object} swagger.BaseResponse{errors=[]object,data=domain.CustomerVerifyPhotoResponse}
// @Failure 400 {object} swagger.BadRequestErrorValidationResponse{errors=[]swagger.ValidationErrors,data=object}
// @Failure 408 {object} swagger.RequestTimeoutResponse{errors=[]object,data=object}
// @Failure 500 {object} swagger.InternalServerErrorResponse{errors=[]object,data=object}
// @Param        file   formData  file    true  "file"
// @Param    id path int true "id customer"
// @Router /v1/verify-photo/{id} [post]
func (h *CustomerHandler) VerifyPhoto() {
	pathParam, err := strconv.Atoi(h.Ctx.Input.Param(":id"))

	if err != nil || pathParam < 1 {
		h.ResponseError(h.Ctx, http.StatusBadRequest, response.PathParamInvalidCode, response.ErrorCodeText(response.PathParamInvalidCode, h.Locale.Lang), err)
		return
	}
	_, fileHeader, err := h.GetFile("file")
	if err != nil {
		h.Ctx.Input.SetData("stackTrace", h.ZapLogger.SetMessageLog(err))
		h.ResponseError(h.Ctx, http.StatusBadRequest, response.ApiValidationCodeError, response.ErrorCodeText(response.ApiValidationCodeError, h.Locale.Lang), err)
		return
	}

	result, err := h.CustomerUsecase.VerifyPhotoCustomer(h.Ctx, pathParam, fileHeader)
	if err != nil {
		if errors.Is(err, response.ErrCustomerAlreadyGetVoucher) {
			h.ResponseError(h.Ctx, http.StatusBadRequest, response.CustomerAlreadyGetVoucher, response.ErrorCodeText(response.CustomerAlreadyGetVoucher, h.Locale.Lang), err)
			return
		}
		if errors.Is(err, response.ErrCustomerVerifyImage) {
			h.ResponseError(h.Ctx, http.StatusBadRequest, response.CustomerVerifyImage, response.ErrorCodeText(response.CustomerVerifyImage, h.Locale.Lang), err)
			return
		}
		if errors.Is(err, response.ErrCustomerNotYetBookVoucher) {
			h.ResponseError(h.Ctx, http.StatusBadRequest, response.CustomerNotYetBookVoucher, response.ErrorCodeText(response.CustomerNotYetBookVoucher, h.Locale.Lang), err)
			return
		}
		if errors.Is(err, response.ErrCustomerBookVoucherExpired) {
			h.ResponseError(h.Ctx, http.StatusBadRequest, response.CustomerBookVoucherExpired, response.ErrorCodeText(response.CustomerBookVoucherExpired, h.Locale.Lang), err)
			return
		}
		if errors.Is(err, context.DeadlineExceeded) {
			h.ResponseError(h.Ctx, http.StatusRequestTimeout, response.RequestTimeoutCodeError, response.ErrorCodeText(response.RequestTimeoutCodeError, h.Locale.Lang), err)
			return
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			h.ResponseError(h.Ctx, http.StatusBadRequest, response.DataNotFoundCodeError, response.ErrorCodeText(response.DataNotFoundCodeError, h.Locale.Lang), err)
			return
		}
		h.ResponseError(h.Ctx, http.StatusInternalServerError, response.ServerErrorCode, response.ErrorCodeText(response.ServerErrorCode, h.Locale.Lang), err)
		return
	}
	h.Ok(h.Ctx, h.Tr("message.success"), result)
	return
}

// GetLinkVoucher
// @Title GetLinkVoucher
// @Tags Customer
// @Summary GetLinkVoucher
// @Produce json
// @Param Accept-Language header string false "lang"
// @Success 200 {object} swagger.BaseResponse{data=[]domain.CustomerVoucherBookResponse,errors=[]object}
// @Failure 408 {object} swagger.RequestTimeoutResponse{errors=[]object,data=object}
// @Failure 500 {object} swagger.InternalServerErrorResponse{errors=[]object,data=object}
// @Param    id path int true "id customer"
// @router /v1/link-voucher/{id} [get]
func (h *CustomerHandler) GetLinkVoucher() {
	pathParam, err := strconv.Atoi(h.Ctx.Input.Param(":id"))

	if err != nil || pathParam < 1 {
		h.ResponseError(h.Ctx, http.StatusBadRequest, response.PathParamInvalidCode, response.ErrorCodeText(response.PathParamInvalidCode, h.Locale.Lang), err)
		return
	}

	result, err := h.CustomerUsecase.GetVoucherByCustomerId(h.Ctx, pathParam)
	if err != nil {
		if errors.Is(err, response.ErrVoucherNotAvailable) {
			h.ResponseError(h.Ctx, http.StatusBadRequest, response.VoucherNotAvailable, response.ErrorCodeText(response.VoucherNotAvailable, h.Locale.Lang), err)
			return
		}
		if errors.Is(err, response.ErrTransactionCompletePurchase30Days) {
			h.ResponseError(h.Ctx, http.StatusBadRequest, response.TransactionCompletePurchase30Days, response.ErrorCodeText(response.TransactionCompletePurchase30Days, h.Locale.Lang), err)
			return
		}
		if errors.Is(err, response.ErrTransactionMinimum) {
			h.ResponseError(h.Ctx, http.StatusBadRequest, response.TransactionMinimum, response.ErrorCodeText(response.TransactionMinimum, h.Locale.Lang), err)
			return
		}
		if errors.Is(err, response.ErrCustomerAlreadyBookVoucher) {
			h.ResponseError(h.Ctx, http.StatusBadRequest, response.CustomerAlreadyBookVoucher, response.ErrorCodeText(response.CustomerAlreadyBookVoucher, h.Locale.Lang), err)
			return
		}
		if errors.Is(err, response.ErrCustomerAlreadyGetVoucher) {
			h.ResponseError(h.Ctx, http.StatusBadRequest, response.CustomerAlreadyGetVoucher, response.ErrorCodeText(response.CustomerAlreadyGetVoucher, h.Locale.Lang), err)
			return
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			h.ResponseError(h.Ctx, http.StatusBadRequest, response.DataNotFoundCodeError, response.ErrorCodeText(response.DataNotFoundCodeError, h.Locale.Lang), err)
			return
		}
		if errors.Is(err, context.DeadlineExceeded) {
			h.ResponseError(h.Ctx, http.StatusRequestTimeout, response.RequestTimeoutCodeError, response.ErrorCodeText(response.RequestTimeoutCodeError, h.Locale.Lang), err)
			return
		}
		h.ResponseError(h.Ctx, http.StatusInternalServerError, response.ServerErrorCode, response.ErrorCodeText(response.ServerErrorCode, h.Locale.Lang), err)
		return
	}
	h.Ok(h.Ctx, h.Tr("message.success"), result)
	return
}
