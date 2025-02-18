package service

import (
	"context"
	"errors"
	"sync"

	"github.com/nibroos/s-erp-api/service/internal/dtos"
	"github.com/nibroos/s-erp-api/service/internal/models"
	"github.com/nibroos/s-erp-api/service/internal/repository"
	"github.com/nibroos/s-erp-api/service/internal/utils"
	"github.com/opentracing/opentracing-go"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserService struct {
	repo     *repository.UserRepository
	utilRepo *repository.UtilRepository
	tracer   opentracing.Tracer
}

func NewUserService(repo *repository.UserRepository, utilRepo *repository.UtilRepository, tracer opentracing.Tracer) *UserService {
	return &UserService{repo: repo, utilRepo: utilRepo, tracer: tracer}
}

func (s *UserService) GetUsers(ctx context.Context, filters map[string]string, span opentracing.Span) ([]dtos.UserListDTO, int, error) {
	childSpan := opentracing.StartSpan("UserService-GetUsers", opentracing.ChildOf(span.Context()))
	users, total, err := s.repo.GetUsers(ctx, filters, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, 0, err
	}
	return users, total, nil
}

func (s *UserService) CreateUser(ctx context.Context, tx *gorm.DB, user *models.User, roleIDs []uint32, span opentracing.Span) (*models.User, error) {
	childSpan := opentracing.StartSpan("UserService-CreateUser", opentracing.ChildOf(span.Context()))
	// Hash password before saving
	if user.Password == "" {
		return nil, bcrypt.ErrHashTooShort
	}

	if len(roleIDs) == 0 {
		return nil, errors.New("roleIDs cannot be empty")
	}

	hashedPassword, err := utils.HashPassword(user.Password, childSpan)
	if err != nil {
		return nil, err
	}
	user.Password = hashedPassword

	// Create user
	if err := s.repo.CreateUser(tx, user, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, err
	}

	// Attach roles
	if err := s.repo.AttachRoles(tx, user, roleIDs, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, err
	}

	return user, nil
}

func (s *UserService) GetUserByID(ctx context.Context, params *dtos.GetUserByIDParams, span opentracing.Span) (*dtos.UserDetailDTO, error) {
	childSpan := opentracing.StartSpan("UserService-GetUserByID", opentracing.ChildOf(span.Context()))
	user, err := s.repo.GetUserByID(ctx, params)
	if err != nil {
		defer childSpan.Finish()
		return nil, err
	}
	return user, nil
}

func (s *UserService) UpdateUser(ctx context.Context, tx *gorm.DB, user *models.User, roleIDs []uint32, span opentracing.Span) (*models.User, error) {
	childSpan := opentracing.StartSpan("UserService-UpdateUser", opentracing.ChildOf(span.Context()))

	if len(roleIDs) == 0 {
		return nil, errors.New("roleIDs cannot be empty")
	}

	// Update user
	if err := s.repo.UpdateUser(tx, user, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, err
	}

	// Attach roles
	if err := s.repo.AttachRoles(tx, user, roleIDs, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, err
	}

	return user, nil
}

func (s *UserService) Authenticate(ctx context.Context, email, password string) (*dtos.UserDetailDTO, error) {
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(*user.Password), []byte(password)); err != nil {
		return nil, errors.New("invalid credentials")
	}

	return user, nil
}

func (s *UserService) DeleteUser(ctx context.Context, tx *gorm.DB, params *dtos.GetUserParams, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("UserService-DeleteUser", opentracing.ChildOf(span.Context()))

	var wg sync.WaitGroup
	wg.Add(2)

	// Channels for concurrent execution
	deleteRolesChan := make(chan error, 1)
	deleteUserChan := make(chan error, 1)

	// Goroutine for deleting roles
	go func() {
		defer wg.Done()
		err := s.repo.DeleteRolesByUserID(tx, params, childSpan)
		deleteRolesChan <- err
	}()

	// Goroutine for deleting user
	go func() {
		defer wg.Done()
		err := s.repo.DeleteUser(tx, params, childSpan)
		deleteUserChan <- err
	}()

	// Wait for both goroutines to finish
	wg.Wait()
	close(deleteRolesChan)
	close(deleteUserChan)

	deleteRolesErr := <-deleteRolesChan
	deleteUserErr := <-deleteUserChan

	if deleteRolesErr != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return deleteRolesErr
	}

	if deleteUserErr != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return deleteUserErr
	}

	return nil
}

func (s *UserService) RestoreUser(ctx context.Context, tx *gorm.DB, params *dtos.GetUserParams, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("UserService-RestoreUser", opentracing.ChildOf(span.Context()))
	// Restore user
	if err := s.repo.RestoreUser(tx, params, span); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return err
	}

	return nil
}
