package rest

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"github.com/jmoiron/sqlx"
	"github.com/nibroos/s-erp-api/service/internal/dtos"
	"github.com/nibroos/s-erp-api/service/internal/models"
	"github.com/nibroos/s-erp-api/service/internal/utils"
	"github.com/nibroos/s-erp-api/service/internal/validators"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Test doubles
// ---------------------------------------------------------------------------

// mockContactService is a testify stub for ContactServiceProvider, letting each
// test drive the controller down a specific branch without a database.
type mockContactService struct {
	mock.Mock
}

var _ ContactServiceProvider = (*mockContactService)(nil)

func (m *mockContactService) ListContacts(ctx *fiber.Ctx, filters map[string]string) ([]dtos.ContactListDTO, int, error) {
	args := m.Called(ctx, filters)
	var list []dtos.ContactListDTO
	if v := args.Get(0); v != nil {
		list = v.([]dtos.ContactListDTO)
	}
	return list, args.Int(1), args.Error(2)
}

func (m *mockContactService) CreateContact(ctx *fiber.Ctx, contact *models.Contact) (*models.Contact, error) {
	args := m.Called(ctx, contact)
	var out *models.Contact
	if v := args.Get(0); v != nil {
		out = v.(*models.Contact)
	}
	return out, args.Error(1)
}

func (m *mockContactService) GetContactByID(ctx *fiber.Ctx, params *dtos.GetContactParams) (*dtos.ContactDetailDTO, error) {
	args := m.Called(ctx, params)
	var out *dtos.ContactDetailDTO
	if v := args.Get(0); v != nil {
		out = v.(*dtos.ContactDetailDTO)
	}
	return out, args.Error(1)
}

func (m *mockContactService) UpdateContact(ctx *fiber.Ctx, contact *models.Contact) (*models.Contact, error) {
	args := m.Called(ctx, contact)
	var out *models.Contact
	if v := args.Get(0); v != nil {
		out = v.(*models.Contact)
	}
	return out, args.Error(1)
}

func (m *mockContactService) DeleteContact(ctx *fiber.Ctx, id uint) error {
	return m.Called(ctx, id).Error(0)
}

func (m *mockContactService) RestoreContact(ctx *fiber.Ctx, id uint) error {
	return m.Called(ctx, id).Error(0)
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

const (
	testJWTSecret = "unit-test-secret"
	testUserID    = uint(42)
)

// apiResp mirrors utils.Response with Data kept raw so each test decodes it itself.
type apiResp struct {
	Data    json.RawMessage `json:"data"`
	Meta    *utils.Meta     `json:"meta"`
	Message string          `json:"message"`
	Status  int16           `json:"status"`
	Errors  interface{}     `json:"errors"`
}

// route wires one handler onto a throwaway Fiber app. withFilters mimics the
// filters middleware that populates ctx.Locals("filters") in production.
func newTestApp(handler fiber.Handler, withFilters bool) *fiber.App {
	app := fiber.New()
	if withFilters {
		app.Use(func(c *fiber.Ctx) error {
			c.Locals("filters", map[string]string{"page": "1", "per_page": "10"})
			return c.Next()
		})
	}
	app.Post("/", handler)
	return app
}

// call fires a request at the app and decodes the JSON envelope.
func call(t *testing.T, app *fiber.App, body, token string) (*http.Response, apiResp) {
	t.Helper()
	var r io.Reader
	if body != "" {
		r = strings.NewReader(body)
	}
	req := httptest.NewRequest(http.MethodPost, "/", r)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	raw, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	var out apiResp
	_ = json.Unmarshal(raw, &out)
	return resp, out
}

// authToken mints a JWT that middleware.GetAuthUser will accept.
func authToken(t *testing.T) string {
	t.Helper()
	t.Setenv("JWT_SECRET", testJWTSecret)
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": float64(testUserID),
		"exp":     time.Now().Add(time.Hour).Unix(),
	})
	signed, err := tok.SignedString([]byte(testJWTSecret))
	require.NoError(t, err)
	return signed
}

// stubValidatorDB points the package-level validator DB at sqlmock, so the
// `exists:` / `unique_ig:` rules resolve without Postgres. Expectations are
// matched out of order because Go map iteration decides which rule runs first.
func stubValidatorDB(t *testing.T) sqlmock.Sqlmock {
	t.Helper()
	conn, m, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close() })
	m.MatchExpectationsInOrder(false)
	validators.InitValidator(sqlx.NewDb(conn, "postgres"))
	return m
}

