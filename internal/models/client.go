package models

import "time"

type Client struct {
	Id           string    `json:"id" db:"id"`
	Name         string    `json:"name" db:"name"`
	HashedSecret string    `json:"hashedSecret" db:"hashed_secret"`
	CreatedAt    time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt    time.Time `json:"updatedAt" db:"updated_at"`
}
