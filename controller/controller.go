package controller

import (
	"fmt"
	"net/http"
	"regexp"
	"time"

	"github.com/atony2099/time_manager/db"
	"github.com/gin-gonic/gin"
)

type TaskRequest struct {
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
	Duration  int    `json:"duration"`
	Task      string `json:"task"`
	Project   string `json:"project"`
}

func NewTask(c *gin.Context) {
	var req TaskRequest
	var err error

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Parse the time strings into Time objects
	startTime, err := time.Parse(time.RFC3339, req.StartTime)
	if err != nil {
		// Handle error
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	endTime, err := time.Parse(time.RFC3339, req.EndTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	duration := req.Duration
	if duration >= 1500 {
		duration = 1500
	}

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

func TasklogsRange(c *gin.Context) {
	startStr := c.Query("start")
	endStr := c.Query("end")
	start, err1 := time.Parse("2006-01-02", startStr)
	end, err2 := time.Parse("2006-01-02", endStr)

	// Check if there was an error parsing the dates
	if err1 != nil || err2 != nil {
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "Invalid date format"})
		return
	}
	queryLogs(start, end, c)

}

func TasklogsDay(c *gin.Context) {

	input := c.Param("input")
	re := regexp.MustCompile(`\d+`)
	match := re.FindString(input)

	if match == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "input error"})
		return
	}

	days := 1
	if _, err := fmt.Sscan(match, &days); err != nil || days <= 0 {
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "input error"})
		return
	}

	loc, _ := time.LoadLocation("Asia/Shanghai")
	current := time.Now().In(loc)
	beijingToday := time.Date(current.Year(), current.Month(), current.Day(), 0, 0, 0, 0, loc)
	utcToday := beijingToday.UTC()
	end := utcToday.Add(24 * time.Hour).Add(-time.Second)
	start := utcToday.AddDate(0, 0, -days+1)
	queryLogs(start, end, c)

}

func queryLogs(start, end time.Time, c *gin.Context) {

	d := db.DurationsByDate(start, end)

	for key, times := range d {
		d[key] = times / time.Second
	}

	m, err := db.GetTaskLog(start, end)

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{"total": d, "logs": m}})
}

type DailyLog struct {
	EndTime string `json:"end_time"`
	Total   int    `json:"total"`
}

func GetDayTotal(c *gin.Context) {
	loc, _ := time.LoadLocation("Asia/Shanghai")
	current := time.Now().In(loc)
	beijingToday := time.Date(current.Year(), current.Month(), current.Day(), 0, 0, 0, 0, loc)
	utcToday := beijingToday.UTC()
	end := utcToday.Add(24 * time.Hour).Add(-time.Second)

	days := 1
	if input := c.Param("input"); input != "" {
		if _, err := fmt.Sscan(input, &days); err != nil || days <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "input error"})
			return
		}
	}
	start := utcToday.AddDate(0, 0, -days+1)

	logs, err := db.GetDailyLog(start, end)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// initialize dlogs with an item for every hour from start to end
	var dlogs []DailyLog

	now := time.Now()
	for t := start; t.Before(now); t = t.Add(time.Hour) {
		dlogs = append(dlogs, DailyLog{EndTime: t.In(loc).Format(time.DateTime), Total: 0})
	}

	dlogs = append(dlogs, DailyLog{EndTime: now.In(loc).Format(time.DateTime), Total: 0})

	for _, log := range logs {
		// find the corresponding item in dlogs for this log
		for i := range dlogs {
			itemT, _ := time.ParseInLocation(time.DateTime, dlogs[i].EndTime, loc)
			// if the log ends before this hour, add the duration to the total for this hour
			if log.EndTime.Before(itemT) {
				dlogs[i].Total += log.Duration
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "data": dlogs})
}
