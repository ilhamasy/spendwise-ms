package handler

import (
	"net/http"
	"strconv"

	"spendwise-ms/internal/config"
	"spendwise-ms/internal/dto"
	"spendwise-ms/internal/service"

	"github.com/gin-gonic/gin"
)

func getGoalService() *service.GoalService {
	return service.NewGoalService(config.DB)
}

func GetGoals(c *gin.Context) {
	userID := c.GetString("userId")
	status := c.DefaultQuery("status", "active")

	resp, err := getGoalService().ListGoals(userID, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "server_error", Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func CreateGoal(c *gin.Context) {
	userID := c.GetString("userId")

	var req dto.CreateGoalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "validation_error", Message: err.Error()})
		return
	}

	resp, err := getGoalService().CreateGoal(userID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "bad_request", Message: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, resp)
}

func GetGoalByID(c *gin.Context) {
	userID := c.GetString("userId")
	id := c.Param("id")

	resp, err := getGoalService().GetGoal(userID, id)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: "not_found", Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func UpdateGoal(c *gin.Context) {
	userID := c.GetString("userId")
	id := c.Param("id")

	var req dto.UpdateGoalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "validation_error", Message: err.Error()})
		return
	}

	resp, err := getGoalService().UpdateGoal(userID, id, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "bad_request", Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func ArchiveGoal(c *gin.Context) {
	userID := c.GetString("userId")
	id := c.Param("id")

	resp, err := getGoalService().ArchiveGoal(userID, id)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: "not_found", Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func UnarchiveGoal(c *gin.Context) {
	userID := c.GetString("userId")
	id := c.Param("id")

	resp, err := getGoalService().UnarchiveGoal(userID, id)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: "not_found", Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func DeleteGoal(c *gin.Context) {
	userID := c.GetString("userId")
	id := c.Param("id")

	if err := getGoalService().DeleteGoal(userID, id); err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: "not_found", Message: err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

func AddContribution(c *gin.Context) {
	userID := c.GetString("userId")
	goalID := c.Param("id")

	var req dto.CreateContributionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "validation_error", Message: err.Error()})
		return
	}

	contribResp, goalResp, err := getGoalService().AddContribution(userID, goalID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "bad_request", Message: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"contribution": contribResp,
		"goal":         goalResp,
	})
}

func GetContributionHistory(c *gin.Context) {
	userID := c.GetString("userId")
	goalID := c.Param("id")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	resp, err := getGoalService().GetContributions(userID, goalID, page, limit)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: "not_found", Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}
