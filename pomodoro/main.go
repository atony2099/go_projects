package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type TaskRequest struct {
	StartTime string `json:"start_time"`
	Duration  int    `json:"duration"`
	Task      string `json:"task"`
	Project   string `json:"project"`
}

type TaskLog struct {
	gorm.Model
	StartTime time.Time
	EndTime   time.Time
	Duration  int
	Task      string
	Project   string
}

func main() {

	viper.AutomaticEnv()
	dsn := viper.GetString("DB")
	if dsn == "" {
		panic("env error")
	}

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {

		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	// Define a handler function for the task endpoint
	r.POST("/tasks", func(c *gin.Context) {
		// Parse the task request from the request body
		var req TaskRequest
		var err error
		defer func() {
			fmt.Println(err)

		}()

		if err := c.BindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Parse the time strings into Time objects
		startTime, err := time.Parse(time.RFC3339, req.StartTime)
		fmt.Println(err, req.StartTime)
		if err != nil {
			// Handle error
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		endTime := startTime.Add(time.Duration(req.Duration) * time.Second)

		if err != nil {
			// Handle error
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Calculate the duration in seconds
		duration := int(endTime.Sub(startTime).Seconds())

		// Create a new task record
		task := TaskLog{
			StartTime: startTime,
			EndTime:   endTime,
			Duration:  duration,
			Task:      req.Task,
			Project:   req.Project,
		}

		// Save the task to the database
		err = db.Create(&task).Error

		if err != nil {
			// Handle error
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Return a success response
		c.JSON(http.StatusOK, gin.H{
			"message": "Task created",
			"task":    task,
		})
	})

	// Start the server
	r.Run(":8080")
}
