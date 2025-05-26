package dtos

type FormCrmCustomerRequest struct {
	ID             *uint   `json:"id"`
	CustomerTypeID *uint   `json:"customer_type_id"`
	AgentID        *uint   `json:"agent_id"`
	CurrencyID     *uint   `json:"currency_id"`
	Shortname      *string `json:"shortname"`
	Code           *string `json:"code"`
	Name           string  `json:"name"`
	Address        *string `json:"address"`
	Phone          *string `json:"phone"`
	Email          *string `json:"email"`
	Pic            *string `json:"pic"`
	Status         int8    `json:"status"`
	IsCrm          int8    `json:"is_crm"`

	Remark            *string                        `json:"remark"`
	OwnerName         *string                        `json:"owner_name"`
	OwnerPhone        *string                        `json:"owner_phone"`
	OwnerEmail        *string                        `json:"owner_email"`
	CategoryTypeID    *uint                          `json:"category_type_id"`
	ContractDate      *string                        `json:"contract_date"`
	IsContract        *int8                          `json:"is_contract"`
	PicName           *string                        `json:"pic_name"`
	PicPhone          *string                        `json:"pic_phone"`
	PicEmails         []FormCustomerPICEmailsRequest `json:"pic_emails"`
	CustomerContracts []FormCustomerContractsRequest `json:"customer_contracts"`
}

type FormCustomerPICEmailsRequest struct {
	ID          *uint   `json:"id" db:"id"`
	Name        *string `json:"name" db:"name"`
	IsMain      *int8   `json:"is_main" db:"is_main"`
	CreatedByID *uint   `json:"created_by_id" db:"created_by_id"`
	UpdatedByID *uint   `json:"updated_by_id" db:"updated_by_id"`
	DeletedByID *uint   `json:"deleted_by_id" db:"deleted_by_id"`
}

type FormCustomerContractsRequest struct {
	ID         *uint `json:"id" db:"id"`
	CustomerID *uint `json:"customer_id" db:"customer_id"`
	ProductID  *uint `json:"product_id" db:"product_id"`

	AgreeAt       *string  `json:"agree_at" db:"agree_at"`
	DueAt         *string  `json:"due_at" db:"due_at"`
	Price         *float64 `json:"price" db:"price"`
	PaymentTypeID *uint    `json:"payment_type_id" db:"payment_type_id"`

	Qty            *float64 `json:"qty" db:"qty"`
	InstallationAt *string  `json:"installation_at" db:"installation_at"`
	WarrantyAt     *string  `json:"warranty_at" db:"warranty_at"`
	Remark         *string  `json:"remark" db:"remark"`

	CreatedByID *uint `json:"created_by_id" db:"created_by_id"`
	UpdatedByID *uint `json:"updated_by_id" db:"updated_by_id"`
	DeletedByID *uint `json:"deleted_by_id" db:"deleted_by_id"`

	ProductName      *string `json:"product_name" db:"product_name"`
	PaymentTypeName  *string `json:"payment_type_name" db:"payment_type_name"`
	ItemGroupName    *string `json:"item_group_name" db:"item_group_name"`
	ItemSubGroupName *string `json:"item_sub_group_name" db:"item_sub_group_name"`
}
