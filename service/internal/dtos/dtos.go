package dtos

import (
	"time"
)

type GetUsersRequest struct {
	Global         string  `json:"global"`
	Username       string  `json:"username"`
	Name           string  `json:"name"`
	Email          string  `json:"email"`
	PerPage        *string `json:"per_page" default:"10"`         // Default per_page to 10
	Page           *string `json:"page" default:"1"`              // Default page to 1
	OrderColumn    string  `json:"order_column" default:"id"`     // Default order column to "id"
	OrderDirection string  `json:"order_direction" default:"asc"` // Default order direction to "asc"
}

type CreateUserRequest struct {
	Name     string   `json:"name"`
	Username *string  `json:"username"`
	Email    string   `json:"email"`
	Address  *string  `json:"address"`
	Password string   `json:"password"`
	RoleIDs  []uint32 `json:"role_ids"`
}

type UpdateUserRequest struct {
	ID       uint     `json:"id"`
	Username *string  `json:"username"`
	Name     string   `json:"name"`
	Email    string   `json:"email"`
	Address  *string  `json:"address"`
	Password *string  `json:"password"`
	RoleIDs  []uint32 `json:"role_ids"`
}

type GetUserByIDParams struct {
	ID        uint `json:"id"`
	IsDeleted *int
}

type GetUserByIDRequest struct {
	ID uint `json:"id"`
}

type DeleteUserRequest struct {
	ID uint `json:"id"`
}

type UserListDTO struct {
	ID       int     `json:"id"`
	Username *string `json:"username"`
	Name     string  `json:"name"`
	Email    string  `json:"email"`
}

