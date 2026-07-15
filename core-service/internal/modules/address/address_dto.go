package address

type CreateAddressRequest struct {
	Label       string `json:"label"`
	FullAddress string `json:"full_address"`
	City        string `json:"city"`
	District    string `json:"district"`
	PostalCode  string `json:"postal_code"`
	IsPrimary   bool   `json:"is_primary"`
}

type UpdateAddressRequest struct {
	Label       string `json:"label"`
	FullAddress string `json:"full_address"`
	City        string `json:"city"`
	District    string `json:"district"`
	PostalCode  string `json:"postal_code"`
	IsPrimary   bool   `json:"is_primary"`
}

type AddressResponse struct {
	ID          int    `json:"id"`
	UserID      int    `json:"user_id"`
	Label       string `json:"label"`
	FullAddress string `json:"full_address"`
	City        string `json:"city"`
	District    string `json:"district"`
	PostalCode  string `json:"postal_code"`
	IsPrimary   bool   `json:"is_primary"`
}
