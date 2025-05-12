package dtos

type GetTicketsRequest struct {
	Global         *string `json:"global"`
	Title          *string `json:"title"`
	PoBuyerNo      *string `json:"po_buyer_no"`
	TicketNo       *string `json:"ticket_no"`
	Remark         *string `json:"remark"`
	CustomerID     *int    `json:"customer_id"`
	OrderTypeID    *int    `json:"order_type_id"`
	CurrencyID     *int    `json:"currency_id"`
	VatID          *int    `json:"vat_id"`
	PaymentID      *int    `json:"payment_id"`
	Pph23ID        *int    `json:"pph23_id"`
	BranchID       *int    `json:"branch_id"`
	Status         *string `json:"status"`
	DateType       *string `json:"date_type"` // 1 = due_at, 2 = expired_at
	StartDate      *string `json:"start_date"`
	EndDate        *string `json:"end_date"`
	PerPage        *string `json:"per_page" default:"10"`         // Default per_page to 10
	Page           *string `json:"page" default:"1"`              // Default page to 1
	OrderColumn    *string `json:"order_column" default:"id"`     // Default order column to "id"
	OrderDirection *string `json:"order_direction" default:"asc"` // Default order direction to "asc"
}

type FormTicketRequest struct {
	ID                   *uint                            `json:"id"`
	CustomerID           *uint                            `json:"customer_id"`
	ProductID            *uint                            `json:"product_id"`
	BranchID             *uint                            `json:"branch_id"`
	PriorityType         *string                          `json:"priority_type"`
	RevNo                *int                             `json:"rev_no"`
	Title                string                           `json:"title" db:"title"`
	TicketNo             *string                          `json:"ticket_no" db:"ticket_no"`
	TicketNoOri          *string                          `json:"ticket_no_ori" db:"ticket_no_ori"`
	IssueDesc            *string                          `json:"issue_desc" db:"issue_desc"`
	IssueSolution        *string                          `json:"issue_solution" db:"issue_solution"`
	ReportedAt           *string                          `json:"reported_at" db:"reported_at"`
	Remark               *string                          `json:"remark" db:"remark"`
	Status               string                           `json:"status" db:"status"`
	Schedule             *UpdateScheduleRequest           `json:"schedule"`
	IssueAttachments     []UpdateSalesOrderAttachmentsDTO `json:"issue_attachments"`
	SolutionAttachments  []UpdateSalesOrderAttachmentsDTO `json:"solution_attachments"`
	DeletedIssueFiles    []uint                           `json:"deleted_issue_files"`
	DeletedSolutionFiles []uint                           `json:"deleted_solution_files"`

	CustomerCode string `json:"customer_code"`
}

type UpdateTicketAttachmentsDTO struct {
	ID       *uint   `json:"id" db:"id"`
	RefID    *uint   `json:"ref_id" db:"ref_id"`
	RefType  *string `json:"ref_type" db:"ref_type"`
	FileType *string `json:"file_type" db:"file_type"`
	FileUrl  *string `json:"file_url" db:"file_url"`
	FileName *string `json:"file_name" db:"file_name"`
	Remark   *string `json:"remark" db:"remark"`
}

type GetTicketByIDRequest struct {
	ID interface{} `json:"id"`
}

type GetTicketParams struct {
	ID        uint
	IsDeleted *int
}

func NewGetTicketParams(id uint) *GetTicketParams {
	defaultIsDeleted := 0
	return &GetTicketParams{
		ID:        id,
		IsDeleted: &defaultIsDeleted,
	}
}

type GetTicketSoDtParams struct {
	ID        uint
	TicketID  uint
	IsDeleted *int
}

type DeleteTicketRequest struct {
	ID uint `json:"id"`
}