type UserDetailDTO struct {
	ID          uint     `json:"id"`
	Name        string   `json:"name"`
	Username    *string  `json:"username"`
	Email       string   `json:"email"`
	Address     *string  `json:"address"`
	Password    *string  `json:"password"`
	Roles       []string `json:"roles"`
	Permissions []string `json:"permissions"`
	CreatedAt   *string  `json:"created_at"`
}
type GetUsersResult struct {
	Users []UserListDTO
	Total int
	Err   error
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
type RegisterRequest struct {
	Name     string  `json:"name"`
	Username *string `json:"username"`
	Email    string  `json:"email"`
	Address  *string `json:"address"`
	Password string  `json:"password"`
}

type CreateIdentifierRequest struct {
	UserID           uint   `json:"user_id"`
	TypeIdentifierID uint   `json:"type_identifier_id"`
	RefNum           string `json:"ref_num"`
	Status           uint   `json:"status"`
}

type UpdateIdentifierRequest struct {
	ID               uint   `json:"id"`
	UserID           uint   `json:"user_id"`
	TypeIdentifierID *uint  `json:"type_identifier_id"`
	RefNum           string `json:"ref_num"`
	Status           uint   `json:"status"`
}

type GetIdentifierByIDRequest struct {
	ID uint `json:"id"`
}

type GetIdentifierParams struct {
	ID        uint
	UserID    uint
	IsDeleted *int
}

func NewGetIdentifierParams(id uint) *GetIdentifierParams {
	defaultIsDeleted := 0
	return &GetIdentifierParams{
		ID:        id,
		IsDeleted: &defaultIsDeleted,
	}
}

type DeleteIdentifierRequest struct {
	ID uint `json:"id"`
}

type IdentifierListDTO struct {
	ID                 int     `json:"id" db:"id"`
	UserID             uint    `json:"user_id" db:"user_id"`
	UserName           string  `json:"user_name" db:"user_name"`
	TypeIdentifierID   uint    `json:"type_identifier_id" db:"type_identifier_id"`
	TypeIdentifierName string  `json:"type_identifier_name" db:"type_identifier_name"`
	RefNum             string  `json:"ref_num" db:"ref_num"`
	Status             uint    `json:"status" db:"status"`
	CreatedAt          *string `json:"created_at" db:"created_at"`
	UpdatedAt          *string `json:"updated_at" db:"updated_at"`
}

type IdentifierDetailDTO struct {
	ID                 uint       `json:"id" db:"id"`
	UserID             uint       `json:"user_id" db:"user_id"`
	UserName           string     `json:"user_name" db:"user_name"`
	TypeIdentifierID   uint       `json:"type_identifier_id" db:"type_identifier_id"`
	TypeIdentifierName string     `json:"type_identifier_name" db:"type_identifier_name"`
	RefNum             string     `json:"ref_num" db:"ref_num"`
	Status             uint       `json:"status" db:"status"`
	CreatedAt          *time.Time `json:"created_at" db:"created_at"`
	UpdatedAt          *time.Time `json:"updated_at" db:"updated_at"`
	DeletedAt          *time.Time `json:"deleted_at" db:"deleted_at"`
}
type ListIdentifiersResult struct {
	Identifiers []IdentifierListDTO
	Total       int
	Err         error
}

type CreateContactRequest struct {
	TypeContactID uint   `json:"type_contact_id"`
	UserID        uint   `json:"user_id"`
	RefNum        string `json:"ref_num"`
	Status        uint   `json:"status"`
}

type UpdateContactRequest struct {
	ID            uint   `json:"id"`
	UserID        uint   `json:"user_id"`
	TypeContactID *uint  `json:"type_contact_id"`
	RefNum        string `json:"ref_num"`
	Status        uint   `json:"status"`
}

type GetContactByIDRequest struct {
	ID uint `json:"id"`
}

type GetContactParams struct {
	ID        uint
	UserID    uint
	IsDeleted *int
}

func NewGetContactParams(id uint) *GetContactParams {
	defaultIsDeleted := 0
	return &GetContactParams{
		ID:        id,
		IsDeleted: &defaultIsDeleted,
	}
}

type DeleteContactRequest struct {
	ID uint `json:"id"`
}

type ContactListDTO struct {
	ID              int     `json:"id" db:"id"`
	UserID          uint    `json:"user_id" db:"user_id"`
	UserName        string  `json:"user_name" db:"user_name"`
	TypeContactID   uint    `json:"type_contact_id" db:"type_contact_id"`
	TypeContactName string  `json:"type_contact_name" db:"type_contact_name"`
	RefNum          string  `json:"ref_num" db:"ref_num"`
	Status          uint    `json:"status" db:"status"`
	CreatedAt       *string `json:"created_at" db:"created_at"`
	UpdatedAt       *string `json:"updated_at" db:"updated_at"`
}

type ContactDetailDTO struct {
	ID              uint       `json:"id" db:"id"`
	UserID          uint       `json:"user_id" db:"user_id"`
	UserName        string     `json:"user_name" db:"user_name"`
	TypeContactID   uint       `json:"type_contact_id" db:"type_contact_id"`
	TypeContactName string     `json:"type_contact_name" db:"type_contact_name"`
	RefNum          string     `json:"ref_num" db:"ref_num"`
	Status          uint       `json:"status" db:"status"`
	CreatedAt       *time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       *time.Time `json:"updated_at" db:"updated_at"`
	DeletedAt       *time.Time `json:"deleted_at" db:"deleted_at"`
}
type ListContactsResult struct {
	Contacts []ContactListDTO
	Total    int
	Err      error
}

type CreateAddressRequest struct {
	TypeAddressID uint   `json:"type_address_id"`
	UserID        uint   `json:"user_id"`
	RefNum        string `json:"ref_num"`
	Status        uint   `json:"status"`
}

type UpdateAddressRequest struct {
	ID            uint   `json:"id"`
	UserID        uint   `json:"user_id"`
	TypeAddressID *uint  `json:"type_address_id"`
	RefNum        string `json:"ref_num"`
	Status        uint   `json:"status"`
}

type GetAddressByIDRequest struct {
	ID uint `json:"id"`
}

type GetAddressParams struct {
	ID        uint
	UserID    uint
	IsDeleted *int
}

func NewGetAddressParams(id uint) *GetAddressParams {
	defaultIsDeleted := 0
	return &GetAddressParams{
		ID:        id,
		IsDeleted: &defaultIsDeleted,
	}
}

type DeleteAddressRequest struct {
	ID uint `json:"id"`
}

type AddressListDTO struct {
	ID              int     `json:"id" db:"id"`
	UserID          uint    `json:"user_id" db:"user_id"`
	UserName        string  `json:"user_name" db:"user_name"`
	TypeAddressID   uint    `json:"type_address_id" db:"type_address_id"`
	TypeAddressName string  `json:"type_address_name" db:"type_address_name"`
	RefNum          string  `json:"ref_num" db:"ref_num"`
	Status          uint    `json:"status" db:"status"`
	CreatedAt       *string `json:"created_at" db:"created_at"`
	UpdatedAt       *string `json:"updated_at" db:"updated_at"`
}

type AddressDetailDTO struct {
	ID              uint       `json:"id" db:"id"`
	UserID          uint       `json:"user_id" db:"user_id"`
	UserName        string     `json:"user_name" db:"user_name"`
	TypeAddressID   uint       `json:"type_address_id" db:"type_address_id"`
	TypeAddressName string     `json:"type_address_name" db:"type_address_name"`
	RefNum          string     `json:"ref_num" db:"ref_num"`
	Status          uint       `json:"status" db:"status"`
	CreatedAt       *time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       *time.Time `json:"updated_at" db:"updated_at"`
	DeletedAt       *time.Time `json:"deleted_at" db:"deleted_at"`
}
type ListAddressesResult struct {
	Addresses []AddressListDTO
	Total     int
	Err       error
}

// type Scheduler struct {
// 	ID          uint       `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
// 	Name        string     `json:"name" gorm:"column:name"`
// 	Description *string     `json:"description" gorm:"column:description"`
// 	Cron        string     `json:"cron" gorm:"column:cron"`
// 	Payload     string     `json:"payload" gorm:"column:payload"`
// 	Status      string     `json:"status" gorm:"column:status"`
// 	StartAt     time.Time  `json:"start_at" gorm:"column:start_at"`
// 	EndAt       *time.Time `json:"end_at" gorm:"column:end_at"`
// }

type SchedulerListDTO struct {
	ID          int     `json:"id" db:"id"`
	Name        string  `json:"name" db:"name"`
	Description *string `json:"description" db:"description"`
	Cron        string  `json:"cron" db:"cron"`
	Payload     string  `json:"payload" db:"payload"`
	Status      string  `json:"status" db:"status"`
	StartAt     *string `json:"start_at" db:"start_at"`
	EndAt       *string `json:"end_at" db:"end_at"`
}

type GetItemGroupsRequest struct {
	Global         string `json:"global"`
	Name           string `json:"name"`
	PerPage        string `json:"per_page" default:"10"`         // Default per_page to 10
	Page           string `json:"page" default:"1"`              // Default page to 1
	OrderColumn    string `json:"order_column" default:"id"`     // Default order column to "id"
	OrderDirection string `json:"order_direction" default:"asc"` // Default order direction to "asc"
}

type CreateItemGroupRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description"`
	Remark      *string `json:"remark"`
	Status      int8    `json:"status"`
}

