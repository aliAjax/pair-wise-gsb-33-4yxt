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

// FavoriteHandler exposes favorite endpoints.
type FavoriteHandler struct {
	svc    *service.FavoriteService
	logger *slog.Logger
}

// NewFavoriteHandler creates a FavoriteHandler.
func NewFavoriteHandler(svc *service.FavoriteService, logger *slog.Logger) *FavoriteHandler {
	return &FavoriteHandler{svc: svc, logger: logger}
}

// List handles GET /favorites.
func (h *FavoriteHandler) List(c *gin.Context) {
	filter := service.ListFilter{TargetType: c.Query("target_type")}
	if raw := c.Query("group_id"); raw != "" {
		id, err := strconv.ParseUint(raw, 10, 64)
		if err != nil || id == 0 {
			c.Error(util.NewAppError(http.StatusBadRequest, constants.CodeBadRequest, "invalid group id"))
			return
		}
		gid := uint(id)
		filter.GroupID = &gid
	}
	filter.Uncategorized = c.Query("uncategorized") == "true" || c.Query("uncategorized") == "1"
	items, err := h.svc.ListByUser(middleware.GetUserID(c), filter)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, dto.OK(items))
}

// Add handles POST /favorites.
func (h *FavoriteHandler) Add(c *gin.Context) {
	var req dto.FavoriteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(util.NewAppError(http.StatusBadRequest, constants.CodeBadRequest, constants.MsgInvalidParam))
		return
	}
	f, err := h.svc.Add(middleware.GetUserID(c), req.TargetType, req.TargetID, req.GroupID)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusCreated, dto.OK(f))
}

// Move handles PUT /favorites/:targetType/:targetId/group.
func (h *FavoriteHandler) Move(c *gin.Context) {
	targetType := c.Param("targetType")
	targetID, err := strconv.ParseUint(c.Param("targetId"), 10, 64)
	if err != nil {
		c.Error(util.NewAppError(http.StatusBadRequest, constants.CodeBadRequest, "invalid target id"))
		return
	}
	var req dto.FavoriteMoveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(util.NewAppError(http.StatusBadRequest, constants.CodeBadRequest, constants.MsgInvalidParam))
		return
	}
	f, err := h.svc.Move(middleware.GetUserID(c), targetType, uint(targetID), req.GroupID)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, dto.OK(f))
}

// Remove handles DELETE /favorites/:targetType/:targetId.
func (h *FavoriteHandler) Remove(c *gin.Context) {
	targetType := c.Param("targetType")
	targetID, err := strconv.ParseUint(c.Param("targetId"), 10, 64)
	if err != nil {
		c.Error(util.NewAppError(http.StatusBadRequest, constants.CodeBadRequest, "invalid target id"))
		return
	}
	if err := h.svc.Remove(middleware.GetUserID(c), targetType, uint(targetID)); err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"removed": true}))
}
