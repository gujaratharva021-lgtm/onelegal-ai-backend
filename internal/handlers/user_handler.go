package handlers

import (
	"net/http"

	"legaltech-backend/internal/models"
	"legaltech-backend/internal/repositories"
	"legaltech-backend/internal/services"
	"legaltech-backend/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type UserHandler struct {
	userRepo    *repositories.UserRepository
	authService *services.AuthService
}

func NewUserHandler(authService *services.AuthService) *UserHandler {
	return &UserHandler{
		userRepo:    repositories.NewUserRepository(),
		authService: authService,
	}
}

func currentUserID(c *gin.Context) (uuid.UUID, bool) {
	userIDVal, exists := c.Get("user_id")
	if !exists {
		return uuid.UUID{}, false
	}
	userID, ok := userIDVal.(uuid.UUID)
	return userID, ok
}

func currentUserRole(c *gin.Context) (string, bool) {
	roleVal, exists := c.Get("role")
	if !exists {
		return "", false
	}
	role, ok := roleVal.(string)
	return role, ok
}

func (h *UserHandler) GetProfile(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	user, err := h.userRepo.FindByID(userID)
	if err != nil {
		response.Error(c, http.StatusNotFound, "User not found")
		return
	}

	response.Success(c, http.StatusOK, "Profile fetched successfully", user)
}

func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req models.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	user, err := h.userRepo.FindByID(userID)
	if err != nil {
		response.Error(c, http.StatusNotFound, "User not found")
		return
	}

	user.Name = req.Name
	user.Phone = req.Phone
	user.AvatarURL = req.AvatarURL
	user.Bio = req.Bio
	user.LawFirm = req.LawFirm
	user.BarNumber = req.BarNumber
	if req.GSTNumber != nil {
		user.GSTNumber = *req.GSTNumber
	}
	if req.AccountHolderName != nil {
		user.AccountHolderName = *req.AccountHolderName
	}
	if req.BankName != nil {
		user.BankName = *req.BankName
	}
	if req.AccountNumber != nil {
		user.AccountNumber = *req.AccountNumber
	}
	if req.IFSCCode != nil {
		user.IFSCCode = *req.IFSCCode
	}
	if req.UpiID != nil {
		user.UpiID = *req.UpiID
	}

	if err := h.userRepo.Update(user); err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to update profile")
		return
	}

	response.Success(c, http.StatusOK, "Profile updated successfully", user)
}

func (h *UserHandler) ChangePassword(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req models.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.authService.ChangePassword(userID, req.CurrentPassword, req.NewPassword); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Password changed successfully", nil)
}

// UpdateDeviceToken registers/refreshes this user's FCM push token — called
// by the app right after login and whenever Firebase rotates the token.
func (h *UserHandler) UpdateDeviceToken(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req models.UpdateDeviceTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	user, err := h.userRepo.FindByID(userID)
	if err != nil {
		response.Error(c, http.StatusNotFound, "User not found")
		return
	}

	user.DeviceToken = req.DeviceToken
	if err := h.userRepo.Update(user); err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to update device token")
		return
	}

	response.Success(c, http.StatusOK, "Device token registered", nil)
}
