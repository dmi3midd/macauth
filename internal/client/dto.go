package client

type LinkRequest struct {
	Name   string `json:"name" validate:"required,min=3,max=64"`
	Secret string `json:"secret" validate:"required,min=8,max=64"`
}

type LinkResponse struct {
	ClientId string `json:"clientId"`
}
