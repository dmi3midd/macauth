package models

import "time"

type User struct {
	Id             string    `json:"id" db:"id"`
	Username       string    `json:"username" db:"username"`
	Email          string    `json:"email" db:"email"`
	IsAdmin        bool      `json:"isAdmin" db:"is_admin"`
	HashedPassword string    `json:"hashedPassword" db:"hashed_password"`
	CreatedAt      time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt      time.Time `json:"updatedAt" db:"updated_at"`
}

type UserDto struct {
	UserId   string
	Username string
	Email    string
	IsAdmin  bool
}

func NewUserDto(user *User) *UserDto {
	return &UserDto{
		UserId:   user.Id,
		Username: user.Username,
		Email:    user.Email,
		IsAdmin:  user.IsAdmin,
	}
}

type AuthDto struct {
	ClientId string
	User     UserDto
	Tokens   TokensPair
}
