package models

import "time"

type CreateUserRequest struct {
	Name string `json:"name" validate:"required,notblank,min=2,max=100"`
	DOB  string `json:"dob" validate:"required,datetime=2006-01-02"`
}

type UpdateUserRequest struct {
	Name string `json:"name" validate:"required,notblank,min=2,max=100"`
	DOB  string `json:"dob" validate:"required,datetime=2006-01-02"`
}

type User struct {
	ID   int32
	Name string
	DOB  time.Time
	Age  int
}

type UserList struct {
	Users  []User
	Total  int64
	Limit  int32
	Offset int32
}

type UserResponse struct {
	ID   int32  `json:"id"`
	Name string `json:"name"`
	DOB  string `json:"dob"`
	Age  int    `json:"age"`
}

type ListUsersResponse struct {
	Data   []UserResponse `json:"data"`
	Total  int64          `json:"total"`
	Limit  int32          `json:"limit"`
	Offset int32          `json:"offset"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
