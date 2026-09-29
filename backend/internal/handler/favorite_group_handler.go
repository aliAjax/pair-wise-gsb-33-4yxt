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

// FavoriteGroupHandler exposes personal favorite group endpoints.
type FavoriteGroupHandler struct {
	svc    *service.FavoriteGroupService
	logger *slog.Logger
}

// NewFavoriteGroupHandler creates a FavoriteGroupHandler.
func NewFavoriteGroupHandler(svc *service.FavoriteGroupService, logger *slog.Logger) *FavoriteGroupHandler {
	return &FavoriteGroupHandler{svc: svc, logger: logger}
}

// List handles GET /favorite-groups.
func (h *FavoriteGroupHandler) List(c *gin.Context) {
	items, err := h.svc.List(middleware.GetUserID(c))
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, dto.OK(items))
}

// Create handles POST /favorite-groups.
func (h *FavoriteGroupHandler) Create(c *gin.Context) {
	var req dto.FavoriteGroupRequest
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

// Rename handles PUT /favorite-groups/:id.
func (h *FavoriteGroupHandler) Rename(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.Error(util.NewAppError(http.StatusBadRequest, constants.CodeBadRequest, "invalid group id"))
		return
	}
	var req dto.FavoriteGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(util.NewAppError(http.StatusBadRequest, constants.CodeBadRequest, constants.MsgInvalidParam))
		return
	}
	g, err := h.svc.Rename(middleware.GetUserID(c), uint(id), req.Name)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, dto.OK(g))
}

// Delete handles DELETE /favorite-groups/:id.
func (h *FavoriteGroupHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.Error(util.NewAppError(http.StatusBadRequest, constants.CodeBadRequest, "invalid group id"))
		return
	}
	detached, err := h.svc.Delete(middleware.GetUserID(c), uint(id))
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"deleted": true, "detached": detached}))
}