func rowCount(n int) *sqlmock.Rows {
	return sqlmock.NewRows([]string{"count"}).AddRow(n)
}

// existsPasses makes `exists:mix_values,id` and `exists:users,id` succeed.
func existsPasses(m sqlmock.Sqlmock) {
	m.ExpectQuery("FROM mix_values").WillReturnRows(rowCount(1))
	m.ExpectQuery("FROM users").WillReturnRows(rowCount(1))
}

func sampleDetail(id uint) *dtos.ContactDetailDTO {
	return &dtos.ContactDetailDTO{ID: id, UserID: testUserID, RefNum: "REF-1", Status: 1}
}

var errBoom = errors.New("boom")

// ---------------------------------------------------------------------------
// ListContacts
// ---------------------------------------------------------------------------

func TestListContacts_MissingFilters_400(t *testing.T) {
	svc := new(mockContactService)
	resp, body := call(t, newTestApp(NewContactController(svc).ListContacts, false), "", "")

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	assert.Equal(t, "Invalid filters", body.Message)
	svc.AssertNotCalled(t, "ListContacts", mock.Anything, mock.Anything)
}

func TestListContacts_ServiceError_500(t *testing.T) {
	svc := new(mockContactService)
	svc.On("ListContacts", mock.Anything, mock.Anything).Return(nil, 0, errBoom)

	resp, body := call(t, newTestApp(NewContactController(svc).ListContacts, true), "", "")

	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	assert.Equal(t, errBoom.Error(), body.Message)
	svc.AssertExpectations(t)
}

func TestListContacts_Success_200(t *testing.T) {
	svc := new(mockContactService)
	want := []dtos.ContactListDTO{{ID: 1, RefNum: "REF-1"}, {ID: 2, RefNum: "REF-2"}}
	svc.On("ListContacts", mock.Anything, mock.Anything).Return(want, 2, nil)

	resp, body := call(t, newTestApp(NewContactController(svc).ListContacts, true), "", "")

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "Contacts fetched successfully", body.Message)

	var got []dtos.ContactListDTO
	require.NoError(t, json.Unmarshal(body.Data, &got))
	assert.Equal(t, want, got)

	require.NotNil(t, body.Meta)
	assert.Equal(t, 2, body.Meta.Total)
	assert.Equal(t, 10, body.Meta.PerPage)
	svc.AssertExpectations(t)
}

// ---------------------------------------------------------------------------
// CreateContact
// ---------------------------------------------------------------------------

func TestCreateContact_InvalidBody_400(t *testing.T) {
	svc := new(mockContactService)
	resp, body := call(t, newTestApp(NewContactController(svc).CreateContact, true), `{"bad"`, "")

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	assert.Equal(t, "Invalid request", body.Message)
	svc.AssertNotCalled(t, "CreateContact", mock.Anything, mock.Anything)
}

func TestCreateContact_ValidationFails_400(t *testing.T) {
	// Empty body: required rules fail. `exists` short-circuits on nil values,
	// so no DB is touched on this path.
	stubValidatorDB(t)
	svc := new(mockContactService)

	resp, body := call(t, newTestApp(NewContactController(svc).CreateContact, true), `{}`, "")

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	assert.Equal(t, "Validation failed", body.Message)
	svc.AssertNotCalled(t, "CreateContact", mock.Anything, mock.Anything)
}

func TestCreateContact_ServiceError_500(t *testing.T) {
	existsPasses(stubValidatorDB(t))
	svc := new(mockContactService)
	svc.On("CreateContact", mock.Anything, mock.Anything).Return(nil, errBoom)

	resp, body := call(t, newTestApp(NewContactController(svc).CreateContact, true),
		`{"type_contact_id":1,"user_id":42,"ref_num":"REF-1","status":1}`, "")

	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	assert.Equal(t, "Failed to create contact", body.Message)
	svc.AssertExpectations(t)
}