type UpdateItemGroupRequest struct {
	ID          uint    `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
	Remark      *string `json:"remark"`
	Status      int8    `json:"status"`
}

type GetItemGroupByIDRequest struct {
	ID uint `json:"id"`
}

type GetItemGroupParams struct {
	ID        uint
	IsDeleted *int
}

func NewGetItemGroupParams(id uint) *GetItemGroupParams {
	defaultIsDeleted := 0
	return &GetItemGroupParams{
		ID:        id,
		IsDeleted: &defaultIsDeleted,
	}
}

type DeleteItemGroupRequest struct {
	ID uint `json:"id"`
}

type ItemGroupListDTO struct {
	ID            int     `json:"id" db:"id"`
	Name          string  `json:"name" db:"name"`
	Description   *string `json:"description" db:"description"`
	Remark        *string `json:"remark" db:"remark"`
	Status        int8    `json:"status" db:"status"`
	CreatedByName *string `json:"created_by_name" db:"created_by_name"`
	UpdatedByName *string `json:"updated_by_name" db:"updated_by_name"`
	CreatedAt     *string `json:"created_at" db:"created_at"`
	UpdatedAt     *string `json:"updated_at" db:"updated_at"`
	DeleteAt      *string `json:"deleted_at" db:"deleted_at"`
}

type ItemGroupDetailDTO struct {
	ID            uint    `json:"id" db:"id"`
	Name          string  `json:"name" db:"name"`
	Description   string  `json:"description" db:"description"`
	Remark        *string `json:"remark" db:"remark"`
	Status        int8    `json:"status" db:"status"`
	CreatedByName *string `json:"created_by_name" db:"created_by_name"`
	UpdatedByName *string `json:"updated_by_name" db:"updated_by_name"`
	CreatedAt     *string `json:"created_at" db:"created_at"`
	UpdatedAt     *string `json:"updated_at" db:"updated_at"`
	DeletedAt     *string `json:"deleted_at" db:"deleted_at"`
}
type GetItemGroupsResult struct {
	ItemGroups []ItemGroupListDTO
	Total      int
	Err        error
}

type GetItemSubGroupsRequest struct {
	Global         string  `json:"global"`
	Name           string  `json:"name"`
	PerPage        *string `json:"per_page" default:"10"`         // Default per_page to 10
	Page           *string `json:"page" default:"1"`              // Default page to 1
	OrderColumn    string  `json:"order_column" default:"id"`     // Default order column to "id"
	OrderDirection string  `json:"order_direction" default:"asc"` // Default order direction to "asc"
}

type CreateItemSubGroupRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description"`
	Remark      *string `json:"remark"`
	Status      int8    `json:"status"`
	ItemGroupID uint    `json:"item_group_id"`
}

