package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/yourusername/user-management-api/internal/models"
	sqlcgen "github.com/yourusername/user-management-api/internal/repository/sqlc"
)

var ErrUserNotFound = errors.New("user not found")

type UserRepository interface {
	Create(ctx context.Context, name string, dob time.Time) (models.User, error)
	GetByID(ctx context.Context, id int32) (models.User, error)
	Update(ctx context.Context, id int32, name string, dob time.Time) (models.User, error)
	Delete(ctx context.Context, id int32) error
	List(ctx context.Context, limit int32, offset int32) ([]models.User, error)
	Count(ctx context.Context) (int64, error)
}

type userRepository struct {
	queries *sqlcgen.Queries
}

func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepository{
		queries: sqlcgen.New(db),
	}
}

func (r *userRepository) Create(ctx context.Context, name string, dob time.Time) (models.User, error) {
	user, err := r.queries.CreateUser(ctx, sqlcgen.CreateUserParams{
		Name: name,
		Dob:  dob,
	})
	if err != nil {
		return models.User{}, err
	}

	return toModel(user), nil
}

func (r *userRepository) GetByID(ctx context.Context, id int32) (models.User, error) {
	user, err := r.queries.GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.User{}, ErrUserNotFound
		}

		return models.User{}, err
	}

	return toModel(user), nil
}

func (r *userRepository) Update(ctx context.Context, id int32, name string, dob time.Time) (models.User, error) {
	user, err := r.queries.UpdateUser(ctx, sqlcgen.UpdateUserParams{
		ID:   id,
		Name: name,
		Dob:  dob,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.User{}, ErrUserNotFound
		}

		return models.User{}, err
	}

	return toModel(user), nil
}

func (r *userRepository) Delete(ctx context.Context, id int32) error {
	rowsAffected, err := r.queries.DeleteUser(ctx, id)
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrUserNotFound
	}

	return nil
}

func (r *userRepository) List(ctx context.Context, limit int32, offset int32) ([]models.User, error) {
	users, err := r.queries.ListUsers(ctx, sqlcgen.ListUsersParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, err
	}

	results := make([]models.User, 0, len(users))
	for _, user := range users {
		results = append(results, toModel(user))
	}

	return results, nil
}

func (r *userRepository) Count(ctx context.Context) (int64, error) {
	return r.queries.CountUsers(ctx)
}

func toModel(user sqlcgen.User) models.User {
	return models.User{
		ID:   user.ID,
		Name: user.Name,
		DOB:  user.Dob,
	}
}
