package handler

import (
	"net/http"
	"strconv"
	"time"

	"spendwise-ms/internal/config"
	"spendwise-ms/internal/dto"
	"spendwise-ms/internal/repository"
	"spendwise-ms/internal/service"

	"github.com/gin-gonic/gin"
)

func getTxnService() *service.TransactionService {
	return service.NewTransactionService(config.DB)
}

func CreateTransaction(c *gin.Context) {
	userID := c.GetString("userId")

	var req dto.CreateTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "validation_error", Message: err.Error()})
		return
	}

	resp, err := getTxnService().CreateTransaction(userID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "bad_request", Message: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, resp)
}

func GetTransactions(c *gin.Context) {
	userID := c.GetString("userId")

	filter := repository.TransactionFilter{
		Type:       c.DefaultQuery("type", ""),
		CategoryID: c.DefaultQuery("categoryId", ""),
		StartDate:  c.DefaultQuery("startDate", ""),
		EndDate:    c.DefaultQuery("endDate", ""),
		Sort:       c.DefaultQuery("sort", "date_desc"),
		Search:     c.DefaultQuery("search", ""),
	}

	if month := c.DefaultQuery("month", ""); month != "" {
		year := c.DefaultQuery("year", strconv.Itoa(time.Now().Year()))
		filter.StartDate = year + "-" + month + "-01"
		lastDay := time.Date(time.Now().Year(), time.Month(getMonth(month)), 0, 0, 0, 0, 0, time.UTC).Day()
		filter.EndDate = year + "-" + month + "-" + strconv.Itoa(lastDay)
	} else if year := c.DefaultQuery("year", ""); year != "" {
		filter.StartDate = year + "-01-01"
		filter.EndDate = year + "-12-31"
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	filter.Page = page
	filter.Limit = limit

	resp, err := getTxnService().GetTransactions(userID, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "server_error", Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func GetTransactionByID(c *gin.Context) {
	userID := c.GetString("userId")
	id := c.Param("id")

	resp, err := getTxnService().GetTransactionByID(userID, id)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: "not_found", Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func UpdateTransaction(c *gin.Context) {
	userID := c.GetString("userId")
	id := c.Param("id")

	var req dto.UpdateTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "validation_error", Message: err.Error()})
		return
	}

	resp, err := getTxnService().UpdateTransaction(userID, id, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "bad_request", Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func DeleteTransaction(c *gin.Context) {
	userID := c.GetString("userId")
	id := c.Param("id")

	if err := getTxnService().DeleteTransaction(userID, id); err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: "not_found", Message: err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

func SyncTransactions(c *gin.Context) {
	userID := c.GetString("userId")

	var req dto.SyncRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "validation_error", Message: err.Error()})
		return
	}

	resp, err := getTxnService().SyncTransactions(userID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "bad_request", Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func getMonth(monthStr string) int {
	parsed, err := strconv.Atoi(monthStr)
	if err != nil || parsed < 1 || parsed > 12 {
		return 1
	}
	return parsed + 1
}
