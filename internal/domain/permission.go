package domain

import "time"

type Permission struct {
	Id         string    `json:"id" db:"id"`
	UserId     string    `json:"userId" db:"user_id"`
	ClientId   string    `json:"clientId" db:"client_id"`
	Permission string    `json:"permission" db:"permission"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
}
