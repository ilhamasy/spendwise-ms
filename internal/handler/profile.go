package handler

import (
	"net/http"

	"spendwise-ms/internal/config"
	"spendwise-ms/internal/dto"
	"spendwise-ms/internal/service"

	"github.com/gin-gonic/gin"
)

func getProfileService() *service.ProfileService {
	return service.NewProfileService(config.DB)
}

func GetProfile(c *gin.Context) {
	userID := c.GetString("userId")

	resp, err := getProfileService().GetProfile(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: "not_found", Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func UpdateProfile(c *gin.Context) {
	userID := c.GetString("userId")

	var req dto.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "validation_error", Message: err.Error()})
		return
	}

	resp, err := getProfileService().UpdateProfile(userID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "bad_request", Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}
