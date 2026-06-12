package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/yourusername/user-management-api/internal/models"
	"github.com/yourusername/user-management-api/internal/repository"
	"github.com/yourusername/user-management-api/internal/utils"
)

var (
	ErrInvalidUserID     = errors.New("user id must be a positive integer")
	ErrInvalidDOBFormat  = errors.New("dob must use YYYY-MM-DD format")
	ErrFutureDOB         = errors.New("dob cannot be in the future")
	ErrInvalidPagination = errors.New("limit must be greater than zero and offset cannot be negative")
)

type UserService interface {
	CreateUser(ctx context.Context, request models.CreateUserRequest) (models.User, error)
	GetUserByID(ctx context.Context, id int32) (models.User, error)
	UpdateUser(ctx context.Context, id int32, request models.UpdateUserRequest) (models.User, error)
	DeleteUser(ctx context.Context, id int32) error
	ListUsers(ctx context.Context, limit int32, offset int32) (models.UserList, error)
}

type userService struct {
	repository       repository.UserRepository
	defaultPageLimit int32
	maxPageLimit     int32
	now              func() time.Time
}

func NewUserService(
	repository repository.UserRepository,
	defaultPageLimit int32,
	maxPageLimit int32,
	now func() time.Time,
) UserService {
	if now == nil {
		now = time.Now
	}

	return &userService{
		repository:       repository,
		defaultPageLimit: defaultPageLimit,
		maxPageLimit:     maxPageLimit,
		now:              now,
	}
}

func (s *userService) CreateUser(ctx context.Context, request models.CreateUserRequest) (models.User, error) {
	dob, err := parseDOB(request.DOB, s.now())
	if err != nil {
		return models.User{}, err
	}

	user, err := s.repository.Create(ctx, strings.TrimSpace(request.Name), dob)
	if err != nil {
		return models.User{}, err
	}

	return s.withAge(user), nil
}

func (s *userService) GetUserByID(ctx context.Context, id int32) (models.User, error) {
	if id <= 0 {
		return models.User{}, ErrInvalidUserID
	}

	user, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return models.User{}, err
	}

	return s.withAge(user), nil
}

func (s *userService) UpdateUser(ctx context.Context, id int32, request models.UpdateUserRequest) (models.User, error) {
	if id <= 0 {
		return models.User{}, ErrInvalidUserID
	}

	dob, err := parseDOB(request.DOB, s.now())
	if err != nil {
		return models.User{}, err
	}

	user, err := s.repository.Update(ctx, id, strings.TrimSpace(request.Name), dob)
	if err != nil {
		return models.User{}, err
	}

	return s.withAge(user), nil
}

func (s *userService) DeleteUser(ctx context.Context, id int32) error {
	if id <= 0 {
		return ErrInvalidUserID
	}

	return s.repository.Delete(ctx, id)
}

func (s *userService) ListUsers(ctx context.Context, limit int32, offset int32) (models.UserList, error) {
	limit, offset, err := s.normalizePagination(limit, offset)
	if err != nil {
		return models.UserList{}, err
	}

	users, err := s.repository.List(ctx, limit, offset)
	if err != nil {
		return models.UserList{}, err
	}

	total, err := s.repository.Count(ctx)
	if err != nil {
		return models.UserList{}, err
	}

	return models.UserList{
		Users:  s.withAges(users),
		Total:  total,
		Limit:  limit,
		Offset: offset,
	}, nil
}

func (s *userService) normalizePagination(limit int32, offset int32) (int32, int32, error) {
	if limit == 0 {
		limit = s.defaultPageLimit
	}

	if limit <= 0 || offset < 0 {
		return 0, 0, ErrInvalidPagination
	}

	if limit > s.maxPageLimit {
		limit = s.maxPageLimit
	}

	return limit, offset, nil
}

func (s *userService) withAge(user models.User) models.User {
	user.Age = utils.CalculateAge(user.DOB, s.now())
	return user
}

func (s *userService) withAges(users []models.User) []models.User {
	results := make([]models.User, 0, len(users))
	for _, user := range users {
		results = append(results, s.withAge(user))
	}

	return results
}

func parseDOB(value string, now time.Time) (time.Time, error) {
	dob, err := time.ParseInLocation("2006-01-02", value, time.UTC)
	if err != nil {
		return time.Time{}, ErrInvalidDOBFormat
	}

	today := time.Date(now.UTC().Year(), now.UTC().Month(), now.UTC().Day(), 0, 0, 0, 0, time.UTC)
	if dob.After(today) {
		return time.Time{}, ErrFutureDOB
	}

	return dob, nil
}
