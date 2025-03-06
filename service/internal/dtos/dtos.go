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
	ID         int     `json:"id" db:"id"`
	Username   *string `json:"username" db:"username"`
	Name       string  `json:"name" db:"name"`
	Email      string  `json:"email" db:"email"`
	Status     *int    `json:"status" db:"status"`
	Address    *string `json:"address" db:"address"`
	BranchID   *uint   `json:"branch_id" db:"branch_id"`
	BranchName *string `json:"branch_name" db:"branch_name"`
}

type UserDetailDTO struct {
	ID          uint     `json:"id"`
	BranchID    *uint    `json:"branch_id" db:"branch_id"`
	BranchName  *string  `json:"branch_name" db:"branch_name"`
	Name        string   `json:"name"`
	Username    *string  `json:"username"`
	Email       string   `json:"email"`
	Address     *string  `json:"address"`
	Status      *int     `json:"status"`
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

type GetUserParams struct {
	ID        uint
	IsDeleted *int
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
	Description   *string `json:"description" db:"description"`
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
	PerPage        *string `json:"per_page" default:"10"` // Default per_page to 10
	ParentIds      *string `json:"parent_ids"`
	ParentId       *uint   `json:"parent_id"`
	Page           *string `json:"page" default:"1"`              // Default page to 1
	OrderColumn    string  `json:"order_column" default:"id"`     // Default order column to "id"
	OrderDirection string  `json:"order_direction" default:"asc"` // Default order direction to "asc"
}

type CreateItemSubGroupRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description"`
	Remark      *string `json:"remark"`
	Status      int8    `json:"status"`
	ItemGroupID uint    `json:"item_group_id" gorm:"column:parent_id"`
}

type UpdateItemSubGroupRequest struct {
	ID          uint    `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
	Remark      *string `json:"remark"`
	Status      int8    `json:"status"`
	ItemGroupID uint    `json:"item_group_id" gorm:"column:parent_id"`
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
	ParentID      *uint   `json:"parent_id" db:"parent_id"`
	Name          string  `json:"name" db:"name"`
	GroupName     string  `json:"group_name" db:"group_name"`
	Description   *string `json:"description" db:"description"`
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
	ParentID      *uint   `json:"parent_id" db:"parent_id"`
	ItemGroupID   *uint   `json:"item_group_id" db:"item_group_id"`
	Name          string  `json:"name" db:"name"`
	GroupName     string  `json:"group_name" db:"group_name"`
	Description   *string `json:"description" db:"description"`
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
	IsPrimary *int
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

type GetBranchesRequest struct {
	ParentID       *int   `json:"parent_id"`
	Global         string `json:"global"`
	Name           string `json:"name"`
	PerPage        string `json:"per_page" default:"10"`         // Default per_page to 10
	Page           string `json:"page" default:"1"`              // Default page to 1
	OrderColumn    string `json:"order_column" default:"id"`     // Default order column to "id"
	OrderDirection string `json:"order_direction" default:"asc"` // Default order direction to "asc"
}

type CreateBranchRequest struct {
	ParentID         *uint   `json:"parent_id" db:"parent_id"`
	CompanyProfileID *uint   `json:"company_profile_id" db:"company_profile_id"`
	OwnerName        *string `json:"owner_name" db:"owner_name"`
	SignName         *string `json:"sign_name" db:"sign_name"`
	Name             string  `json:"name" db:"name"`
	Address          *string `json:"address" db:"address"`
	Phone            *string `json:"phone" db:"phone"`
	Email            *string `json:"email" db:"email"`
	Website          *string `json:"website" db:"website"`
	Logo             *string `json:"logo" db:"logo"`
	Sign             *string `json:"sign" db:"sign"`
	Description      *string `json:"description" db:"description"`
	Remark           *string `json:"remark" db:"remark"`
	Status           *int    `json:"status" db:"status"`
}

type UpdateBranchRequest struct {
	ID               uint    `json:"id"`
	ParentID         *uint   `json:"parent_id" db:"parent_id"`
	CompanyProfileID *uint   `json:"company_profile_id" db:"company_profile_id"`
	OwnerName        *string `json:"owner_name" db:"owner_name"`
	SignName         *string `json:"sign_name" db:"sign_name"`
	Name             string  `json:"name" db:"name"`
	Address          *string `json:"address" db:"address"`
	Phone            *string `json:"phone" db:"phone"`
	Email            *string `json:"email" db:"email"`
	Website          *string `json:"website" db:"website"`
	Logo             *string `json:"logo" db:"logo"`
	Sign             *string `json:"sign" db:"sign"`
	Description      *string `json:"description" db:"description"`
	Remark           *string `json:"remark" db:"remark"`
	Status           *int    `json:"status" db:"status"`
}

type GetBranchByIDRequest struct {
	ID uint `json:"id"`
}

type GetBranchParams struct {
	ID               uint
	CompanyProfileID *uint
	IsDeleted        *int
}

func NewGetBranchParams(id uint) *GetBranchParams {
	defaultIsDeleted := 0
	return &GetBranchParams{
		ID:        id,
		IsDeleted: &defaultIsDeleted,
	}
}

type DeleteBranchRequest struct {
	ID uint `json:"id"`
}

type BranchListDTO struct {
	ID               int     `json:"id" db:"id"`
	ParentID         *uint   `json:"parent_id" db:"parent_id"`
	CompanyProfileID *uint   `json:"company_profile_id" db:"company_profile_id"`
	CompanyName      *string `json:"company_name" db:"company_name"`
	OwnerName        *string `json:"owner_name" db:"owner_name"`
	SignName         *string `json:"sign_name" db:"sign_name"`
	Name             string  `json:"name" db:"name"`
	Address          *string `json:"address" db:"address"`
	Phone            *string `json:"phone" db:"phone"`
	Email            *string `json:"email" db:"email"`
	Website          *string `json:"website" db:"website"`
	Logo             *string `json:"logo" db:"logo"`
	Sign             *string `json:"sign" db:"sign"`
	Description      *string `json:"description" db:"description"`
	Remark           *string `json:"remark" db:"remark"`
	Status           *int    `json:"status" db:"status"`
	CreatedByName    *string `json:"created_by_name" db:"created_by_name"`
	UpdatedByName    *string `json:"updated_by_name" db:"updated_by_name"`
	CreatedAt        *string `json:"created_at" db:"created_at"`
	UpdatedAt        *string `json:"updated_at" db:"updated_at"`
	DeleteAt         *string `json:"deleted_at" db:"deleted_at"`
}

type BranchDetailDTO struct {
	ID               uint    `json:"id" db:"id"`
	ParentID         *uint   `json:"parent_id" db:"parent_id"`
	CompanyProfileID *uint   `json:"company_profile_id" db:"company_profile_id"`
	CompanyName      *string `json:"company_name" db:"company_name"`
	OwnerName        *string `json:"owner_name" db:"owner_name"`
	SignName         *string `json:"sign_name" db:"sign_name"`
	Name             *string `json:"name" db:"name"`
	Address          *string `json:"address" db:"address"`
	Phone            *string `json:"phone" db:"phone"`
	Email            *string `json:"email" db:"email"`
	Website          *string `json:"website" db:"website"`
	Logo             *string `json:"logo" db:"logo"`
	Sign             *string `json:"sign" db:"sign"`
	Description      *string `json:"description" db:"description"`
	Remark           *string `json:"remark" db:"remark"`
	Status           *int    `json:"status" db:"status"`
	CreatedByName    *string `json:"created_by_name" db:"created_by_name"`
	UpdatedByName    *string `json:"updated_by_name" db:"updated_by_name"`
	CreatedAt        *string `json:"created_at" db:"created_at"`
	UpdatedAt        *string `json:"updated_at" db:"updated_at"`
	DeletedAt        *string `json:"deleted_at" db:"deleted_at"`
}
type GetBranchesResult struct {
	Branches []BranchListDTO
	Total    int
	Err      error
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
	Description   *string `json:"description" db:"description"`
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

type GetOrderTypesRequest struct {
	Global         string `json:"global"`
	Name           string `json:"name"`
	SimulateError  string `json:"simulate_error"`
	PerPage        string `json:"per_page" default:"10"`         // Default per_page to 10
	Page           string `json:"page" default:"1"`              // Default page to 1
	OrderColumn    string `json:"order_column" default:"id"`     // Default order column to "id"
	OrderDirection string `json:"order_direction" default:"asc"` // Default order direction to "asc"
}

type CreateOrderTypeRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description"`
	Remark      *string `json:"remark"`
	Status      int8    `json:"status"`
}

