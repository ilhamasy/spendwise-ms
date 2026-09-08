package handler

import (
	"net/http"

	"spendwise-ms/internal/config"
	"spendwise-ms/internal/dto"
	"spendwise-ms/internal/service"

	"github.com/gin-gonic/gin"
)

func getCatService() *service.CategoryService {
	return service.NewCategoryService(config.DB)
}

func GetCategories(c *gin.Context) {
	userID := c.GetString("userId")
	filterType := c.DefaultQuery("type", "")

	resp, err := getCatService().ListCategories(userID, filterType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "server_error", Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func CreateCategory(c *gin.Context) {
	userID := c.GetString("userId")

	var req dto.CreateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "validation_error", Message: err.Error()})
		return
	}

	resp, err := getCatService().CreateCategory(userID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "bad_request", Message: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, resp)
}

func UpdateCategory(c *gin.Context) {
	userID := c.GetString("userId")
	id := c.Param("id")

	var req dto.UpdateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "validation_error", Message: err.Error()})
		return
	}

	resp, err := getCatService().UpdateCategory(userID, id, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "bad_request", Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func DeleteCategory(c *gin.Context) {
	userID := c.GetString("userId")
	id := c.Param("id")
	reassignTo := c.Query("reassignTo")

	if err := getCatService().DeleteCategory(userID, id, reassignTo); err != nil {
		c.JSON(http.StatusConflict, dto.ErrorResponse{Error: "conflict", Message: err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}