type UpdateItemSubGroupRequest struct {
	ID          uint    `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
	Remark      *string `json:"remark"`
	Status      int8    `json:"status"`
	ItemGroupID uint    `json:"item_group_id"`
}

type GetItemSubGroupByIDRequest struct {
	ID uint `json:"id"`
}

type DeleteItemSubGroupRequest struct {
	ID uint `json:"id"`
}

type GetItemSubGroupParams struct {
	ID        uint
	IsDeleted *int
}

type ItemSubGroupListDTO struct {
	ID            int     `json:"id" db:"id"`
	Name          string  `json:"name" db:"name"`
	Description   string  `json:"description" db:"description"`
	Remark        *string `json:"remark" db:"remark"`
	Status        int8    `json:"status" db:"status"`
	CreatedByName *string `json:"created_by_name" db:"created_by_name"`
	UpdatedByName *string `json:"updated_by_name" db:"updated_by_name"`
	CreatedAt     *string `json:"created_at" db:"created_at"`
	UpdatedAt     *string `json:"updated_at" db:"updated_at"`
	DeleteAt      *string `json:"deleted_at" db:"deleted_at"`
}

type ItemSubGroupDetailDTO struct {
	ID            uint    `json:"id" db:"id"`
	Name          string  `json:"name" db:"name"`
	Description   string  `json:"description" db:"description"`
	Remark        *string `json:"remark" db:"remark"`
	Status        int8    `json:"status" db:"status"`
	CreatedByID   uint    `json:"created_by_id" db:"created_by_id"`
	UpdatedByID   *uint   `json:"updated_by_id" db:"updated_by_id"`
	CreatedByName *string `json:"created_by_name" db:"created_by_name"`
	UpdatedByName *string `json:"updated_by_name" db:"updated_by_name"`
	CreatedAt     *string `json:"created_at" db:"created_at"`
	UpdatedAt     *string `json:"updated_at" db:"updated_at"`
	DeletedAt     *string `json:"deleted_at" db:"deleted_at"`
}
type GetItemSubGroupsResult struct {
	ItemSubGroups []ItemSubGroupListDTO
	Total         int
	Err           error
}

type GetCompanyProfilesRequest struct {
	ParentID       *int   `json:"parent_id"`
	Global         string `json:"global"`
	Name           string `json:"name"`
	PerPage        string `json:"per_page" default:"10"`         // Default per_page to 10
	Page           string `json:"page" default:"1"`              // Default page to 1
	OrderColumn    string `json:"order_column" default:"id"`     // Default order column to "id"
	OrderDirection string `json:"order_direction" default:"asc"` // Default order direction to "asc"
}

type CreateCompanyProfileRequest struct {
	ParentID           *uint   `json:"parent_id" db:"parent_id"`
	IsPrimary          *int    `json:"is_primary" db:"is_primary"`
	CompanyOwnerName   *string `json:"company_owner_name" db:"company_owner_name"`
	CompanySignName    *string `json:"company_sign_name" db:"company_sign_name"`
	CompanyName        string  `json:"company_name" db:"company_name"`
	CompanyAddress     *string `json:"company_address" db:"company_address"`
	CompanyPhone       *string `json:"company_phone" db:"company_phone"`
	CompanyEmail       *string `json:"company_email" db:"company_email"`
	CompanyWebsite     *string `json:"company_website" db:"company_website"`
	CompanyLogo        *string `json:"company_logo" db:"company_logo"`
	CompanySign        *string `json:"company_sign" db:"company_sign"`
	CompanyDescription *string `json:"company_description" db:"company_description"`
	CompanyRemark      *string `json:"company_remark" db:"company_remark"`
	CompanyStatus      *int    `json:"company_status" db:"company_status"`
}

type UpdateCompanyProfileRequest struct {
	ID                 uint    `json:"id"`
	ParentID           *uint   `json:"parent_id" db:"parent_id"`
	IsPrimary          *int    `json:"is_primary" db:"is_primary"`
	CompanyOwnerName   *string `json:"company_owner_name" db:"company_owner_name"`
	CompanySignName    *string `json:"company_sign_name" db:"company_sign_name"`
	CompanyName        string  `json:"company_name" db:"company_name"`
	CompanyAddress     *string `json:"company_address" db:"company_address"`
	CompanyPhone       *string `json:"company_phone" db:"company_phone"`
	CompanyEmail       *string `json:"company_email" db:"company_email"`
	CompanyWebsite     *string `json:"company_website" db:"company_website"`
	CompanyLogo        *string `json:"company_logo" db:"company_logo"`
	CompanySign        *string `json:"company_sign" db:"company_sign"`
	CompanyDescription *string `json:"company_description" db:"company_description"`
	CompanyRemark      *string `json:"company_remark" db:"company_remark"`
	CompanyStatus      *int    `json:"company_status" db:"company_status"`
}

type GetCompanyProfileByIDRequest struct {
	ID uint `json:"id"`
}

type GetCompanyProfileParams struct {
	ID        uint
	IsDeleted *int
}

func NewGetCompanyProfileParams(id uint) *GetCompanyProfileParams {
	defaultIsDeleted := 0
	return &GetCompanyProfileParams{
		ID:        id,
		IsDeleted: &defaultIsDeleted,
	}
}

type DeleteCompanyProfileRequest struct {
	ID uint `json:"id"`
}

type CompanyProfileListDTO struct {
	ID                 int     `json:"id" db:"id"`
	ParentID           *uint   `json:"parent_id" db:"parent_id"`
	IsPrimary          *int    `json:"is_primary" db:"is_primary"`
	CompanyOwnerName   *string `json:"company_owner_name" db:"company_owner_name"`
	CompanySignName    *string `json:"company_sign_name" db:"company_sign_name"`
	CompanyName        string  `json:"company_name" db:"company_name"`
	CompanyAddress     *string `json:"company_address" db:"company_address"`
	CompanyPhone       *string `json:"company_phone" db:"company_phone"`
	CompanyEmail       *string `json:"company_email" db:"company_email"`
	CompanyWebsite     *string `json:"company_website" db:"company_website"`
	CompanyLogo        *string `json:"company_logo" db:"company_logo"`
	CompanySign        *string `json:"company_sign" db:"company_sign"`
	CompanyDescription *string `json:"company_description" db:"company_description"`
	CompanyRemark      *string `json:"company_remark" db:"company_remark"`
	CompanyStatus      *int    `json:"company_status" db:"company_status"`
	CreatedByName      *string `json:"created_by_name" db:"created_by_name"`
	UpdatedByName      *string `json:"updated_by_name" db:"updated_by_name"`
	CreatedAt          *string `json:"created_at" db:"created_at"`
	UpdatedAt          *string `json:"updated_at" db:"updated_at"`
	DeleteAt           *string `json:"deleted_at" db:"deleted_at"`
}

type CompanyProfileDetailDTO struct {
	ID                 uint    `json:"id" db:"id"`
	ParentID           *uint   `json:"parent_id" db:"parent_id"`
	IsPrimary          *int    `json:"is_primary" db:"is_primary"`
	CompanyOwnerName   *string `json:"company_owner_name" db:"company_owner_name"`
	CompanySignName    *string `json:"company_sign_name" db:"company_sign_name"`
	CompanyName        *string `json:"company_name" db:"company_name"`
	CompanyAddress     *string `json:"company_address" db:"company_address"`
	CompanyPhone       *string `json:"company_phone" db:"company_phone"`
	CompanyEmail       *string `json:"company_email" db:"company_email"`
	CompanyWebsite     *string `json:"company_website" db:"company_website"`
	CompanyLogo        *string `json:"company_logo" db:"company_logo"`
	CompanySign        *string `json:"company_sign" db:"company_sign"`
	CompanyDescription *string `json:"company_description" db:"company_description"`
	CompanyRemark      *string `json:"company_remark" db:"company_remark"`
	CompanyStatus      *int    `json:"company_status" db:"company_status"`
	CreatedByName      *string `json:"created_by_name" db:"created_by_name"`
	UpdatedByName      *string `json:"updated_by_name" db:"updated_by_name"`
	CreatedAt          *string `json:"created_at" db:"created_at"`
	UpdatedAt          *string `json:"updated_at" db:"updated_at"`
	DeletedAt          *string `json:"deleted_at" db:"deleted_at"`
}
type GetCompanyProfilesResult struct {
	CompanyProfiles []CompanyProfileListDTO
	Total           int
	Err             error
}

type GetCustomerTypesRequest struct {
	Global         string `json:"global"`
	Name           string `json:"name"`
	SimulateError  string `json:"simulate_error"`
	PerPage        string `json:"per_page" default:"10"`         // Default per_page to 10
	Page           string `json:"page" default:"1"`              // Default page to 1
	OrderColumn    string `json:"order_column" default:"id"`     // Default order column to "id"
	OrderDirection string `json:"order_direction" default:"asc"` // Default order direction to "asc"
}

type CreateCustomerTypeRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description"`
	Remark      *string `json:"remark"`
	Status      int8    `json:"status"`
}

