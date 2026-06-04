package handler

import (
	"net/http"

	"spendwise-ms/internal/config"
	"spendwise-ms/internal/dto"
	"spendwise-ms/internal/service"

	"github.com/gin-gonic/gin"
)

func getSyncService() *service.SyncService {
	return service.NewSyncService(config.DB)
}

func SyncData(c *gin.Context) {
	userID := c.GetString("userId")

	var req dto.DataSyncRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "validation_error", Message: err.Error()})
		return
	}

	resp, err := getSyncService().Sync(userID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "server_error", Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}
