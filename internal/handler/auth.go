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

func setTokenCookies(c *gin.Context, accessToken, refreshToken string) {
	c.SetCookie("spendwise-access-token", accessToken, int(service.AccessTokenTTL.Seconds()), "/", "", false, true)
	c.SetCookie("spendwise-refresh-token", refreshToken, int(service.RefreshTokenTTL.Seconds()), "/api/v1/auth/refresh", "", false, true)
}

func clearTokenCookies(c *gin.Context) {
	c.SetCookie("spendwise-access-token", "", -1, "/", "", false, true)
	c.SetCookie("spendwise-refresh-token", "", -1, "/api/v1/auth/refresh", "", false, true)
}

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

	setTokenCookies(c, accessToken, refreshToken)

	c.JSON(http.StatusCreated, dto.AuthResponse{
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

	setTokenCookies(c, accessToken, refreshToken)

	c.JSON(http.StatusOK, dto.AuthResponse{
		User: dto.UserInfo{
			ID:    user.ID,
			Name:  user.Name,
			Email: user.Email,
		},
	})
}

func RefreshToken(c *gin.Context) {
	tokenString, err := c.Cookie("spendwise-refresh-token")
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "unauthorized", Message: "Refresh token cookie missing"})
		return
	}

	claims, err := service.ValidateToken(tokenString)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "unauthorized", Message: "Invalid refresh token"})
		return
	}

	if claims.TokenType != service.RefreshToken {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "unauthorized", Message: "Token is not a refresh token"})
		return
	}

	accessToken, refreshToken, err := service.GenerateTokens(claims.UserID, claims.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "server_error", Message: "Failed to generate tokens"})
		return
	}

	setTokenCookies(c, accessToken, refreshToken)

	c.JSON(http.StatusOK, dto.AuthResponse{})
}

func Logout(c *gin.Context) {
	clearTokenCookies(c)
	c.JSON(http.StatusOK, gin.H{"message": "Logged out"})
}

func GoogleLogin(c *gin.Context) {
	var req dto.GoogleAuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "validation_error", Message: err.Error()})
		return
	}

	cfg := config.Load()
	if cfg.GoogleClientID == "" || cfg.GoogleClientSecret == "" {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "config_error", Message: "Google OAuth is not configured"})
		return
	}
	user, err := service.GoogleLogin(config.DB, req.Code, req.RedirectUri, cfg.GoogleClientID, cfg.GoogleClientSecret)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "unauthorized", Message: "Google login failed: " + err.Error()})
		return
	}

	accessToken, refreshToken, err := service.GenerateTokens(user.ID, user.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "server_error", Message: "Failed to generate tokens"})
		return
	}

	setTokenCookies(c, accessToken, refreshToken)

	c.JSON(http.StatusOK, dto.AuthResponse{
		User: dto.UserInfo{
			ID:    user.ID,
			Name:  user.Name,
			Email: user.Email,
		},
	})
}
