package customerrs

import (
	"fmt"
	"time"

	"github.com/rs/xid"
)

type APIError struct {
	Code        int       `json:"code"`
	Message     string    `json:"message"`
	UserMessage string    `json:"userMessage,omitempty"`
	Id          string    `json:"id"`
	Timestamp   time.Time `json:"timestamp"`
}

func (e APIError) Error() string {
	return fmt.Sprintf("%v %v %v: %v", e.Timestamp, e.Id, e.Code, e.Message)
}

func NewAPIError(code int, message string, userMessage string) APIError {
	return APIError{
		Code:        code,
		Message:     message,
		UserMessage: userMessage,
		Id:          xid.New().String(),
		Timestamp:   time.Now(),
	}
}