type UpdateOrderTypeRequest struct {
	ID          uint    `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
	Remark      *string `json:"remark"`
	Status      int8    `json:"status"`
}

type GetOrderTypeByIDRequest struct {
	ID uint `json:"id"`
}

type GetOrderTypeParams struct {
	ID        uint
	IsDeleted *int
}

func NewGetOrderTypeParams(id uint) *GetOrderTypeParams {
	defaultIsDeleted := 0
	return &GetOrderTypeParams{
		ID:        id,
		IsDeleted: &defaultIsDeleted,
	}
}

type DeleteOrderTypeRequest struct {
	ID uint `json:"id"`
}

type OrderTypeListDTO struct {
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

type OrderTypeDetailDTO struct {
	ID            uint    `json:"id" db:"id"`
	Name          string  `json:"name" db:"name"`
	Description   *string `json:"description" db:"description"`
	Remark        *string `json:"remark" db:"remark"`
	Status        int8    `json:"status" db:"status"`
	CreatedByName *string `json:"created_by_name" db:"created_by_name"`
	UpdatedByName *string `json:"updated_by_name" db:"updated_by_name"`
	CreatedAt     *string `json:"created_at" db:"created_at"`
	UpdatedAt     *string `json:"updated_at" db:"updated_at"`
	DeletedAt     *string `json:"deleted_at" db:"deleted_at"`
}
type GetOrderTypesResult struct {
	OrderTypes []OrderTypeListDTO
	Total      int
	Err        error
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
	Description   *string `json:"description" db:"description"`
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
	Num         float64 `json:"num"`
	Description *string `json:"description"`
	Remark      *string `json:"remark"`
	Status      int8    `json:"status"`
}

type UpdateCurrencyRequest struct {
	ID          uint    `json:"id"`
	Name        string  `json:"name"`
	Num         float64 `json:"num"`
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
	Num           string  `json:"num" db:"num"`
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
	Num           string  `json:"num" db:"num"`
	Description   *string `json:"description" db:"description"`
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
	Description   *string `json:"description" db:"description"`
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
	Name        string   `json:"name"`
	Num         float64  `json:"num"`
	Description *string  `json:"description"`
	Remark      *string  `json:"remark"`
	Status      int8     `json:"status"`
	Divider     *float64 `json:"divider"`
	ChangedAt   *string  `json:"changed_at"`
	Multiplier  *float64 `json:"multiplier"`
}

type UpdateVatRequest struct {
	ID          uint     `json:"id"`
	Num         float64  `json:"num"`
	Name        string   `json:"name"`
	Description *string  `json:"description"`
	Remark      *string  `json:"remark"`
	Status      int8     `json:"status"`
	Divider     *float64 `json:"divider"`
	Multiplier  *float64 `json:"multiplier"`
	ChangedAt   *string  `json:"changed_at"`
}
type UpdateVatHistoryRequest struct {
	ID         uint     `json:"id"`
	VatID      uint     `json:"vat_id"`
	Num        float64  `json:"num"`
	Remark     *string  `json:"remark"`
	Status     int8     `json:"status"`
	Divider    *float64 `json:"divider"`
	Multiplier *float64 `json:"multiplier"`
	ChangedAt  *string  `json:"changed_at"`
}

type GetVatByIDRequest struct {
	ID uint `json:"id"`
}

type GetVatParams struct {
	ID        uint
	IsDeleted *int
}

type GetVatHistoryParams struct {
	ID        *uint
	VatID     *uint
	IsLatest  *int
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
	Multiplier    *string `json:"multiplier" db:"multiplier"`
	Divider       *string `json:"divider" db:"divider"`
	CreatedByName *string `json:"created_by_name" db:"created_by_name"`
	UpdatedByName *string `json:"updated_by_name" db:"updated_by_name"`
	CreatedAt     *string `json:"created_at" db:"created_at"`
	UpdatedAt     *string `json:"updated_at" db:"updated_at"`
	DeleteAt      *string `json:"deleted_at" db:"deleted_at"`
}

type VatDetailDTO struct {
	ID            uint    `json:"id" db:"id"`
	Name          string  `json:"name" db:"name"`
	Num           string  `json:"num" db:"num"`
	Description   *string `json:"description" db:"description"`
	Remark        *string `json:"remark" db:"remark"`
	Status        int8    `json:"status" db:"status"`
	Multiplier    *string `json:"multiplier" db:"multiplier"`
	Divider       *string `json:"divider" db:"divider"`
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

type VatHistoryListDTO struct {
	ID            int     `json:"id" db:"id"`
	VatID         int     `json:"vat_id" db:"vat_id"`
	Num           float64 `json:"num" db:"num"`
	Divider       float64 `json:"divider" db:"divider"`
	Multiplier    float64 `json:"multiplier" db:"multiplier"`
	ChangedAt     *string `json:"changed_at" db:"changed_at"`
	Status        int8    `json:"status" db:"status"`
	Remark        *string `json:"remark" db:"remark"`
	Name          string  `json:"name" db:"name"`
	CreatedByName *string `json:"created_by_name" db:"created_by_name"`
	UpdatedByName *string `json:"updated_by_name" db:"updated_by_name"`
	CreatedAt     *string `json:"created_at" db:"created_at"`
	UpdatedAt     *string `json:"updated_at" db:"updated_at"`
	DeleteAt      *string `json:"deleted_at" db:"deleted_at"`
}

type VatHistoryDetailDTO struct {
	ID            int     `json:"id" db:"id"`
	VatID         int     `json:"vat_id" db:"vat_id"`
	Num           float64 `json:"num" db:"num"`
	Divider       float64 `json:"divider" db:"divider"`
	Multiplier    float64 `json:"multiplier" db:"multiplier"`
	ChangedAt     *string `json:"changed_at" db:"changed_at"`
	Status        int8    `json:"status" db:"status"`
	Remark        *string `json:"remark" db:"remark"`
	Name          string  `json:"name" db:"name"`
	CreatedByName *string `json:"created_by_name" db:"created_by_name"`
	UpdatedByName *string `json:"updated_by_name" db:"updated_by_name"`
	CreatedAt     *string `json:"created_at" db:"created_at"`
	UpdatedAt     *string `json:"updated_at" db:"updated_at"`
	DeleteAt      *string `json:"deleted_at" db:"deleted_at"`
}

type GetVatHistoriesResult struct {
	VatHistories []VatHistoryListDTO
	Total        int
	Err          error
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
	Description   *string `json:"description" db:"description"`
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

type GetMsItemsRequest struct {
	Global         string `json:"global"`
	Name           string `json:"name"`
	ItemGroupID    string `json:"item_group_id"`
	ItemSubGroupID string `json:"item_sub_group_id"`
	Status         string `json:"status"`
	PerPage        string `json:"per_page" default:"10"`         // Default per_page to 10
	Page           string `json:"page" default:"1"`              // Default page to 1
	OrderColumn    string `json:"order_column" default:"id"`     // Default order column to "id"
	OrderDirection string `json:"order_direction" default:"asc"` // Default order direction to "asc"
}

type CreateMsItemUnitsRequest struct {
	UnitID     uint    `json:"unit_id"`
	Conversion float64 `json:"conversion"`
	PriceSell  float64 `json:"price_sell"`
	PriceBuy   float64 `json:"price_buy"`
}

type CreateMsItemRequest struct {
	ItemSubGroupID uint    `json:"item_sub_group_id"`
	ItemUnitID     uint    `json:"item_unit_id"`
	Code           *string `json:"code"`
	Name           string  `json:"name"`
	Specification  *string `json:"specification"`
	Description    *string `json:"description"`
	TpbCode        *string `json:"tpb_code"`
	// PriceSell      *float64 `json:"price_sell"`
	// PriceBuy       *float64 `json:"price_buy"`
	MinimumStock *float64 `json:"minimum_stock"`
	IsAllBranch  *int     `json:"is_all_branch"`
	Status       int8     `json:"status"`
	Units        []CreateMsItemUnitsRequest
}

type UpdateMsItemRequest struct {
	ID             uint     `json:"id"`
	ItemSubGroupID uint     `json:"item_sub_group_id"`
	ItemUnitID     uint     `json:"item_unit_id"`
	Code           *string  `json:"code"`
	Name           string   `json:"name"`
	Specification  *string  `json:"specification"`
	Description    *string  `json:"description"`
	TpbCode        *string  `json:"tpb_code"`
	PriceSell      *float64 `json:"price_sell"`
	PriceBuy       *float64 `json:"price_buy"`
	MinimumStock   *float64 `json:"minimum_stock"`
	IsAllBranch    *int     `json:"is_all_branch"`
	Status         int8     `json:"status"`
}

type GetMsItemByIDRequest struct {
	ID uint `json:"id"`
}

type GetMsItemParams struct {
	ID        uint
	IsDeleted *int
}

func NewGetMsItemParams(id uint) *GetMsItemParams {
	defaultIsDeleted := 0
	return &GetMsItemParams{
		ID:        id,
		IsDeleted: &defaultIsDeleted,
	}
}

type GetMsItemItemUnitParams struct {
	ID        uint
	MsItemID  uint
	UnitID    uint
	IsDeleted *int
}

type GetProductItemUnitParams struct {
	ID        uint
	ProductID uint
	UnitID    uint
	IsDeleted *int
}

type DeleteMsItemRequest struct {
	ID uint `json:"id"`
}

// bi.description as branch_item_description, bi.tpb_code as branch_item_tpb_code, bi.price_sell as branch_item_price_sell, bi.price_buy as branch_item_price_buy, bi.minimum_stock as branch_item_minimum_stock, bi.status as branch_item_status, bi.created_at as branch_item_created_at, bi.updated_at as branch_item_updated_at, bi.deleted_at as branch_item_deleted_at,
type MsItemListDTO struct {
	ID               int     `json:"id" db:"id"`
	ItemSubGroupID   *uint   `json:"item_sub_group_id" db:"item_sub_group_id"`
	ItemGroupID      *uint   `json:"item_group_id" db:"item_group_id"`
	ItemUnitID       *uint   `json:"item_unit_id" db:"item_unit_id"`
	ItemUnitUnitID   *uint   `json:"item_unit_unit_id" db:"item_unit_unit_id"`
	BranchID         *uint   `json:"branch_id" db:"branch_id"`
	BranchItemID     *uint   `json:"branch_item_id" db:"branch_item_id"`
	ItemSubGroupName *string `json:"item_sub_group_name" db:"item_sub_group_name"`
	ItemGroupName    *string `json:"item_group_name" db:"item_group_name"`
	UnitName         *string `json:"unit_name" db:"unit_name"`
	Name             string  `json:"name" db:"name"`
	Code             *string `json:"code" db:"code"`
	Specification    *string `json:"specification" db:"specification"`
	Description      *string `json:"description" db:"description"`
	TpbCode          *string `json:"tpb_code" db:"tpb_code"`
	PriceSell        *string `json:"price_sell" db:"price_sell"`
	PriceBuy         *string `json:"price_buy" db:"price_buy"`
	MinimumStock     *string `json:"minimum_stock" db:"minimum_stock"`
	IsAllBranch      *int    `json:"is_all_branch" db:"is_all_branch"`
	Status           int8    `json:"status" db:"status"`
	CreatedByName    *string `json:"created_by_name" db:"created_by_name"`
	UpdatedByName    *string `json:"updated_by_name" db:"updated_by_name"`
	CreatedAt        *string `json:"created_at" db:"created_at"`
	UpdatedAt        *string `json:"updated_at" db:"updated_at"`
	DeleteAt         *string `json:"deleted_at" db:"deleted_at"`
}

type MsItemDetailDTO struct {
	ID               uint    `json:"id" db:"id"`
	ItemSubGroupID   *uint   `json:"item_sub_group_id" db:"item_sub_group_id"`
	ItemGroupID      *uint   `json:"item_group_id" db:"item_group_id"`
	ItemUnitID       *uint   `json:"item_unit_id" db:"item_unit_id"`
	ItemSubGroupName *string `json:"item_sub_group_name" db:"item_sub_group_name"`
	ItemGroupName    *string `json:"item_group_name" db:"item_group_name"`
	UnitName         *string `json:"unit_name" db:"unit_name"`
	Name             string  `json:"name" db:"name"`
	Code             *string `json:"code" db:"code"`
	Specification    *string `json:"specification" db:"specification"`
	Description      *string `json:"description" db:"description"`
	TpbCode          *string `json:"tpb_code" db:"tpb_code"`
	PriceSell        *string `json:"price_sell" db:"price_sell"`
	PriceBuy         *string `json:"price_buy" db:"price_buy"`
	MinimumStock     *string `json:"minimum_stock" db:"minimum_stock"`
	IsAllBranch      *int    `json:"is_all_branch" db:"is_all_branch"`
	Status           int8    `json:"status" db:"status"`
	CreatedByName    *string `json:"created_by_name" db:"created_by_name"`
	UpdatedByName    *string `json:"updated_by_name" db:"updated_by_name"`
	CreatedAt        *string `json:"created_at" db:"created_at"`
	UpdatedAt        *string `json:"updated_at" db:"updated_at"`
	DeletedAt        *string `json:"deleted_at" db:"deleted_at"`
}
type GetMsItemsResult struct {
	MsItems []MsItemListDTO
	Total   int
	Err     error
}

type GetItemUnitsRequest struct {
	Global         string `json:"global"`
	Name           string `json:"name"`
	UnitID         string `json:"unit_id"`
	ProductID      string `json:"product_id"`
	Status         string `json:"status"`
	PerPage        string `json:"per_page" default:"10"`         // Default per_page to 10
	Page           string `json:"page" default:"1"`              // Default page to 1
	OrderColumn    string `json:"order_column" default:"id"`     // Default order column to "id"
	OrderDirection string `json:"order_direction" default:"asc"` // Default order direction to "asc"
}

type CreateItemUnitRequest struct {
	ProductID  uint     `json:"product_id"`
	UnitID     uint     `json:"unit_id"`
	Conversion *float64 `json:"conversion"`
	PriceSell  *float64 `json:"price_sell"`
	PriceBuy   *float64 `json:"price_buy"`
	Status     int8     `json:"status"`
}

type UpdateItemUnitRequest struct {
	ID         uint     `json:"id"`
	ProductID  uint     `json:"product_id"`
	UnitID     uint     `json:"unit_id"`
	Conversion *float64 `json:"conversion"`
	PriceSell  *float64 `json:"price_sell"`
	PriceBuy   *float64 `json:"price_buy"`
	Status     int8     `json:"status"`
}

type GetItemUnitByIDRequest struct {
	ID uint `json:"id"`
}

type GetItemUnitParams struct {
	ID        uint
	IsDeleted *int
}

func NewGetItemUnitParams(id uint) *GetItemUnitParams {
	defaultIsDeleted := 0
	return &GetItemUnitParams{
		ID:        id,
		IsDeleted: &defaultIsDeleted,
	}
}

type DeleteItemUnitRequest struct {
	ID uint `json:"id"`
}

type ItemUnitListDTO struct {
	ID            int      `json:"id" db:"id"`
	ProductID     uint     `json:"product_id" db:"product_id"`
	UnitID        *uint    `json:"unit_id" db:"unit_id"`
	ProductName   *string  `json:"product_name" db:"product_name"`
	UnitName      *string  `json:"unit_name" db:"unit_name"`
	PriceSell     *float64 `json:"price_sell" db:"price_sell"`
	PriceBuy      *float64 `json:"price_buy" db:"price_buy"`
	Conversion    *float64 `json:"conversion" db:"conversion"`
	Status        int8     `json:"status" db:"status"`
	CreatedByName *string  `json:"created_by_name" db:"created_by_name"`
	UpdatedByName *string  `json:"updated_by_name" db:"updated_by_name"`
	CreatedAt     *string  `json:"created_at" db:"created_at"`
	UpdatedAt     *string  `json:"updated_at" db:"updated_at"`
	DeleteAt      *string  `json:"deleted_at" db:"deleted_at"`
}

type ItemUnitDetailDTO struct {
	ID            uint     `json:"id" db:"id"`
	ProductID     uint     `json:"product_id" db:"product_id"`
	UnitID        *uint    `json:"unit_id" db:"unit_id"`
	ProductName   *string  `json:"product_name" db:"product_name"`
	UnitName      *string  `json:"unit_name" db:"unit_name"`
	PriceSell     *float64 `json:"price_sell" db:"price_sell"`
	PriceBuy      *float64 `json:"price_buy" db:"price_buy"`
	Conversion    *float64 `json:"conversion" db:"conversion"`
	Status        int8     `json:"status" db:"status"`
	CreatedByName *string  `json:"created_by_name" db:"created_by_name"`
	UpdatedByName *string  `json:"updated_by_name" db:"updated_by_name"`
	CreatedAt     *string  `json:"created_at" db:"created_at"`
	UpdatedAt     *string  `json:"updated_at" db:"updated_at"`
	DeletedAt     *string  `json:"deleted_at" db:"deleted_at"`
}
type GetItemUnitsResult struct {
	ItemUnits []ItemUnitListDTO
	Total     int
	Err       error
}

type GetBranchItemsRequest struct {
	Global         string `json:"global"`
	Name           string `json:"name"`
	BranchID       uint   `json:"branch_id"`
	UnitID         uint   `json:"unit_id"`
	ProductID      uint   `json:"product_id"`
	Status         string `json:"status"`
	PerPage        string `json:"per_page" default:"10"`         // Default per_page to 10
	Page           string `json:"page" default:"1"`              // Default page to 1
	OrderColumn    string `json:"order_column" default:"id"`     // Default order column to "id"
	OrderDirection string `json:"order_direction" default:"asc"` // Default order direction to "asc"
}

type CreateBranchItemRequest struct {
	BranchID      uint     `json:"branch_id"`
	ItemUnitID    uint     `json:"item_unit_id"`
	Code          *string  `json:"code"`
	FactoryCode   *string  `json:"factory_code"`
	Name          string   `json:"name"`
	Sku           *string  `json:"sku"`
	Barcode       *string  `json:"barcode"`
	Specification *string  `json:"specification"`
	Description   *string  `json:"description"`
	Remark        *string  `json:"remark"`
	TpbCode       *string  `json:"tpb_code"`
	MinimumStock  *float64 `json:"minimum_stock"`
	PriceSell     *float64 `json:"price_sell"`
	PriceBuy      *float64 `json:"price_buy"`
	Status        int8     `json:"status"`
	ExpiredAt     *string  `json:"expired_at"`
}

type UpdateBranchItemRequest struct {
	ID            uint     `json:"id"`
	BranchID      uint     `json:"branch_id"`
	ItemUnitID    uint     `json:"item_unit_id"`
	Code          *string  `json:"code"`
	FactoryCode   *string  `json:"factory_code"`
	Name          string   `json:"name"`
	Sku           *string  `json:"sku"`
	Barcode       *string  `json:"barcode"`
	Specification *string  `json:"specification"`
	Description   *string  `json:"description"`
	Remark        *string  `json:"remark"`
	TpbCode       *string  `json:"tpb_code"`
	MinimumStock  *float64 `json:"minimum_stock"`
	PriceSell     *float64 `json:"price_sell"`
	PriceBuy      *float64 `json:"price_buy"`
	Status        int8     `json:"status"`
	ExpiredAt     *string  `json:"expired_at"`
}

type GetBranchItemByIDRequest struct {
	ID uint `json:"id"`
}

type GetBranchItemParams struct {
	ID        uint
	IsDeleted *int
}

func NewGetBranchItemParams(id uint) *GetBranchItemParams {
	defaultIsDeleted := 0
	return &GetBranchItemParams{
		ID:        id,
		IsDeleted: &defaultIsDeleted,
	}
}

type DeleteBranchItemRequest struct {
	ID uint `json:"id"`
}

type BranchItemListDTO struct {
	ID            int      `json:"id" db:"id"`
	ProductID     uint     `json:"product_id" db:"product_id"`
	BranchID      uint     `json:"branch_id" db:"branch_id"`
	UnitID        uint     `json:"unit_id" db:"unit_id"`
	ProductName   *string  `json:"product_name" db:"product_name"`
	BranchName    *string  `json:"branch_name" db:"branch_name"`
	UnitName      *string  `json:"unit_name" db:"unit_name"`
	Name          *string  `json:"name" db:"name"`
	Specification *string  `json:"specification" gorm:"column:specification"`
	Description   *string  `json:"description" gorm:"column:description"`
	TpbCode       *string  `json:"tpb_code" gorm:"column:tpb_code"`
	MinimumStock  *float64 `json:"minimum_stock" gorm:"column:minimum_stock"`
	PriceSell     *float64 `json:"price_sell" db:"price_sell"`
	PriceBuy      *float64 `json:"price_buy" db:"price_buy"`
	Status        int8     `json:"status" db:"status"`
	CreatedByName *string  `json:"created_by_name" db:"created_by_name"`
	UpdatedByName *string  `json:"updated_by_name" db:"updated_by_name"`
	CreatedAt     *string  `json:"created_at" db:"created_at"`
	UpdatedAt     *string  `json:"updated_at" db:"updated_at"`
	DeleteAt      *string  `json:"deleted_at" db:"deleted_at"`
}

type BranchItemDetailDTO struct {
	ID            uint     `json:"id" db:"id"`
	ProductID     uint     `json:"product_id" db:"product_id"`
	BranchID      uint     `json:"branch_id" db:"branch_id"`
	UnitID        uint     `json:"unit_id" db:"unit_id"`
	ProductName   *string  `json:"ms_item_name" db:"ms_item_name"`
	BranchName    *string  `json:"branch_name" db:"branch_name"`
	UnitName      *string  `json:"unit_name" db:"unit_name"`
	Name          *string  `json:"name" db:"name"`
	Specification *string  `json:"specification" gorm:"column:specification"`
	Description   *string  `json:"description" gorm:"column:description"`
	TpbCode       *string  `json:"tpb_code" gorm:"column:tpb_code"`
	MinimumStock  *float64 `json:"minimum_stock" gorm:"column:minimum_stock"`
	PriceSell     *float64 `json:"price_sell" db:"price_sell"`
	PriceBuy      *float64 `json:"price_buy" db:"price_buy"`
	Status        int8     `json:"status" db:"status"`
	CreatedByName *string  `json:"created_by_name" db:"created_by_name"`
	UpdatedByName *string  `json:"updated_by_name" db:"updated_by_name"`
	CreatedAt     *string  `json:"created_at" db:"created_at"`
	UpdatedAt     *string  `json:"updated_at" db:"updated_at"`
	DeletedAt     *string  `json:"deleted_at" db:"deleted_at"`
}
type GetBranchItemsResult struct {
	BranchItems []BranchItemListDTO
	Total       int
	Err         error
}

type GetCustomersRequest struct {
	Global         string `json:"global"`
	Name           string `json:"name"`
	CustomerTypeID string `json:"customer_type_id"`
	PerPage        string `json:"per_page" default:"10"`         // Default per_page to 10
	Page           string `json:"page" default:"1"`              // Default page to 1
	OrderColumn    string `json:"order_column" default:"id"`     // Default order column to "id"
	OrderDirection string `json:"order_direction" default:"asc"` // Default order direction to "asc"
}

type CreateCustomerRequest struct {
	CustomerTypeID *uint   `json:"customer_type_id"`
	AgentID        *uint   `json:"agent_id"`
	Code           *string `json:"code"`
	Name           string  `json:"name"`
	Address        *string `json:"address"`
	Phone          *string `json:"phone"`
	Email          *string `json:"email"`
	Pic            *string `json:"pic"`
	Status         int8    `json:"status"`
}

type UpdateCustomerRequest struct {
	ID             uint    `json:"id"`
	CustomerTypeID *uint   `json:"customer_type_id"`
	AgentID        *uint   `json:"agent_id"`
	Code           *string `json:"code"`
	Name           string  `json:"name"`
	Address        *string `json:"address"`
	Phone          *string `json:"phone"`
	Email          *string `json:"email"`
	Pic            *string `json:"pic"`
	Status         int8    `json:"status"`
}

type GetCustomerByIDRequest struct {
	ID uint `json:"id"`
}

type GetCustomerParams struct {
	ID        uint
	IsDeleted *int
}

func NewGetCustomerParams(id uint) *GetCustomerParams {
	defaultIsDeleted := 0
	return &GetCustomerParams{
		ID:        id,
		IsDeleted: &defaultIsDeleted,
	}
}

type DeleteCustomerRequest struct {
	ID uint `json:"id"`
}

type CustomerListDTO struct {
	ID               int     `json:"id" db:"id"`
	CustomerTypeID   *uint   `json:"customer_type_id" db:"customer_type_id"`
	CustomerTypeName *string `json:"customer_type_name" db:"customer_type_name"`
	AgentID          *uint   `json:"agent_id" db:"agent_id"`
	AgentName        *string `json:"agent_name" db:"agent_name"`
	Code             *string `json:"code" db:"code"`
	Name             string  `json:"name" db:"name"`
	Address          *string `json:"address" db:"address"`
	Phone            *string `json:"phone" db:"phone"`
	Email            *string `json:"email" db:"email"`
	Pic              *string `json:"pic" db:"pic"`
	Status           int8    `json:"status" db:"status"`
	CreatedByName    *string `json:"created_by_name" db:"created_by_name"`
	UpdatedByName    *string `json:"updated_by_name" db:"updated_by_name"`
	CreatedAt        *string `json:"created_at" db:"created_at"`
	UpdatedAt        *string `json:"updated_at" db:"updated_at"`
	DeleteAt         *string `json:"deleted_at" db:"deleted_at"`
}

type CustomerDetailDTO struct {
	ID               uint    `json:"id" db:"id"`
	CustomerTypeID   *uint   `json:"customer_type_id" db:"customer_type_id"`
	CustomerTypeName *string `json:"customer_type_name" db:"customer_type_name"`
	AgentID          *uint   `json:"agent_id" db:"agent_id"`
	AgentName        *string `json:"agent_name" db:"agent_name"`
	Code             *string `json:"code" db:"code"`
	Name             string  `json:"name" db:"name"`
	Address          *string `json:"address" db:"address"`
	Phone            *string `json:"phone" db:"phone"`
	Email            *string `json:"email" db:"email"`
	Pic              *string `json:"pic" db:"pic"`
	Status           int8    `json:"status" db:"status"`
	CreatedByName    *string `json:"created_by_name" db:"created_by_name"`
	UpdatedByName    *string `json:"updated_by_name" db:"updated_by_name"`
	CreatedAt        *string `json:"created_at" db:"created_at"`
	UpdatedAt        *string `json:"updated_at" db:"updated_at"`
	DeletedAt        *string `json:"deleted_at" db:"deleted_at"`
}
type GetCustomersResult struct {
	Customers []CustomerListDTO
	Total     int
	Err       error
}

type GetProductsRequest struct {
	Global         string `json:"global"`
	Name           string `json:"name"`
	Sku            string `json:"sku"`
	Barcode        string `json:"barcode"`
	ExpiredAt      string `json:"expired_at"`
	Remark         string `json:"remark"`
	FactoryCode    string `json:"factory_code"`
	ItemSubGroupID string `json:"item_sub_group_id"`
	UnitID         string `json:"unit_id"`
	Status         string `json:"status"`
	PerPage        string `json:"per_page" default:"10"`         // Default per_page to 10
	Page           string `json:"page" default:"1"`              // Default page to 1
	OrderColumn    string `json:"order_column" default:"id"`     // Default order column to "id"
	OrderDirection string `json:"order_direction" default:"asc"` // Default order direction to "asc"
}

type CreateBomsRequest struct {
	ProductItemID uint    `json:"product_item_id"`
	ItemUnitID    uint    `json:"item_unit_id"`
	Qty           float64 `json:"qty"`
	Remark        *string `json:"remark"`
}

type CreateProductRequest struct {
	ItemSubGroupID uint                       `json:"item_sub_group_id"`
	ItemUnitID     uint                       `json:"item_unit_id"`
	Code           *string                    `json:"code"`
	FactoryCode    *string                    `json:"factory_code"`
	Name           string                     `json:"name"`
	Sku            *string                    `json:"sku"`
	Barcode        *string                    `json:"barcode"`
	Specification  *string                    `json:"specification"`
	Description    *string                    `json:"description"`
	Remark         *string                    `json:"remark"`
	PriceSell      *float64                   `json:"price_sell"`
	PriceBuy       *float64                   `json:"price_buy"`
	Margin         *float64                   `json:"margin"`
	TpbCode        *string                    `json:"tpb_code"`
	MinimumStock   *float64                   `json:"minimum_stock"`
	IsAllBranch    *int                       `json:"is_all_branch"`
	Status         int8                       `json:"status"`
	ExpiredAt      *string                    `json:"expired_at"`
	Units          []CreateMsItemUnitsRequest `json:"units"`
	Boms           []CreateBomsRequest        `json:"boms"`
}

type UpdateBomsRequest struct {
	ID            *uint   `json:"id"`
	ProductItemID uint    `json:"product_item_id"`
	ItemUnitID    uint    `json:"item_unit_id"`
	Qty           float64 `json:"qty"`
	Remark        *string `json:"remark"`
}

type UpdateProductRequest struct {
	ID             uint                `json:"id"`
	ItemSubGroupID uint                `json:"item_sub_group_id"`
	ItemUnitID     uint                `json:"item_unit_id"`
	Code           *string             `json:"code"`
	FactoryCode    *string             `json:"factory_code"`
	Name           string              `json:"name"`
	Sku            *string             `json:"sku"`
	Barcode        *string             `json:"barcode"`
	Specification  *string             `json:"specification"`
	Description    *string             `json:"description"`
	Remark         *string             `json:"remark"`
	PriceSell      *float64            `json:"price_sell"`
	PriceBuy       *float64            `json:"price_buy"`
	Margin         *float64            `json:"margin"`
	TpbCode        *string             `json:"tpb_code"`
	MinimumStock   *float64            `json:"minimum_stock"`
	IsAllBranch    *int                `json:"is_all_branch"`
	Status         int8                `json:"status"`
	ExpiredAt      *string             `json:"expired_at"`
	Boms           []UpdateBomsRequest `json:"boms"`
}

type GetProductByIDRequest struct {
	ID uint `json:"id"`
}

type GetProductParams struct {
	ID        uint
	IsDeleted *int
}

func NewGetProductParams(id uint) *GetProductParams {
	defaultIsDeleted := 0
	return &GetProductParams{
		ID:        id,
		IsDeleted: &defaultIsDeleted,
	}
}

type GetProductBomParams struct {
	ID        uint
	ProductID uint
	IsDeleted *int
}

type DeleteProductRequest struct {
	ID uint `json:"id"`
}

type ProductListDTO struct {
	ID               int     `json:"id" db:"id"`
	ProductID        *uint   `json:"product_id" db:"product_id"`
	ItemSubGroupID   uint    `json:"item_sub_group_id" db:"item_sub_group_id"`
	ItemGroupID      *uint   `json:"item_group_id" db:"item_group_id"`
	ItemUnitID       *uint   `json:"item_unit_id" db:"item_unit_id"`
	ItemUnitUnitID   *uint   `json:"item_unit_unit_id" db:"item_unit_unit_id"`
	BranchID         *uint   `json:"branch_id" db:"branch_id"`
	BranchItemID     *uint   `json:"branch_item_id" db:"branch_item_id"`
	ItemSubGroupName *string `json:"item_sub_group_name" db:"item_sub_group_name"`
	ItemGroupName    *string `json:"item_group_name" db:"item_group_name"`
	UnitName         *string `json:"unit_name" db:"unit_name"`
	BranchName       *string `json:"branch_name" db:"branch_name"`
	Code             *string `json:"code" db:"code"`
	FactoryCode      *string `json:"factory_code" db:"factory_code"`
	Name             string  `json:"name" db:"name"`
	Sku              *string `json:"sku" db:"sku"`
	Barcode          *string `json:"barcode" db:"barcode"`
	Specification    *string `json:"specification" db:"specification"`
	Description      *string `json:"description" db:"description"`
	TpbCode          *string `json:"tpb_code" db:"tpb_code"`
	MinimumStock     *string `json:"minimum_stock" db:"minimum_stock"`
	IsAllBranch      *int    `json:"is_all_branch" db:"is_all_branch"`
	Remark           *string `json:"remark" db:"remark"`
	PriceSell        *string `json:"price_sell" db:"price_sell"`
	PriceBuy         *string `json:"price_buy" db:"price_buy"`
	Margin           *string `json:"margin" db:"margin"`
	ExpiredAt        *string `json:"expired_at" db:"expired_at"`
	Status           int8    `json:"status" db:"status"`
	CreatedByName    *string `json:"created_by_name" db:"created_by_name"`
	UpdatedByName    *string `json:"updated_by_name" db:"updated_by_name"`
	CreatedAt        *string `json:"created_at" db:"created_at"`
	UpdatedAt        *string `json:"updated_at" db:"updated_at"`
	DeleteAt         *string `json:"deleted_at" db:"deleted_at"`
}

type ProductBomListDTO struct {
	ID               *uint   `json:"id" db:"id"`
	BomID            *uint   `json:"bom_id" db:"bom_id"`
	ProductID        uint    `json:"product_id" db:"product_id"`
	ProductItemID    uint    `json:"product_item_id" db:"product_item_id"`
	ItemSubGroupID   *uint   `json:"item_sub_group_id" db:"item_sub_group_id"`
	ItemGroupID      *uint   `json:"item_group_id" db:"item_group_id"`
	ItemUnitID       *uint   `json:"item_unit_id" db:"item_unit_id"`
	ItemSubGroupName *string `json:"item_sub_group_name" db:"item_sub_group_name"`
	ItemGroupName    *string `json:"item_group_name" db:"item_group_name"`
	Qty              float64 `json:"qty" db:"qty"`
	Remark           *string `json:"remark" db:"remark"`
	ProductItemName  *string `json:"product_item_name" db:"product_item_name"`
	ItemUnitName     *string `json:"item_unit_name" db:"item_unit_name"`
	CreatedByName    *string `json:"created_by_name" db:"created_by_name"`
	UpdatedByName    *string `json:"updated_by_name" db:"updated_by_name"`
	CreatedAt        *string `json:"created_at" db:"created_at"`
	UpdatedAt        *string `json:"updated_at" db:"updated_at"`
	DeleteAt         *string `json:"deleted_at" db:"deleted_at"`
}

type ProductDetailDTO struct {
	ID               uint                `json:"id" db:"id"`
	ProductID        *uint               `json:"product_id" db:"product_id"`
	ItemSubGroupID   uint                `json:"item_sub_group_id" db:"item_sub_group_id"`
	ItemGroupID      *uint               `json:"item_group_id" db:"item_group_id"`
	ItemUnitID       uint                `json:"item_unit_id" db:"item_unit_id"`
	ItemUnitUnitID   *uint               `json:"item_unit_unit_id" db:"item_unit_unit_id"`
	BranchID         *uint               `json:"branch_id" db:"branch_id"`
	BranchItemID     *uint               `json:"branch_item_id" db:"branch_item_id"`
	ItemSubGroupName *string             `json:"item_sub_group_name" db:"item_sub_group_name"`
	ItemGroupName    *string             `json:"item_group_name" db:"item_group_name"`
	UnitName         *string             `json:"unit_name" db:"unit_name"`
	BranchName       *string             `json:"branch_name" db:"branch_name"`
	Code             *string             `json:"code" db:"code"`
	FactoryCode      *string             `json:"factory_code" db:"factory_code"`
	Name             string              `json:"name" db:"name"`
	Sku              *string             `json:"sku" db:"sku"`
	Barcode          *string             `json:"barcode" db:"barcode"`
	Specification    *string             `json:"specification" db:"specification"`
	Description      *string             `json:"description" db:"description"`
	TpbCode          *string             `json:"tpb_code" db:"tpb_code"`
	MinimumStock     *string             `json:"minimum_stock" db:"minimum_stock"`
	IsAllBranch      *int                `json:"is_all_branch" db:"is_all_branch"`
	Remark           *string             `json:"remark" db:"remark"`
	PriceSell        *string             `json:"price_sell" db:"price_sell"`
	PriceBuy         *string             `json:"price_buy" db:"price_buy"`
	Margin           *string             `json:"margin" db:"margin"`
	ExpiredAt        *string             `json:"expired_at" db:"expired_at"`
	Status           int8                `json:"status" db:"status"`
	CreatedByName    *string             `json:"created_by_name" db:"created_by_name"`
	UpdatedByName    *string             `json:"updated_by_name" db:"updated_by_name"`
	CreatedAt        *string             `json:"created_at" db:"created_at"`
	UpdatedAt        *string             `json:"updated_at" db:"updated_at"`
	DeleteAt         *string             `json:"deleted_at" db:"deleted_at"`
	Boms             []ProductBomListDTO `json:"boms"`
}

type GetProductsResult struct {
	Products []ProductListDTO
	Total    int
	Err      error
}