func TestCreateContact_ReloadNotFound_404(t *testing.T) {
	existsPasses(stubValidatorDB(t))
	svc := new(mockContactService)
	svc.On("CreateContact", mock.Anything, mock.Anything).Return(&models.Contact{ID: 7}, nil)
	svc.On("GetContactByID", mock.Anything, mock.Anything).Return(nil, errBoom)

	resp, body := call(t, newTestApp(NewContactController(svc).CreateContact, true),
		`{"type_contact_id":1,"user_id":42,"ref_num":"REF-1","status":1}`, "")

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	assert.Equal(t, "Contact not found", body.Message)
	svc.AssertExpectations(t)
}

func TestCreateContact_Success_201(t *testing.T) {
	existsPasses(stubValidatorDB(t))
	svc := new(mockContactService)
	svc.On("CreateContact", mock.Anything, mock.MatchedBy(func(c *models.Contact) bool {
		// The handler must map the request onto the model before persisting.
		return c.TypeContactID == 1 && c.UserID == 42 && c.RefNum == "REF-1" && c.CreatedAt != nil
	})).Return(&models.Contact{ID: 7}, nil)
	svc.On("GetContactByID", mock.Anything, &dtos.GetContactParams{ID: 7}).Return(sampleDetail(7), nil)

	resp, body := call(t, newTestApp(NewContactController(svc).CreateContact, true),
		`{"type_contact_id":1,"user_id":42,"ref_num":"REF-1","status":1}`, "")

	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	assert.Equal(t, "Contact created successfully", body.Message)

	var got []dtos.ContactDetailDTO
	require.NoError(t, json.Unmarshal(body.Data, &got))
	require.Len(t, got, 1)
	assert.Equal(t, uint(7), got[0].ID)
	svc.AssertExpectations(t)
}

// ---------------------------------------------------------------------------
// GetContactByID
// ---------------------------------------------------------------------------

func TestGetContactByID_InvalidBody_400(t *testing.T) {
	svc := new(mockContactService)
	resp, _ := call(t, newTestApp(NewContactController(svc).GetContactByID, true), `{"id"`, "")
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestGetContactByID_ZeroID_400(t *testing.T) {
	svc := new(mockContactService)
	resp, body := call(t, newTestApp(NewContactController(svc).GetContactByID, true), `{"id":0}`, "")

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	assert.Equal(t, "Contact not found", body.Message)
	svc.AssertNotCalled(t, "GetContactByID", mock.Anything, mock.Anything)
}

func TestGetContactByID_NotFound_404(t *testing.T) {
	svc := new(mockContactService)
	svc.On("GetContactByID", mock.Anything, &dtos.GetContactParams{ID: 9}).Return(nil, errBoom)

	resp, body := call(t, newTestApp(NewContactController(svc).GetContactByID, true), `{"id":9}`, "")

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	assert.Equal(t, "Contact not found", body.Message)
	svc.AssertExpectations(t)
}

func TestGetContactByID_Success_200(t *testing.T) {
	svc := new(mockContactService)
	svc.On("GetContactByID", mock.Anything, &dtos.GetContactParams{ID: 9}).Return(sampleDetail(9), nil)

	resp, body := call(t, newTestApp(NewContactController(svc).GetContactByID, true), `{"id":9}`, "")

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "Contact fetched successfully", body.Message)

	var got []dtos.ContactDetailDTO
	require.NoError(t, json.Unmarshal(body.Data, &got))
	require.Len(t, got, 1)
	assert.Equal(t, uint(9), got[0].ID)
	svc.AssertExpectations(t)
}

// ---------------------------------------------------------------------------
// UpdateContact
// ---------------------------------------------------------------------------

// updateBody is valid for ContactUpdateRequest (user_id + ref_num required).
const updateBody = `{"id":5,"type_contact_id":3,"user_id":42,"ref_num":"REF-9","status":1}`

// updateValidationPasses satisfies exists:mix_values, exists:users and
// unique_ig:contacts (which passes only when the count is 0).
func updateValidationPasses(m sqlmock.Sqlmock) {
	existsPasses(m)
	m.ExpectQuery("FROM contacts").WillReturnRows(rowCount(0))
}

func TestUpdateContact_InvalidBody_400(t *testing.T) {
	svc := new(mockContactService)
	resp, body := call(t, newTestApp(NewContactController(svc).UpdateContact, true), `{"id"`, "")

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	assert.Equal(t, "Invalid request", body.Message)
}