type TicketListDTO struct {
	ID            int     `json:"id" db:"id"`
	TicketID      *uint   `json:"ticket_id" db:"ticket_id"`
	CustomerID    *uint   `json:"customer_id" db:"customer_id"`
	BranchID      *uint   `json:"branch_id" db:"branch_id"`
	PriorityType  *string `json:"priority_type" db:"priority_type"`
	IsSchedule    *string `json:"is_schedule" db:"is_schedule"`
	Title         *string `json:"title" db:"title"`
	TicketNo      *string `json:"ticket_no" db:"ticket_no"`
	TicketNoOri   *string `json:"ticket_no_ori" db:"ticket_no_ori"`
	Remark        *string `json:"remark" db:"remark"`
	Status        string  `json:"status" db:"status"`
	ReportedAt    *string `json:"reported_at" db:"reported_at"`
	CreatedByID   *uint   `json:"crweated_by_id" db:"created_by_id"`
	UpdatedByID   *uint   `json:"updated_by_id" db:"updated_by_id"`
	DeletedByID   *uint   `json:"deleted_by_id" db:"deleted_by_id"`
	CreatedByName *string `json:"created_by_name" db:"created_by_name"`
	UpdatedByName *string `json:"updated_by_name" db:"updated_by_name"`
	CreatedAt     *string `json:"created_at" db:"created_at"`
	UpdatedAt     *string `json:"updated_at" db:"updated_at"`
	DeleteAt      *string `json:"deleted_at" db:"deleted_at"`

	// so_dt_vat_id, currency_name vat_name pph23_name
	ProductID    *string `json:"product_id" db:"product_id"`
	CustomerName *string `json:"customer_name" db:"customer_name"`
	ProductName  *string `json:"product_name" db:"product_name"`
}

type TicketDetailDTO struct {
	ID                  uint                       `json:"id" db:"id"`
	TicketID            *uint                      `json:"ticket_id" db:"ticket_id"`
	ProductIDID         *uint                      `json:"product_id" db:"product_id"`
	CustomerID          *uint                      `json:"customer_id" db:"customer_id"`
	BranchID            *uint                      `json:"branch_id" db:"branch_id"`
	PriorityType        *string                    `json:"priority_type" db:"priority_type"`
	IsSchedule          *string                    `json:"is_schedule" db:"is_schedule"`
	Title               *string                    `json:"title" db:"title"`
	TicketNo            *string                    `json:"ticket_no" db:"ticket_no"`
	TicketNoOri         *string                    `json:"ticket_no_ori" db:"ticket_no_ori"`
	IssueDesc           *string                    `json:"issue_desc" db:"issue_desc"`
	IssueSolution       *string                    `json:"issue_solution" db:"issue_solution"`
	Remark              *string                    `json:"remark" db:"remark"`
	Status              string                     `json:"status" db:"status"`
	ReportedAt          *string                    `json:"reported_at" db:"reported_at"`
	CreatedByID         *uint                      `json:"crweated_by_id" db:"created_by_id"`
	UpdatedByID         *uint                      `json:"updated_by_id" db:"updated_by_id"`
	DeletedByID         *uint                      `json:"deleted_by_id" db:"deleted_by_id"`
	CreatedByName       *string                    `json:"created_by_name" db:"created_by_name"`
	UpdatedByName       *string                    `json:"updated_by_name" db:"updated_by_name"`
	CreatedAt           *string                    `json:"created_at" db:"created_at"`
	UpdatedAt           *string                    `json:"updated_at" db:"updated_at"`
	DeleteAt            *string                    `json:"deleted_at" db:"deleted_at"`
	Schedule            *ScheduleDetailDTO         `json:"schedule"`
	IssueAttachments    []SalesOrderAttachmentsDTO `json:"issue_attachments"`
	SolutionAttachments []SalesOrderAttachmentsDTO `json:"solution_attachments"`
}

type TicketAttachmentsDTO struct {
	ID         *uint   `json:"id" db:"id"`
	RefID      *uint   `json:"ref_id" db:"ref_id"`
	RefType    *string `json:"ref_type" db:"ref_type"`
	FileType   *string `json:"file_type" db:"file_type"`
	FileUrl    *string `json:"file_url" db:"file_url"`
	FileName   *string `json:"file_name" db:"file_name"`
	Remark     *string `json:"remark" db:"remark"`
	FileSize   *int64  `json:"file_size" db:"file_size"`
	DeviceType *string `json:"device_type" db:"device_type"`
	CreatedAt  *string `json:"created_at" db:"created_at"`
	DeletedAt  *string `json:"deleted_at" db:"deleted_at"`

	CreatedByName *string `json:"created_by_name" db:"created_by_name"`
	UpdatedByName *string `json:"updated_by_name" db:"updated_by_name"`
}

type TicketStatusWidget struct {
	Status      string `json:"status" db:"status"`
	TicketCount int    `json:"ticket_count" db:"ticket_count"`
	WidgetType  string `json:"widget_type" db:"widget_type"`
}
