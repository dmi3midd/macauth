package models

type Permission struct {
	Id         string `json:"id" db:"id"`
	UserId     string `json:"userId" db:"user_id"`
	ClientId   string `json:"clientId" db:"client_id"`
	Permission string `json:"permission" db:"permission"`
	CreatedAt  string `json:"created_at" db:"created_at"`
	UpdatedAt  string `json:"updated_at" db:"updated_at"`
}

type PermissionDto struct {
	Id         string `json:"id" db:"id"`
	UserId     string `json:"userId" db:"user_id"`
	ClientId   string `json:"clientId" db:"client_id"`
	Permission string `json:"permission" db:"permission"`
}

func (p *Permission) NewPermissionDto() *PermissionDto {
	return &PermissionDto{
		Id:         p.Id,
		UserId:     p.UserId,
		ClientId:   p.ClientId,
		Permission: p.Permission,
	}
}
