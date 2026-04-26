package customerrs

// InternalServerError creates a 500 Internal Server Error response.
func InternalServerError(sysErr error) APIError {
	return NewAPIError(
		500,
		sysErr.Error(),
		"Internal server error",
	)
}

// ClientAlreadyExist creates a 409 Conflict error when attempting to create a client with a duplicate name.
func ClientAlreadyExist(sysErr error) APIError {
	return NewAPIError(
		409,
		sysErr.Error(),
		"Client already exist with this name",
	)
}

// ClientIdIsRequired creates a 400 Bad Request error when client ID is not provided in url params.
func ClientIdIsRequired(sysErr error) APIError {
	return NewAPIError(
		400,
		sysErr.Error(),
		"Client ID is required",
	)
}

// ClientNotFound creates a 404 Not Found error when attempting to unlink a non-existing client.
func ClientNotFound(sysErr error) APIError {
	return NewAPIError(
		404,
		sysErr.Error(),
		"Client not found",
	)
}

// UserAlreadyExist creates a 400 Bad Request error when user already exist with the provided email.
func UserAlreadyExist(sysErr error) APIError {
	return NewAPIError(
		400,
		sysErr.Error(),
		"User already exist with this email",
	)
}

func UserNotFound(sysErr error) APIError {
	return NewAPIError(
		404,
		sysErr.Error(),
		"User not found",
	)
}

func InvalidPassword(sysErr error) APIError {
	return NewAPIError(
		400,
		sysErr.Error(),
		"Invalid password",
	)
}
