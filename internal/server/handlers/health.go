package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

var applyHealthValue string

// RepoHandler handles repository related requests
type HealthHandler struct {
	Version string
	Commit  string
	Date    string
}

// NewRepoHandler creates a new RepoHandler
func NewHealthHandler(applyHealth chan string, version, commit, date string) *HealthHandler {
	go func() {
		for result := range applyHealth {
			applyHealthValue = result
		}
	}()
	return &HealthHandler{
		Version: version,
		Commit:  commit,
		Date:    date,
	}
}

func (h *HealthHandler) HealthCheck(c *gin.Context) {
	response := struct {
		ServerHealth string    `json:"serverStatus"`
		ApplyHealth  string    `json:"applyStatus"`
		Timestamp    time.Time `json:"timestamp"`
		Version      string    `json:"version"`
		Commit       string    `json:"commit"`
		Date         string    `json:"date"`
	}{
		ServerHealth: "healthy",
		ApplyHealth:  applyHealthValue,
		Timestamp:    time.Now(),
		Version:      h.Version,
		Commit:       h.Commit,
		Date:         h.Date,
	}

	c.JSON(http.StatusOK, response)
}
