package handler

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/gbplantwiki/gbplantwiki/internal/constants"
	"github.com/gbplantwiki/gbplantwiki/internal/dto"
	"github.com/gbplantwiki/gbplantwiki/internal/middleware"
	"github.com/gbplantwiki/gbplantwiki/internal/service"
	"github.com/gbplantwiki/gbplantwiki/internal/util"
)

// FavoriteGroupHandler exposes personal favorite-group endpoints.
type FavoriteGroupHandler struct {
	svc    *service.FavoriteGroupService
	logger *slog.Logger
}

// NewFavoriteGroupHandler creates a FavoriteGroupHandler.
func NewFavoriteGroupHandler(svc *service.FavoriteGroupService, logger *slog.Logger) *FavoriteGroupHandler {
	return &FavoriteGroupHandler{svc: svc, logger: logger}
}

// List handles GET /favorites/groups.
func (h *FavoriteGroupHandler) List(c *gin.Context) {
	data, err := h.svc.ListWithCounts(middleware.GetUserID(c))
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, dto.OK(data))
}

// Create handles POST /favorites/groups.
func (h *FavoriteGroupHandler) Create(c *gin.Context) {
	var req dto.FavoriteGroupCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(util.NewAppError(http.StatusBadRequest, constants.CodeBadRequest, constants.MsgInvalidParam))
		return
	}
	g, err := h.svc.Create(middleware.GetUserID(c), req.Name)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusCreated, dto.OK(g))
}

// Delete handles DELETE /favorites/groups/:id.
func (h *FavoriteGroupHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.Error(util.NewAppError(http.StatusBadRequest, constants.CodeBadRequest, "invalid group id"))
		return
	}
	if err := h.svc.Delete(middleware.GetUserID(c), uint(id)); err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"removed": true}))
}
