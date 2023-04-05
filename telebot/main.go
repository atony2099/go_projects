package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"syscall"
	"time"

	"github.com/atony2099/go_project/telebot/bot"
	"github.com/atony2099/go_project/telebot/db"
	"github.com/atony2099/go_project/telebot/elapse"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/go-co-op/gocron"
	"github.com/spf13/viper"
)

var TodayPassRemainInterva = 30

var NoStudyrRemainInterva = 30

type TaskRequest struct {
	StartTime string `json:"start_time"`
	Duration  int    `json:"duration"`
	Task      string `json:"task"`
	Project   string `json:"project"`
}

func main() {

	// helleo
	viper.AutomaticEnv()

	token := viper.GetString("TOKEN")
	dsn := viper.GetString("DB")
	chatid := viper.GetInt64("CHATID")

	// this one
	if token == "" || dsn == "" || chatid == 0 {
		log.Fatalf("env error:token: %s, dsn: %s, chatid: %d ", token, dsn, chatid)
	}

	db.Init(dsn)
	bot.NewBot(token, chatid)
	go bot.HandleCommand()
	go doCron()

	go router()

	s := make(chan os.Signal, 1)
	signal.Notify(s, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	fmt.Println("start success")
	<-s
	fmt.Println("clean")

}

func router() {
	router := gin.Default()

	router.Use(cors.Default())

	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	router.POST("/tasks", func(c *gin.Context) {
		// Parse the task request from the request body
		handleTask(c)

	})

	group := router.Group("/api")
	group.GET("/day/:input", func(c *gin.Context) {
		handleDetai(c)
	})

	group.GET("/day/range/:input", func(c *gin.Context) {
		handleRange(c)
	})

	group.GET("/day/range", func(c *gin.Context) {
		handleRange(c)
	})
	// Define a handler function for the task endpoint

	router.Run(":8080")
}

func handleTask(c *gin.Context) {
	var req TaskRequest
	var err error

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

	// Create a new task
	err = db.CreateTask(startTime, endTime, duration, req.Project, req.Task)

	if err != nil {
		// Handle error
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Return a success response
	c.JSON(http.StatusOK, gin.H{
		"message": "Task created",
	})
}

func handleRange(c *gin.Context) {
	startStr := c.Query("start")
	endStr := c.Query("end")

	start, err1 := time.Parse("2006-01-02", startStr)
	end, err2 := time.Parse("2006-01-02", endStr)

	// Check if there was an error parsing the dates
	if err1 != nil || err2 != nil {
		fmt.Println(start, end, "----")
		c.JSON(http.StatusOK, gin.H{
			"code": 1,
			"msg":  "Invalid date format",
		})
		return
	}

	days := int(end.Sub(start).Hours()/24) + 1
	_, d, _ := db.Detail(start, end, days)
	for key, times := range d {
		d[key] = times / time.Second
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": d,
	})
}

func handleDetai(c *gin.Context) {

	input := c.Param("input")

	re := regexp.MustCompile(`\d+`)
	match := re.FindString(input)
	if match == "" {
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "input error"})
	}

	current := time.Now()
	loc, _ := time.LoadLocation("Asia/Shanghai")
	start := time.Date(current.Year(), current.Month(), current.Day(), 0, 0, 0, 0, loc).UTC()
	end := start.Add(24 * time.Hour).Add(-1 * time.Second)

	var days int
	fmt.Sscan(match, &days)
	if days <= 0 {
		days = 1
	}
	start = start.Add(-time.Duration(days-1) * 24 * time.Hour)

	fmt.Println("start", start, "end", end, "days", days, "input", input)
	_, d, _ := db.Detail(start, end, days)

	for key, times := range d {
		d[key] = times / time.Second
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": d,
	})
}

func doCron() {
	bj, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		panic(err)
	}
	s := gocron.NewScheduler(bj)

	s.Cron(fmt.Sprintf("*/%d * * * *", TodayPassRemainInterva)).Do(func() {

		beijingLocation, _ := time.LoadLocation("Asia/Shanghai")
		h := time.Now().In(beijingLocation).Hour()
		if h <= 8 {
			return
		}

		// elapse time
		pass := elapse.Combine()
		bot.SendBotMsg(pass)

		po := db.Pomodoro()
		bot.SendBotMsg(po)

	})

	s.Cron(fmt.Sprintf("*/%d * * * *", NoStudyrRemainInterva)).Do(func() {

		beijingLocation, _ := time.LoadLocation("Asia/Shanghai")
		h := time.Now().In(beijingLocation).Hour()
		if h <= 9 {
			return
		}

		s := db.TodayLast(20)
		bot.SendBotMsg(s)

	})

	s.StartAsync()
}
