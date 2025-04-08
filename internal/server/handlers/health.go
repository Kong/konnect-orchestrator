package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

var applyHealthValue string

type HealthHandler struct {
	Version string
	Commit  string
	Date    string
}

func NewHealthHandler(version, commit, date string) *HealthHandler {
	return &HealthHandler{
		Version: version,
		Commit:  commit,
		Date:    date,
	}
}

func (h *HealthHandler) HealthCheck(c *gin.Context) {
	response := struct {
		ServerHealth string    `json:"serverStatus"`
		Timestamp    time.Time `json:"timestamp"`
		Version      string    `json:"version"`
		Commit       string    `json:"commit"`
		Date         string    `json:"date"`
	}{
		ServerHealth: "healthy",
		Timestamp:    time.Now(),
		Version:      h.Version,
		Commit:       h.Commit,
		Date:         h.Date,
	}

	c.JSON(http.StatusOK, response)
}