type UpdateCustomerTypeRequest struct {
	ID          uint    `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
	Remark      *string `json:"remark"`
	Status      int8    `json:"status"`
}

type GetCustomerTypeByIDRequest struct {
	ID uint `json:"id"`
}

type GetCustomerTypeParams struct {
	ID        uint
	IsDeleted *int
}

func NewGetCustomerTypeParams(id uint) *GetCustomerTypeParams {
	defaultIsDeleted := 0
	return &GetCustomerTypeParams{
		ID:        id,
		IsDeleted: &defaultIsDeleted,
	}
}

type DeleteCustomerTypeRequest struct {
	ID uint `json:"id"`
}

type CustomerTypeListDTO struct {
	ID            int     `json:"id" db:"id"`
	Name          string  `json:"name" db:"name"`
	Description   *string `json:"description" db:"description"`
	Remark        *string `json:"remark" db:"remark"`
	Status        int8    `json:"status" db:"status"`
	CreatedByName *string `json:"created_by_name" db:"created_by_name"`
	UpdatedByName *string `json:"updated_by_name" db:"updated_by_name"`
	CreatedAt     *string `json:"created_at" db:"created_at"`
	UpdatedAt     *string `json:"updated_at" db:"updated_at"`
	DeleteAt      *string `json:"deleted_at" db:"deleted_at"`
}

type CustomerTypeDetailDTO struct {
	ID            uint    `json:"id" db:"id"`
	Name          string  `json:"name" db:"name"`
	Description   string  `json:"description" db:"description"`
	Remark        *string `json:"remark" db:"remark"`
	Status        int8    `json:"status" db:"status"`
	CreatedByName *string `json:"created_by_name" db:"created_by_name"`
	UpdatedByName *string `json:"updated_by_name" db:"updated_by_name"`
	CreatedAt     *string `json:"created_at" db:"created_at"`
	UpdatedAt     *string `json:"updated_at" db:"updated_at"`
	DeletedAt     *string `json:"deleted_at" db:"deleted_at"`
}
type GetCustomerTypesResult struct {
	CustomerTypes []CustomerTypeListDTO
	Total         int
	Err           error
}

type GetMixValuesRequest struct {
	Global         string `json:"global"`
	Name           string `json:"name"`
	PerPage        string `json:"per_page" default:"10"`         // Default per_page to 10
	Page           string `json:"page" default:"1"`              // Default page to 1
	OrderColumn    string `json:"order_column" default:"id"`     // Default order column to "id"
	OrderDirection string `json:"order_direction" default:"asc"` // Default order direction to "asc"
}

type CreateMixValueRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description"`
	Remark      *string `json:"remark"`
	Status      int8    `json:"status"`
}

