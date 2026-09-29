package router

import (
	"github.com/gin-gonic/gin"

	"github.com/gbplantwiki/gbplantwiki/internal/config"
	"github.com/gbplantwiki/gbplantwiki/internal/handler"
	"github.com/gbplantwiki/gbplantwiki/internal/middleware"
)

func registerFavoriteRoutes(v1 *gin.RouterGroup, cfg *config.Config, h *handler.FavoriteHandler, gh *handler.FavoriteGroupHandler, limiter *middleware.RateLimiter) {
	favorites := v1.Group("/favorites", middleware.AuthRequired(cfg))
	favorites.GET("", h.List)
	favorites.POST("", limiter.Limit(), h.Add)
	favorites.GET("/groups", gh.List)
	favorites.POST("/groups", limiter.Limit(), gh.Create)
	favorites.DELETE("/groups/:id", gh.Delete)
	favorites.PUT("/:targetType/:targetId/group", h.Move)
	favorites.DELETE("/:targetType/:targetId", h.Remove)
}
