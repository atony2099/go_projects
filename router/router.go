package router

import (
	"net/http"

	"github.com/atony2099/time_manager/controller"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func SetupRouter() http.Handler {
	r := gin.Default()
	r.Use(cors.Default())

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	r.POST("/tasks", controller.NewTask)

	api := r.Group("/api")
	api.GET("/day/:input", controller.TasklogsDay)
	api.GET("/day/range", controller.TasklogsRange)

	api.GET("/cumulative/:input", controller.GetDayTotal)

	return r

}
