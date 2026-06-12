package handler

import (
	"errors"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"

	"github.com/yourusername/user-management-api/internal/models"
	"github.com/yourusername/user-management-api/internal/repository"
	"github.com/yourusername/user-management-api/internal/service"
)

type UserHandler struct {
	service   service.UserService
	validator *validator.Validate
}

func NewValidator() (*validator.Validate, error) {
	validate := validator.New()
	if err := validate.RegisterValidation("notblank", func(field validator.FieldLevel) bool {
		return strings.TrimSpace(field.Field().String()) != ""
	}); err != nil {
		return nil, err
	}

	return validate, nil
}

func NewUserHandler(service service.UserService, validator *validator.Validate) *UserHandler {
	return &UserHandler{
		service:   service,
		validator: validator,
	}
}

func (h *UserHandler) CreateUser(c *fiber.Ctx) error {
	var request models.CreateUserRequest

	if err := c.BodyParser(&request); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	if err := h.validator.Struct(request); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, validationErrorMessage(err))
	}

	user, err := h.service.CreateUser(c.UserContext(), request)
	if err != nil {
		return mapServiceError(err)
	}

	return c.Status(fiber.StatusCreated).JSON(toUserResponse(user))
}

func (h *UserHandler) GetUserByID(c *fiber.Ctx) error {
	id, err := parseUserID(c.Params("id"))
	if err != nil {
		return mapServiceError(err)
	}

	user, err := h.service.GetUserByID(c.UserContext(), id)
	if err != nil {
		return mapServiceError(err)
	}

	return c.Status(fiber.StatusOK).JSON(toUserResponse(user))
}

func (h *UserHandler) UpdateUser(c *fiber.Ctx) error {
	id, err := parseUserID(c.Params("id"))
	if err != nil {
		return mapServiceError(err)
	}

	var request models.UpdateUserRequest
	if err := c.BodyParser(&request); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	if err := h.validator.Struct(request); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, validationErrorMessage(err))
	}

	user, err := h.service.UpdateUser(c.UserContext(), id, request)
	if err != nil {
		return mapServiceError(err)
	}

	return c.Status(fiber.StatusOK).JSON(toUserResponse(user))
}

func (h *UserHandler) DeleteUser(c *fiber.Ctx) error {
	id, err := parseUserID(c.Params("id"))
	if err != nil {
		return mapServiceError(err)
	}

	if err := h.service.DeleteUser(c.UserContext(), id); err != nil {
		return mapServiceError(err)
	}

	return c.Status(fiber.StatusNoContent).Send(nil)
}

func (h *UserHandler) ListUsers(c *fiber.Ctx) error {
	limit, err := parseOptionalInt32(c.Query("limit"), "limit")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	offset, err := parseOptionalInt32(c.Query("offset"), "offset")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	result, err := h.service.ListUsers(c.UserContext(), limit, offset)
	if err != nil {
		return mapServiceError(err)
	}

	return c.Status(fiber.StatusOK).JSON(toListUsersResponse(result))
}

func parseUserID(value string) (int32, error) {
	id, err := strconv.ParseInt(value, 10, 32)
	if err != nil || id <= 0 {
		return 0, service.ErrInvalidUserID
	}

	return int32(id), nil
}

func parseOptionalInt32(value string, fieldName string) (int32, error) {
	if value == "" {
		return 0, nil
	}

	parsed, err := strconv.ParseInt(value, 10, 32)
	if err != nil {
		return 0, errors.New(fieldName + " must be a valid integer")
	}

	return int32(parsed), nil
}

func validationErrorMessage(err error) string {
	var validationErrors validator.ValidationErrors
	if !errors.As(err, &validationErrors) {
		return "validation failed"
	}

	messages := make([]string, 0, len(validationErrors))
	for _, validationErr := range validationErrors {
		field := strings.ToLower(validationErr.Field())

		switch validationErr.Tag() {
		case "required":
			messages = append(messages, field+" is required")
		case "notblank":
			messages = append(messages, field+" cannot be blank")
		case "min":
			messages = append(messages, field+" must be at least "+validationErr.Param()+" characters long")
		case "max":
			messages = append(messages, field+" must be at most "+validationErr.Param()+" characters long")
		case "datetime":
			messages = append(messages, field+" must use YYYY-MM-DD format")
		default:
			messages = append(messages, field+" is invalid")
		}
	}

	return strings.Join(messages, ", ")
}

func mapServiceError(err error) error {
	switch {
	case errors.Is(err, service.ErrInvalidUserID),
		errors.Is(err, service.ErrInvalidDOBFormat),
		errors.Is(err, service.ErrFutureDOB),
		errors.Is(err, service.ErrInvalidPagination):
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	case errors.Is(err, repository.ErrUserNotFound):
		return fiber.NewError(fiber.StatusNotFound, err.Error())
	default:
		return err
	}
}

func toUserResponse(user models.User) models.UserResponse {
	return models.UserResponse{
		ID:   user.ID,
		Name: user.Name,
		DOB:  user.DOB.Format("2006-01-02"),
		Age:  user.Age,
	}
}

func toListUsersResponse(result models.UserList) models.ListUsersResponse {
	response := models.ListUsersResponse{
		Data:   make([]models.UserResponse, 0, len(result.Users)),
		Total:  result.Total,
		Limit:  result.Limit,
		Offset: result.Offset,
	}

	for _, user := range result.Users {
		response.Data = append(response.Data, toUserResponse(user))
	}

	return response
}
