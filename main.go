package main

import (
	"fmt"

	"github.com/dhawalhost/leapmailr/config"
	"github.com/dhawalhost/leapmailr/handlers"
	"github.com/dhawalhost/leapmailr/middleware"

	"github.com/gin-gonic/gin"
)

func main() {
	conf := config.LoadConfig()
	if conf.EnvMode == "release" {
		gin.SetMode(gin.ReleaseMode)

	}
	r := gin.Default()

	r.Use(middleware.LimitMiddleware())

	r.POST("/api/v1/contact", handlers.HandleContactForm)

	r.Run(fmt.Sprintf(":%v", conf.Port))

}
