package apierror

import "macauth/internal/service"

var ErrorMap = map[error]func(err error) error{
	service.ErrUserAlreadyExist: func(err error) error {
		return NewConflictError(err, "User already exist with this email")
	},
	service.ErrUserNotFound: func(err error) error {
		return NewNotFoundError(err, "User does not exist")
	},
	service.ErrInvalidPassword: func(err error) error {
		return NewBadRequestError(err, "Invalid password")
	},
	service.ErrUnexpectedSigningMethod: func(err error) error {
		return NewInternalServerError(err)
	},
	service.ErrInvalidRefreshToken: func(err error) error {
		return NewBadRequestError(err, "Invalid refresh token")
	},
	service.ErrSubjectAndIDNotFound: func(err error) error {
		return NewNotFoundError(err, "Subject and ID not found")
	},
	service.ErrInvalidAccessToken: func(err error) error {
		return NewBadRequestError(err, "Invalid access token")
	},
	service.ErrTokenNotFound: func(err error) error {
		return NewNotFoundError(err, "Token not found")
	},
	service.ErrTokenUsed: func(err error) error {
		return NewBadRequestError(err, "Reset token already used")
	},
	service.ErrTokenExpired: func(err error) error {
		return NewBadRequestError(err, "Reset token expired")
	},
	service.ErrInvalidToken: func(err error) error {
		return NewBadRequestError(err, "Invalid reset token")
	},
	service.ErrEmptyPermissions: func(err error) error {
		return NewBadRequestError(err, "Empty permissions")
	},
	service.ErrPermissionNotFound: func(err error) error {
		return NewNotFoundError(err, "Permission not found")
	},
	service.ErrTooManyPermissions: func(err error) error {
		return NewBadRequestError(err, "Too many permissions")
	},
}

func MapError(err error) error {
	if err == nil {
		return nil
	}
	if mappedErr, ok := ErrorMap[err]; ok {
		return mappedErr(err)
	}
	return NewInternalServerError(err)
}
