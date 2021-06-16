package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/mahadeva604/audio-storage/pkg/service"

	"github.com/swaggo/gin-swagger"
	"github.com/swaggo/gin-swagger/swaggerFiles"

	_ "github.com/mahadeva604/audio-storage/docs"
)

type Handler struct {
	services *service.Service
}

func NewHandler(services *service.Service) *Handler {
	return &Handler{services: services}
}

func (h *Handler) InitRoutes() *gin.Engine {
	router := gin.New()

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	auth := router.Group("/auth")
	{
		auth.POST("/sign-up", h.signUp)
		auth.POST("/sign-in", h.signIn)
		auth.POST("/refresh", h.refreshTokens)
	}

	api := router.Group("/api", h.userIdentity)
	{
		audio := api.Group("/audio")
		{
			audio.GET("/", h.getAllAudio)
			audio.POST("/", h.uploadAudio)
			audio.PUT("/:id", h.addDescription)
			audio.GET("/:id", h.downloadAudio)
		}

		share := api.Group("share")
		{
			share.POST("/:id", h.shareAudio)
			share.DELETE("/:id", h.unshareAudio)
		}

		api.GET("/shares", h.getSharedAudio)
	}

	return router
}
