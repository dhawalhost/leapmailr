package main

import (
	"fmt"

	"github.com/dhawalhost/leapmailr/config"
	"github.com/dhawalhost/leapmailr/handlers"
	"github.com/dhawalhost/leapmailr/logging"
	"github.com/dhawalhost/leapmailr/middleware"

	"github.com/gin-gonic/gin"
)

func main() {
	conf := config.LoadConfig()
	if conf.EnvMode == "release" {
		gin.SetMode(gin.ReleaseMode)

	}
	logger := logging.InitLogger()
	defer func() {
		if err := logger.Sync(); err != nil {
			fmt.Println(err)
		}
	}()
	if conf.EnvMode == "release" {
		logger.Info("release mode")
	}
	r := gin.New()
	r.Use(middleware.CorsMiddleware())
	r.Use(middleware.LoggerMiddleware())
	r.Use(gin.Recovery())
	rg := r.Group("/api/v1", middleware.LimitMiddleware())
	rg.POST("/contact", handlers.HandleContactForm)
	r.GET("/health", handlers.HandleHealthCheck)
	if err := r.Run(fmt.Sprintf(":%v", conf.Port)); err != nil {
		panic(err)
	}
}
