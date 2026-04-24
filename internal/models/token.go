package models

type Token struct {
	Id           string `json:"id" db:"id"`
	RefreshToken string `json:"refreshToken" db:"refresh_token"`
	UserId       string `json:"userId" db:"user_id"`
	ClientId     string `json:"clientId" db:"client_id"`
}
