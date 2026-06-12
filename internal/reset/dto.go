package reset

type InitiateResetRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type InitiateResetResponse struct {
	ResetToken string `json:"resetToken"`
	Email      string `json:"email"`
}

type ConfirmResetRequest struct {
	ResetToken  string `json:"resetToken" validate:"required,eq=64"`
	NewPassword string `json:"newPassword" validate:"required,min=8,max=64"`
}
