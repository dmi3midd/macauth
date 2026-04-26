package models

import "time"

type User struct {
	Id             string    `json:"id" db:"id"`
	Username       string    `json:"username" db:"username"`
	Email          string    `json:"email" db:"email"`
	HashedPassword string    `json:"hashedPassword" db:"hashed_password"`
	CreatedAt      time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt      time.Time `json:"updatedAt" db:"updated_at"`
}

type UserDto struct {
	UserId   string
	Username string
	Email    string
}

func NewUserDto(user *User) *UserDto {
	return &UserDto{
		UserId:   user.Id,
		Username: user.Username,
		Email:    user.Email,
	}
}

type AuthDto struct {
	ClientId string
	User     UserDto
	Tokens   TokensPair
}
