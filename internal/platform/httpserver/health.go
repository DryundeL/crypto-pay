package httpserver

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type HealthHandler struct {
	db      *gorm.DB
	service string
	version string
	started time.Time
}

func NewHealthHandler(db *gorm.DB, service, version string) *HealthHandler {
	return &HealthHandler{
		db:      db,
		service: service,
		version: version,
		started: time.Now().UTC(),
	}
}

type liveResponse struct {
	Status  string `json:"status"`
	Service string `json:"service"`
	Version string `json:"version"`
	Uptime  string `json:"uptime"`
}

type readyResponse struct {
	Status  string            `json:"status"`
	Service string            `json:"service"`
	Version string            `json:"version"`
	Checks  map[string]string `json:"checks"`
}

func (h *HealthHandler) Live(c echo.Context) error {
	return c.JSON(http.StatusOK, liveResponse{
		Status:  "ok",
		Service: h.service,
		Version: h.version,
		Uptime:  time.Since(h.started).Truncate(time.Second).String(),
	})
}

func (h *HealthHandler) Ready(c echo.Context) error {
	checks := map[string]string{"database": "ok"}

	if h.db == nil {
		checks["database"] = "unavailable"
		return c.JSON(http.StatusServiceUnavailable, readyResponse{
			Status:  "not_ready",
			Service: h.service,
			Version: h.version,
			Checks:  checks,
		})
	}

	sqlDB, err := h.db.DB()
	if err != nil {
		checks["database"] = "error"
		return c.JSON(http.StatusServiceUnavailable, readyResponse{
			Status:  "not_ready",
			Service: h.service,
			Version: h.version,
			Checks:  checks,
		})
	}

	if err := sqlDB.PingContext(c.Request().Context()); err != nil {
		checks["database"] = "error"
		return c.JSON(http.StatusServiceUnavailable, readyResponse{
			Status:  "not_ready",
			Service: h.service,
			Version: h.version,
			Checks:  checks,
		})
	}

	return c.JSON(http.StatusOK, readyResponse{
		Status:  "ready",
		Service: h.service,
		Version: h.version,
		Checks:  checks,
	})
}

func RegisterHealthRoutes(e *echo.Echo, h *HealthHandler) {
	e.GET("/health", h.Live)
	e.GET("/health/live", h.Live)
	e.GET("/health/ready", h.Ready)
}