type UpdateMixValueRequest struct {
	ID          uint    `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
	Remark      *string `json:"remark"`
	Status      int8    `json:"status"`
}

type GetMixValueByIDRequest struct {
	ID uint `json:"id"`
}

type GetMixValueParams struct {
	ID        uint
	IsDeleted *int
}

func NewGetMixValueParams(id uint) *GetMixValueParams {
	defaultIsDeleted := 0
	return &GetMixValueParams{
		ID:        id,
		IsDeleted: &defaultIsDeleted,
	}
}

type DeleteMixValueRequest struct {
	ID uint `json:"id"`
}

type MixValueListDTO struct {
	ID            int     `json:"id" db:"id"`
	Name          string  `json:"name" db:"name"`
	Description   *string `json:"description" db:"description"`
	Remark        *string `json:"remark" db:"remark"`
	Status        int8    `json:"status" db:"status"`
	CreatedByName *string `json:"created_by_name" db:"created_by_name"`
	UpdatedByName *string `json:"updated_by_name" db:"updated_by_name"`
	CreatedAt     *string `json:"created_at" db:"created_at"`
	UpdatedAt     *string `json:"updated_at" db:"updated_at"`
	DeleteAt      *string `json:"deleted_at" db:"deleted_at"`
}

type MixValueDetailDTO struct {
	ID            uint    `json:"id" db:"id"`
	Name          string  `json:"name" db:"name"`
	Description   string  `json:"description" db:"description"`
	Remark        *string `json:"remark" db:"remark"`
	Status        int8    `json:"status" db:"status"`
	CreatedByName *string `json:"created_by_name" db:"created_by_name"`
	UpdatedByName *string `json:"updated_by_name" db:"updated_by_name"`
	CreatedAt     *string `json:"created_at" db:"created_at"`
	UpdatedAt     *string `json:"updated_at" db:"updated_at"`
	DeletedAt     *string `json:"deleted_at" db:"deleted_at"`
}
type GetMixValuesResult struct {
	MixValues []MixValueListDTO
	Total     int
	Err       error
}

type GetCurrenciesRequest struct {
	Global         string `json:"global"`
	Name           string `json:"name"`
	PerPage        string `json:"per_page" default:"10"`         // Default per_page to 10
	Page           string `json:"page" default:"1"`              // Default page to 1
	OrderColumn    string `json:"order_column" default:"id"`     // Default order column to "id"
	OrderDirection string `json:"order_direction" default:"asc"` // Default order direction to "asc"
}

type CreateCurrencyRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description"`
	Remark      *string `json:"remark"`
	Status      int8    `json:"status"`
}

