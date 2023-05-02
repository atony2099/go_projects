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
	Parent    string `json:"parent"`
}

func NewTask(c *gin.Context) {
	var req TaskRequest
	var err error

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Parse the time strings into Time objects
	startTime, endTime, err := parseTime(req.StartTime, req.EndTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	duration := min(req.Duration, 1500)

	// Create a new task
	err = db.CreateTaskLog(startTime, endTime, duration, req.Project, req.Task, req.Parent)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Return a success response
	c.JSON(http.StatusOK, gin.H{
		"message": "Task created",
	})
}

func parseTime(start, end string) (time.Time, time.Time, error) {
	startTime, err := time.Parse(time.RFC3339, start)
	if err != nil {
		return startTime, time.Time{}, err
	}

	endTime, err := time.Parse(time.RFC3339, end)
	if err != nil {
		return startTime, endTime, err
	}

	return startTime, endTime, nil
}
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
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
	getDayLogsAndDuration(start, end, c)

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
	getDayLogsAndDuration(start, end, c)

}

func getDayLogsAndDuration(start, end time.Time, c *gin.Context) {
	dayDuration, err := db.DurationsByDate(start, end)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	dayLogs, err := db.GetTaskLog(start, end)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{"total": dayDuration, "logs": dayLogs}})
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
