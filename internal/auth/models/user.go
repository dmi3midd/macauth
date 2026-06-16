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

type Permission struct {
	Id         string    `json:"id" db:"id"`
	UserId     string    `json:"userId" db:"user_id"`
	ClientId   string    `json:"clientId" db:"client_id"`
	Permission string    `json:"permission" db:"permission"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
}

func (u *User) NewUserDto() *UserDto {
	return &UserDto{
		UserId:      u.Id,
		Username:    u.Username,
		Email:       u.Email,
		Permissions: []string{},
	}
}
