package dtos

type GetRolesRequest struct {
	Global         string `json:"global"`
	Name           string `json:"name"`
	PerPage        string `json:"per_page" default:"10"`         // Default per_page to 10
	Page           string `json:"page" default:"1"`              // Default page to 1
	OrderColumn    string `json:"order_column" default:"id"`     // Default order column to "id"
	OrderDirection string `json:"order_direction" default:"asc"` // Default order direction to "asc"
}

type CreateRoleRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description"`
	Remark      *string `json:"remark"`
	Status      int8    `json:"status"`
}

type UpdateRoleRequest struct {
	ID          uint    `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
	Remark      *string `json:"remark"`
	Status      int8    `json:"status"`
}

type GetRoleByIDRequest struct {
	ID uint `json:"id"`
}

type GetRoleParams struct {
	ID        uint
	IsDeleted *int
}

func NewGetRoleParams(id uint) *GetRoleParams {
	defaultIsDeleted := 0
	return &GetRoleParams{
		ID:        id,
		IsDeleted: &defaultIsDeleted,
	}
}

type DeleteRoleRequest struct {
	ID uint `json:"id"`
}

type RoleListDTO struct {
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

type RoleDetailDTO struct {
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
type GetRolesResult struct {
	Roles []RoleListDTO
	Total int
	Err   error
}