func TestUpdateContact_ValidationFails_400(t *testing.T) {
	stubValidatorDB(t)
	svc := new(mockContactService)

	resp, body := call(t, newTestApp(NewContactController(svc).UpdateContact, true), `{"id":5}`, "")

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	assert.Equal(t, "Validation failed", body.Message)
	svc.AssertNotCalled(t, "UpdateContact", mock.Anything, mock.Anything)
}

func TestUpdateContact_ExistingNotFound_404(t *testing.T) {
	updateValidationPasses(stubValidatorDB(t))
	svc := new(mockContactService)
	svc.On("GetContactByID", mock.Anything, &dtos.GetContactParams{ID: 5}).Return(nil, errBoom)

	resp, body := call(t, newTestApp(NewContactController(svc).UpdateContact, true), updateBody, "")

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	assert.Equal(t, "Contact not found", body.Message)
	svc.AssertNotCalled(t, "UpdateContact", mock.Anything, mock.Anything)
}

func TestUpdateContact_ServiceError_500(t *testing.T) {
	updateValidationPasses(stubValidatorDB(t))
	svc := new(mockContactService)
	svc.On("GetContactByID", mock.Anything, &dtos.GetContactParams{ID: 5}).Return(sampleDetail(5), nil)
	svc.On("UpdateContact", mock.Anything, mock.Anything).Return(nil, errBoom)

	resp, body := call(t, newTestApp(NewContactController(svc).UpdateContact, true), updateBody, "")

	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	assert.Equal(t, "Failed to update contact", body.Message)
	svc.AssertExpectations(t)
}

func TestUpdateContact_Success_200(t *testing.T) {
	updateValidationPasses(stubValidatorDB(t))
	svc := new(mockContactService)
	svc.On("GetContactByID", mock.Anything, &dtos.GetContactParams{ID: 5}).Return(sampleDetail(5), nil).Once()
	svc.On("UpdateContact", mock.Anything, mock.MatchedBy(func(c *models.Contact) bool {
		// type_contact_id was supplied, so it must override the existing value.
		return c.ID == 5 && c.TypeContactID == 3 && c.RefNum == "REF-9"
	})).Return(&models.Contact{ID: 5}, nil)
	svc.On("GetContactByID", mock.Anything, &dtos.GetContactParams{ID: 5}).Return(sampleDetail(5), nil).Once()

	resp, body := call(t, newTestApp(NewContactController(svc).UpdateContact, true), updateBody, "")

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "Contact updated successfully", body.Message)
	svc.AssertExpectations(t)
}

// ---------------------------------------------------------------------------
// DeleteContact / RestoreContact
// ---------------------------------------------------------------------------

func TestDeleteContact_ZeroID_400(t *testing.T) {
	svc := new(mockContactService)
	resp, body := call(t, newTestApp(NewContactController(svc).DeleteContact, true), `{"id":0}`, "")

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	assert.Equal(t, "Contact not found", body.Message)
	svc.AssertNotCalled(t, "DeleteContact", mock.Anything, mock.Anything)
}

func TestDeleteContact_NotFound_404(t *testing.T) {
	svc := new(mockContactService)
	svc.On("GetContactByID", mock.Anything, &dtos.GetContactParams{ID: 3}).Return(nil, errBoom)

	resp, _ := call(t, newTestApp(NewContactController(svc).DeleteContact, true), `{"id":3}`, "")

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	svc.AssertNotCalled(t, "DeleteContact", mock.Anything, mock.Anything)
}

func TestDeleteContact_ServiceError_500(t *testing.T) {
	svc := new(mockContactService)
	svc.On("GetContactByID", mock.Anything, &dtos.GetContactParams{ID: 3}).Return(sampleDetail(3), nil)
	svc.On("DeleteContact", mock.Anything, uint(3)).Return(errBoom)

	resp, body := call(t, newTestApp(NewContactController(svc).DeleteContact, true), `{"id":3}`, "")

	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	assert.Equal(t, "Failed to delete contact", body.Message)
	svc.AssertExpectations(t)
}

func TestDeleteContact_Success_200(t *testing.T) {
	svc := new(mockContactService)
	svc.On("GetContactByID", mock.Anything, &dtos.GetContactParams{ID: 3}).Return(sampleDetail(3), nil)
	svc.On("DeleteContact", mock.Anything, uint(3)).Return(nil)

	resp, body := call(t, newTestApp(NewContactController(svc).DeleteContact, true), `{"id":3}`, "")

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "Contact deleted successfully", body.Message)
	svc.AssertExpectations(t)
}

