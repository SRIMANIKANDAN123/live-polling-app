package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"live-polling-app/backend/internal/repositories"
	"live-polling-app/backend/internal/services"
	"live-polling-app/backend/internal/utils"
)

type AuthHandler struct {
	auth  *services.AuthService
	users *repositories.UserRepository
}

func NewAuthHandler(auth *services.AuthService, users *repositories.UserRepository) *AuthHandler {
	return &AuthHandler{auth: auth, users: users}
}

type registerRequest struct {
	Email           string `json:"email" binding:"required"`
	Password        string `json:"password" binding:"required"`
	ConfirmPassword string `json:"confirmPassword"`
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, http.StatusBadRequest, "VALIDATION_ERROR", "Email and password are required")
		return
	}
	if req.ConfirmPassword != "" && req.ConfirmPassword != req.Password {
		utils.Fail(c, http.StatusBadRequest, "PASSWORD_MISMATCH", "Passwords do not match")
		return
	}

	user, token, err := h.auth.Register(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		status, code := mapAuthError(err)
		utils.Fail(c, status, code, err.Error())
		return
	}

	utils.Success(c, http.StatusCreated, gin.H{
		"token": token,
		"user":  gin.H{"id": user.ID.Hex(), "email": user.Email},
	})
}

type loginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, http.StatusBadRequest, "VALIDATION_ERROR", "Email and password are required")
		return
	}

	user, token, err := h.auth.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		status, code := mapAuthError(err)
		utils.Fail(c, status, code, err.Error())
		return
	}

	utils.Success(c, http.StatusOK, gin.H{
		"token": token,
		"user":  gin.H{"id": user.ID.Hex(), "email": user.Email},
	})
}

func (h *AuthHandler) Me(c *gin.Context) {
	userID := c.GetString("userID")
	oid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		utils.Fail(c, http.StatusUnauthorized, "INVALID_TOKEN", "Invalid session")
		return
	}
	user, err := h.users.FindByID(c.Request.Context(), oid)
	if err != nil {
		utils.Fail(c, http.StatusNotFound, "USER_NOT_FOUND", "User not found")
		return
	}
	utils.Success(c, http.StatusOK, gin.H{"id": user.ID.Hex(), "email": user.Email, "createdAt": user.CreatedAt})
}

func mapAuthError(err error) (int, string) {
	switch {
	case errors.Is(err, services.ErrInvalidEmail):
		return http.StatusBadRequest, "INVALID_EMAIL"
	case errors.Is(err, services.ErrWeakPassword):
		return http.StatusBadRequest, "WEAK_PASSWORD"
	case errors.Is(err, services.ErrEmailInUse):
		return http.StatusConflict, "EMAIL_IN_USE"
	case errors.Is(err, services.ErrInvalidCreds):
		return http.StatusUnauthorized, "INVALID_CREDENTIALS"
	default:
		return http.StatusInternalServerError, "INTERNAL_ERROR"
	}
}
