package handler

import (
	"net/http"

	"spendwise-ms/internal/config"
	"spendwise-ms/internal/dto"
	"spendwise-ms/internal/service"

	"github.com/gin-gonic/gin"
)

func getBudgetService() *service.BudgetService {
	return service.NewBudgetService(config.DB)
}

func GetBudgets(c *gin.Context) {
	userID := c.GetString("userId")
	period := c.DefaultQuery("period", "")

	resp, err := getBudgetService().ListBudgets(userID, period)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "server_error", Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func CreateBudget(c *gin.Context) {
	userID := c.GetString("userId")

	var req dto.CreateBudgetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "validation_error", Message: err.Error()})
		return
	}

	resp, err := getBudgetService().CreateBudget(userID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "bad_request", Message: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, resp)
}

func GetBudgetByID(c *gin.Context) {
	userID := c.GetString("userId")
	id := c.Param("id")

	resp, err := getBudgetService().GetBudget(userID, id)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: "not_found", Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func UpdateBudget(c *gin.Context) {
	userID := c.GetString("userId")
	id := c.Param("id")

	var req dto.UpdateBudgetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "validation_error", Message: err.Error()})
		return
	}

	resp, err := getBudgetService().UpdateBudget(userID, id, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "bad_request", Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func DeleteBudget(c *gin.Context) {
	userID := c.GetString("userId")
	id := c.Param("id")

	if err := getBudgetService().DeleteBudget(userID, id); err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: "not_found", Message: err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}