type UpdateCurrencyRequest struct {
	ID          uint    `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
	Remark      *string `json:"remark"`
	Status      int8    `json:"status"`
}

type GetCurrencyByIDRequest struct {
	ID uint `json:"id"`
}

type GetCurrencyParams struct {
	ID        uint
	IsDeleted *int
}

func NewGetCurrencyParams(id uint) *GetCurrencyParams {
	defaultIsDeleted := 0
	return &GetCurrencyParams{
		ID:        id,
		IsDeleted: &defaultIsDeleted,
	}
}

type DeleteCurrencyRequest struct {
	ID uint `json:"id"`
}

type CurrencyListDTO struct {
	ID            int     `json:"id" db:"id"`
	Name          string  `json:"name" db:"name"`
	Description   *string `json:"description" db:"description"`
	Remark        *string `json:"remark" db:"remark"`
	Status        int8    `json:"status" db:"status"`
	CreatedByName *string `json:"created_by_name" db:"created_by_name"`
	UpdatedByName *string `json:"updated_by_name" db:"updated_by_name"`
	CreatedAt     *string `json:"created_at" db:"created_at"`
	UpdatedAt     *string `json:"updated_at" db:"updated_at"`
	DeleteAt      *string `json:"deleted_at" db:"deleted_at"`
}

type CurrencyDetailDTO struct {
	ID            uint    `json:"id" db:"id"`
	Name          string  `json:"name" db:"name"`
	Description   string  `json:"description" db:"description"`
	Remark        *string `json:"remark" db:"remark"`
	Status        int8    `json:"status" db:"status"`
	CreatedByName *string `json:"created_by_name" db:"created_by_name"`
	UpdatedByName *string `json:"updated_by_name" db:"updated_by_name"`
	CreatedAt     *string `json:"created_at" db:"created_at"`
	UpdatedAt     *string `json:"updated_at" db:"updated_at"`
	DeletedAt     *string `json:"deleted_at" db:"deleted_at"`
}
type GetCurrenciesResult struct {
	Currencies []CurrencyListDTO
	Total      int
	Err        error
}

type GetUnitsRequest struct {
	Global         string `json:"global"`
	Name           string `json:"name"`
	PerPage        string `json:"per_page" default:"10"`         // Default per_page to 10
	Page           string `json:"page" default:"1"`              // Default page to 1
	OrderColumn    string `json:"order_column" default:"id"`     // Default order column to "id"
	OrderDirection string `json:"order_direction" default:"asc"` // Default order direction to "asc"
}

type CreateUnitRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description"`
	Remark      *string `json:"remark"`
	Status      int8    `json:"status"`
}

type UpdateUnitRequest struct {
	ID          uint    `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
	Remark      *string `json:"remark"`
	Status      int8    `json:"status"`
}

type GetUnitByIDRequest struct {
	ID uint `json:"id"`
}

type GetUnitParams struct {
	ID        uint
	IsDeleted *int
}

func NewGetUnitParams(id uint) *GetUnitParams {
	defaultIsDeleted := 0
	return &GetUnitParams{
		ID:        id,
		IsDeleted: &defaultIsDeleted,
	}
}

type DeleteUnitRequest struct {
	ID uint `json:"id"`
}

type UnitListDTO struct {
	ID            int     `json:"id" db:"id"`
	Name          string  `json:"name" db:"name"`
	Description   *string `json:"description" db:"description"`
	Remark        *string `json:"remark" db:"remark"`
	Status        int8    `json:"status" db:"status"`
	CreatedByName *string `json:"created_by_name" db:"created_by_name"`
	UpdatedByName *string `json:"updated_by_name" db:"updated_by_name"`
	CreatedAt     *string `json:"created_at" db:"created_at"`
	UpdatedAt     *string `json:"updated_at" db:"updated_at"`
	DeleteAt      *string `json:"deleted_at" db:"deleted_at"`
}

type UnitDetailDTO struct {
	ID            uint    `json:"id" db:"id"`
	Name          string  `json:"name" db:"name"`
	Description   string  `json:"description" db:"description"`
	Remark        *string `json:"remark" db:"remark"`
	Status        int8    `json:"status" db:"status"`
	CreatedByName *string `json:"created_by_name" db:"created_by_name"`
	UpdatedByName *string `json:"updated_by_name" db:"updated_by_name"`
	CreatedAt     *string `json:"created_at" db:"created_at"`
	UpdatedAt     *string `json:"updated_at" db:"updated_at"`
	DeletedAt     *string `json:"deleted_at" db:"deleted_at"`
}
type GetUnitsResult struct {
	Units []UnitListDTO
	Total int
	Err   error
}

type GetVatsRequest struct {
	Global         string `json:"global"`
	Name           string `json:"name"`
	PerPage        string `json:"per_page" default:"10"`         // Default per_page to 10
	Page           string `json:"page" default:"1"`              // Default page to 1
	OrderColumn    string `json:"order_column" default:"id"`     // Default order column to "id"
	OrderDirection string `json:"order_direction" default:"asc"` // Default order direction to "asc"
}

type CreateVatRequest struct {
	Name        string  `json:"name"`
	Num         float64 `json:"num"`
	Description *string `json:"description"`
	Remark      *string `json:"remark"`
	Status      int8    `json:"status"`
}

type UpdateVatRequest struct {
	ID          uint    `json:"id"`
	Num         float64 `json:"num"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
	Remark      *string `json:"remark"`
	Status      int8    `json:"status"`
}

