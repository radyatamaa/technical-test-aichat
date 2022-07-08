package domain

type CustomerVoucherBookResponse struct {
	Expired string `json:"expired"`
}

type CustomerVerifyPhotoResponse struct {
	VoucherCode string `json:"voucher_code"`
}
