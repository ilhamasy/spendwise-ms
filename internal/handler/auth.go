package handler

import (
	"net/http"

	"spendwise-ms/internal/config"
	"spendwise-ms/internal/dto"
	"spendwise-ms/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

func Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "validation_error", Message: err.Error()})
		return
	}
	if err := validate.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "validation_error", Message: err.Error()})
		return
	}

	user, err := service.Register(config.DB, req.Name, req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusConflict, dto.ErrorResponse{Error: "conflict", Message: err.Error()})
		return
	}

	accessToken, refreshToken, err := service.GenerateTokens(user.ID, user.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "server_error", Message: "Failed to generate tokens"})
		return
	}

	c.JSON(http.StatusCreated, dto.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User: dto.UserInfo{
			ID:    user.ID,
			Name:  user.Name,
			Email: user.Email,
		},
	})
}

func Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "validation_error", Message: err.Error()})
		return
	}
	if err := validate.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "validation_error", Message: err.Error()})
		return
	}

	user, err := service.Login(config.DB, req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "unauthorized", Message: err.Error()})
		return
	}

	accessToken, refreshToken, err := service.GenerateTokens(user.ID, user.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "server_error", Message: "Failed to generate tokens"})
		return
	}

	c.JSON(http.StatusOK, dto.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User: dto.UserInfo{
			ID:    user.ID,
			Name:  user.Name,
			Email: user.Email,
		},
	})
}

func RefreshToken(c *gin.Context) {
	var req dto.RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "validation_error", Message: err.Error()})
		return
	}

	claims, err := service.ValidateToken(req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "unauthorized", Message: "Invalid refresh token"})
		return
	}

	accessToken, refreshToken, err := service.GenerateTokens(claims.UserID, claims.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "server_error", Message: "Failed to generate tokens"})
		return
	}

	c.JSON(http.StatusOK, dto.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	})
}
