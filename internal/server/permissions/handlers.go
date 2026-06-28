package permissions

import (
	"encoding/json"
	"macauth/internal/service"
	"macauth/internal/shared/apierror"
	"macauth/internal/shared/httputil"
	"net/http"

	"github.com/go-playground/validator/v10"
)

type PermissionHandler struct {
	permissionService service.PermissionService
	validate          *validator.Validate
}

func NewPermissionHandler(permissionService service.PermissionService) *PermissionHandler {
	return &PermissionHandler{
		permissionService: permissionService,
		validate:          validator.New(),
	}
}

type HasPermissionRequest struct {
	UserId      string   `json:"user_id" validate:"required"`
	ClientId    string   `json:"client_id" validate:"required"`
	Permissions []string `json:"permissions" validate:"required"`
}

type HasPermissionResponse struct {
	HasPermission bool `json:"has_permission"`
}

func (h *PermissionHandler) HasPermission(w http.ResponseWriter, r *http.Request) error {
	reqBody, err := httputil.BindAndValidate[HasPermissionRequest](r, h.validate)
	if err != nil {
		return err
	}

	ctx := r.Context()
	hasPermission, err := h.permissionService.HasPermissions(ctx, reqBody.UserId, reqBody.ClientId, reqBody.Permissions)
	if err != nil {
		return apierror.InternalServerError(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := HasPermissionResponse{HasPermission: hasPermission}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		return apierror.InternalServerError(err)
	}

	return nil
}

type GetPermissionsRequest struct {
	UserId   string `json:"user_id"`
	ClientId string `json:"client_id"`
}

type GetPermissionsResponse struct {
	Permissions []string `json:"permissions"`
}

func (h *PermissionHandler) GetPermissions(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	reqBody, err := httputil.BindAndValidate[GetPermissionsRequest](r, h.validate)
	if err != nil {
		return err
	}

	permissions, err := h.permissionService.GetPermissions(ctx, reqBody.UserId, reqBody.ClientId)
	if err != nil {
		return apierror.InternalServerError(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(permissions); err != nil {
		return apierror.InternalServerError(err)
	}

	return nil
}

type AddPermissionsRequest struct {
	UserId      string   `json:"user_id"`
	ClientId    string   `json:"client_id"`
	Permissions []string `json:"permissions"`
}

type AddPermissionsResponse struct {
	Permissions []string `json:"permissions"`
}

func (h *PermissionHandler) AddPermissions(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	reqBody, err := httputil.BindAndValidate[AddPermissionsRequest](r, h.validate)
	if err != nil {
		return err
	}

	if err := h.permissionService.AddPermissions(ctx, reqBody.UserId, reqBody.ClientId, reqBody.Permissions); err != nil {
		return apierror.InternalServerError(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	resp := AddPermissionsResponse{Permissions: reqBody.Permissions}
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		return apierror.InternalServerError(err)
	}

	return nil
}

type RemovePermissionsRequest struct {
	UserId      string   `json:"user_id"`
	ClientId    string   `json:"client_id"`
	Permissions []string `json:"permissions"`
}

func (h *PermissionHandler) RemovePermissions(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	reqBody, err := httputil.BindAndValidate[RemovePermissionsRequest](r, h.validate)
	if err != nil {
		return err
	}

	if err := h.permissionService.RemovePermissions(ctx, reqBody.UserId, reqBody.ClientId, reqBody.Permissions); err != nil {
		return apierror.InternalServerError(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return nil
}