func TestRestoreContact_ZeroID_400(t *testing.T) {
	svc := new(mockContactService)
	resp, _ := call(t, newTestApp(NewContactController(svc).RestoreContact, true), `{"id":0}`, "")
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestRestoreContact_Success_200(t *testing.T) {
	svc := new(mockContactService)
	// Restore must look the contact up among soft-deleted rows (IsDeleted = 1).
	svc.On("GetContactByID", mock.Anything, mock.MatchedBy(func(p *dtos.GetContactParams) bool {
		return p.ID == 4 && p.IsDeleted != nil && *p.IsDeleted == 1
	})).Return(sampleDetail(4), nil)
	svc.On("RestoreContact", mock.Anything, uint(4)).Return(nil)

	resp, body := call(t, newTestApp(NewContactController(svc).RestoreContact, true), `{"id":4}`, "")

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "Contact restored successfully", body.Message)
	svc.AssertExpectations(t)
}

func TestRestoreContact_ServiceError_500(t *testing.T) {
	svc := new(mockContactService)
	svc.On("GetContactByID", mock.Anything, mock.Anything).Return(sampleDetail(4), nil)
	svc.On("RestoreContact", mock.Anything, uint(4)).Return(errBoom)

	resp, body := call(t, newTestApp(NewContactController(svc).RestoreContact, true), `{"id":4}`, "")

	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	assert.Equal(t, "Failed to restore contact", body.Message)
	svc.AssertExpectations(t)
}

// ---------------------------------------------------------------------------
// *ByAuthUser handlers
// ---------------------------------------------------------------------------

func TestListContactsByAuthUser_NoToken_401(t *testing.T) {
	svc := new(mockContactService)
	resp, body := call(t, newTestApp(NewContactController(svc).ListContactsByAuthUser, true), "", "")

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	assert.Equal(t, "Unauthorized", body.Message)
}

func TestListContactsByAuthUser_ScopesFilterToCaller_200(t *testing.T) {
	token := authToken(t)
	svc := new(mockContactService)
	// The caller's id must be forced into the filters so users only see their own.
	svc.On("ListContacts", mock.Anything, mock.MatchedBy(func(f map[string]string) bool {
		return f["user_id"] == "42"
	})).Return([]dtos.ContactListDTO{{ID: 1}}, 1, nil)

	resp, body := call(t, newTestApp(NewContactController(svc).ListContactsByAuthUser, true), "", token)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "Contacts fetched successfully", body.Message)
	svc.AssertExpectations(t)
}

func TestCreateContactByAuthUser_NoToken_401(t *testing.T) {
	svc := new(mockContactService)
	resp, body := call(t, newTestApp(NewContactController(svc).CreateContactByAuthUser, true),
		`{"type_contact_id":1,"ref_num":"REF-1","status":1}`, "")

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	assert.Equal(t, "Unauthorized", body.Message)
	svc.AssertNotCalled(t, "CreateContact", mock.Anything, mock.Anything)
}

func TestCreateContactByAuthUser_UsesTokenUserID_201(t *testing.T) {
	token := authToken(t)
	existsPasses(stubValidatorDB(t))
	svc := new(mockContactService)
	// user_id in the body is ignored; the JWT's user_id wins.
	svc.On("CreateContact", mock.Anything, mock.MatchedBy(func(c *models.Contact) bool {
		return c.UserID == testUserID
	})).Return(&models.Contact{ID: 8}, nil)
	svc.On("GetContactByID", mock.Anything, &dtos.GetContactParams{ID: 8}).Return(sampleDetail(8), nil)

	resp, body := call(t, newTestApp(NewContactController(svc).CreateContactByAuthUser, true),
		`{"type_contact_id":1,"user_id":999,"ref_num":"REF-1","status":1}`, token)

	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	assert.Equal(t, "Contact created successfully", body.Message)
	svc.AssertExpectations(t)
}