type GetVatByIDRequest struct {
	ID uint `json:"id"`
}

type GetVatParams struct {
	ID        uint
	IsDeleted *int
}

func NewGetVatParams(id uint) *GetVatParams {
	defaultIsDeleted := 0
	return &GetVatParams{
		ID:        id,
		IsDeleted: &defaultIsDeleted,
	}
}

type DeleteVatRequest struct {
	ID uint `json:"id"`
}

type VatListDTO struct {
	ID            int     `json:"id" db:"id"`
	Name          string  `json:"name" db:"name"`
	Num           float64 `json:"num" db:"num"`
	Description   *string `json:"description" db:"description"`
	Remark        *string `json:"remark" db:"remark"`
	Status        int8    `json:"status" db:"status"`
	CreatedByName *string `json:"created_by_name" db:"created_by_name"`
	UpdatedByName *string `json:"updated_by_name" db:"updated_by_name"`
	CreatedAt     *string `json:"created_at" db:"created_at"`
	UpdatedAt     *string `json:"updated_at" db:"updated_at"`
	DeleteAt      *string `json:"deleted_at" db:"deleted_at"`
}

type VatDetailDTO struct {
	ID            uint    `json:"id" db:"id"`
	Name          string  `json:"name" db:"name"`
	Description   string  `json:"description" db:"description"`
	Remark        *string `json:"remark" db:"remark"`
	Status        int8    `json:"status" db:"status"`
	CreatedByName *string `json:"created_by_name" db:"created_by_name"`
	UpdatedByName *string `json:"updated_by_name" db:"updated_by_name"`
	CreatedAt     *string `json:"created_at" db:"created_at"`
	UpdatedAt     *string `json:"updated_at" db:"updated_at"`
	DeletedAt     *string `json:"deleted_at" db:"deleted_at"`
}
type GetVatsResult struct {
	Vats  []VatListDTO
	Total int
	Err   error
}

type GetPph23sRequest struct {
	Global         string `json:"global"`
	Name           string `json:"name"`
	PerPage        string `json:"per_page" default:"10"`         // Default per_page to 10
	Page           string `json:"page" default:"1"`              // Default page to 1
	OrderColumn    string `json:"order_column" default:"id"`     // Default order column to "id"
	OrderDirection string `json:"order_direction" default:"asc"` // Default order direction to "asc"
}

type CreatePph23Request struct {
	Name        string  `json:"name"`
	Num         float64 `json:"num"`
	Description *string `json:"description"`
	Remark      *string `json:"remark"`
	Status      int8    `json:"status"`
}

type UpdatePph23Request struct {
	ID          uint    `json:"id"`
	Num         float64 `json:"num"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
	Remark      *string `json:"remark"`
	Status      int8    `json:"status"`
}

type GetPph23ByIDRequest struct {
	ID uint `json:"id"`
}

type GetPph23Params struct {
	ID        uint
	IsDeleted *int
}

func NewGetPph23Params(id uint) *GetPph23Params {
	defaultIsDeleted := 0
	return &GetPph23Params{
		ID:        id,
		IsDeleted: &defaultIsDeleted,
	}
}

type DeletePph23Request struct {
	ID uint `json:"id"`
}

type Pph23ListDTO struct {
	ID            int     `json:"id" db:"id"`
	Name          string  `json:"name" db:"name"`
	Num           float64 `json:"num" db:"num"`
	Description   *string `json:"description" db:"description"`
	Remark        *string `json:"remark" db:"remark"`
	Status        int8    `json:"status" db:"status"`
	CreatedByName *string `json:"created_by_name" db:"created_by_name"`
	UpdatedByName *string `json:"updated_by_name" db:"updated_by_name"`
	CreatedAt     *string `json:"created_at" db:"created_at"`
	UpdatedAt     *string `json:"updated_at" db:"updated_at"`
	DeleteAt      *string `json:"deleted_at" db:"deleted_at"`
}

type Pph23DetailDTO struct {
	ID            uint    `json:"id" db:"id"`
	Name          string  `json:"name" db:"name"`
	Description   string  `json:"description" db:"description"`
	Remark        *string `json:"remark" db:"remark"`
	Status        int8    `json:"status" db:"status"`
	CreatedByName *string `json:"created_by_name" db:"created_by_name"`
	UpdatedByName *string `json:"updated_by_name" db:"updated_by_name"`
	CreatedAt     *string `json:"created_at" db:"created_at"`
	UpdatedAt     *string `json:"updated_at" db:"updated_at"`
	DeletedAt     *string `json:"deleted_at" db:"deleted_at"`
}
type GetPph23sResult struct {
	Pph23s []Pph23ListDTO
	Total  int
	Err    error
}