func TestGetContactByIDByAuthUser_ZeroID_400(t *testing.T) {
	svc := new(mockContactService)
	resp, _ := call(t, newTestApp(NewContactController(svc).GetContactByIDByAuthUser, true), `{"id":0}`, "")
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestGetContactByIDByAuthUser_NoToken_401(t *testing.T) {
	svc := new(mockContactService)
	resp, body := call(t, newTestApp(NewContactController(svc).GetContactByIDByAuthUser, true), `{"id":9}`, "")

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	assert.Equal(t, "Unauthorized", body.Message)
}

func TestGetContactByIDByAuthUser_ScopesToCaller_200(t *testing.T) {
	token := authToken(t)
	svc := new(mockContactService)
	// Lookup must be scoped by the caller's user id, not just the contact id.
	svc.On("GetContactByID", mock.Anything, &dtos.GetContactParams{ID: 9, UserID: testUserID}).
		Return(sampleDetail(9), nil)

	resp, body := call(t, newTestApp(NewContactController(svc).GetContactByIDByAuthUser, true), `{"id":9}`, token)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "Contact fetched successfully", body.Message)
	svc.AssertExpectations(t)
}

func TestUpdateContactByAuthUser_NoToken_401(t *testing.T) {
	svc := new(mockContactService)

	// Validation runs before the auth check, so give it a passing body.
	updateValidationPasses(stubValidatorDB(t))
	resp, body := call(t, newTestApp(NewContactController(svc).UpdateContactByAuthUser, true), updateBody, "")

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	assert.Equal(t, "Unauthorized", body.Message)
}

func TestUpdateContactByAuthUser_ScopesToCaller_200(t *testing.T) {
	token := authToken(t)
	updateValidationPasses(stubValidatorDB(t))
	svc := new(mockContactService)
	svc.On("GetContactByID", mock.Anything, &dtos.GetContactParams{ID: 5, UserID: testUserID}).
		Return(sampleDetail(5), nil).Once()
	svc.On("UpdateContact", mock.Anything, mock.MatchedBy(func(c *models.Contact) bool {
		return c.ID == 5 && c.UserID == testUserID
	})).Return(&models.Contact{ID: 5}, nil)
	svc.On("GetContactByID", mock.Anything, &dtos.GetContactParams{ID: 5}).
		Return(sampleDetail(5), nil).Once()

	resp, body := call(t, newTestApp(NewContactController(svc).UpdateContactByAuthUser, true), updateBody, token)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "Contact updated successfully", body.Message)
	svc.AssertExpectations(t)
}

func TestDeleteContactByAuthUser_NoToken_401(t *testing.T) {
	svc := new(mockContactService)
	resp, body := call(t, newTestApp(NewContactController(svc).DeleteContactByAuthUser, true), `{"id":3}`, "")

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	assert.Equal(t, "Unauthorized", body.Message)
	svc.AssertNotCalled(t, "DeleteContact", mock.Anything, mock.Anything)
}

func TestDeleteContactByAuthUser_ScopesToCaller_200(t *testing.T) {
	token := authToken(t)
	svc := new(mockContactService)
	svc.On("GetContactByID", mock.Anything, &dtos.GetContactParams{ID: 3, UserID: testUserID}).
		Return(sampleDetail(3), nil)
	svc.On("DeleteContact", mock.Anything, uint(3)).Return(nil)

	resp, body := call(t, newTestApp(NewContactController(svc).DeleteContactByAuthUser, true), `{"id":3}`, token)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "Contact deleted successfully", body.Message)
	svc.AssertExpectations(t)
}

func TestRestoreContactByAuthUser_NoToken_401(t *testing.T) {
	svc := new(mockContactService)
	resp, body := call(t, newTestApp(NewContactController(svc).RestoreContactByAuthUser, true), `{"id":4}`, "")

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	assert.Equal(t, "Unauthorized", body.Message)
}

func TestRestoreContactByAuthUser_ScopesToCaller_200(t *testing.T) {
	token := authToken(t)
	svc := new(mockContactService)
	svc.On("GetContactByID", mock.Anything, mock.MatchedBy(func(p *dtos.GetContactParams) bool {
		return p.ID == 4 && p.UserID == testUserID && p.IsDeleted != nil && *p.IsDeleted == 1
	})).Return(sampleDetail(4), nil)
	svc.On("RestoreContact", mock.Anything, uint(4)).Return(nil)

	resp, body := call(t, newTestApp(NewContactController(svc).RestoreContactByAuthUser, true), `{"id":4}`, token)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "Contact restored successfully", body.Message)
	svc.AssertExpectations(t)
}
