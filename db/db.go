package db

import (
	"fmt"
	"log"
	"math"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type TaskLog struct {
	gorm.Model
	StartTime time.Time
	EndTime   time.Time
	Duration  int
	TaskID    uint
	Project   string
}

type Task struct {
	gorm.Model
	TaskName    string
	ProjectName string
	ParentName  string
}

type DailyLog struct {
	EndTime string `json:"end_time"`
	Total   int    `json:"total"`
}

var db *gorm.DB

var max = 1500

func Open(dsn string) {
	var err error
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("dns is %s db err: %v", dsn, err)
	}
}

func Close() {
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Error getting SQL DB from GORM: %v", err)
		return
	}

	if err := sqlDB.Close(); err != nil {
		log.Fatalf("Error closing SQL DB: %v", err)
		return
	}

	log.Println("Database connection closed")
}

func GetTaskLog(begin, end time.Time) (map[string][]gin.H, error) {

	location, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		return nil, err
	}

	var logs []struct {
		TaskLog
		Task
	}

	db.Table("task_logs").Select("task_logs.*, tasks.*").
		Joins("LEFT JOIN tasks ON task_logs.task_id = tasks.id").
		Where("task_logs.start_time >= ? AND task_logs.start_time < ?", begin.In(location), end.In(location).Add(24*time.Hour)).
		Order("task_logs.start_time").
		Scan(&logs)

	fmt.Println("logs", logs)

	result := make(map[string][]gin.H)
	for _, log := range logs {
		date := log.TaskLog.StartTime.In(location).Format("2006-01-02")
		start := log.TaskLog.StartTime.In(location).Format("15:04:05")

		entry := gin.H{
			"start":    start,
			"duration": log.Duration,
			"task":     fmt.Sprintf("%s-%s", log.ParentName, log.TaskName),
			"project":  log.ProjectName,
		}

		if _, ok := result[date]; !ok {
			result[date] = []gin.H{}
		}

		result[date] = append(result[date], entry)
	}

	return result, nil
}
func GetDailyLog(begin, end time.Time) ([]TaskLog, error) {

	var logs []TaskLog
	err := db.Where("start_time >= ? AND start_time < ?", begin, end).Order("start_time").Find(&logs).Error

	return logs, err
}

func CreateTaskLog(startTime time.Time, endTime time.Time, duration int, project string, taskName string, parentName string) error {
	// Try to retrieve the task from the database or create a new one if it doesn't exist
	var task = Task{ProjectName: project, TaskName: taskName, ParentName: parentName}
	fmt.Println("project", project, "taskName", taskName, "parentName", parentName)
	err := db.Where(&task).FirstOrCreate(&task).Error
	if err != nil {
		return err
	}

	// Create a new task log record with the retrieved or created task ID
	taskLog := TaskLog{
		StartTime: startTime,
		EndTime:   endTime,
		Duration:  duration,
		TaskID:    task.ID,
	}

	// Save the task log to the database
	return db.Create(&taskLog).Error
}

func Pomodoro() string {

	currentTime := time.Now()

	// 获取北京时间当天零点的UTC时间
	beijingLocation, _ := time.LoadLocation("Asia/Shanghai")
	beijingDate := time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(), 0, 0, 0, 0, beijingLocation)
	utcBeijingDate := beijingDate.UTC()

	// 获取北京时间当天23:59:59的UTC时间
	beijingEndDate := beijingDate.Add(24 * time.Hour).Add(-1 * time.Second)
	utcBeijingEndDate := beijingEndDate.UTC()
	fmt.Println(utcBeijingDate, utcBeijingEndDate)

	// 查询当天 duration 总和
	type Result struct {
		Count    int
		Duration int
		Greate   int
	}

	var r Result
	db.Model(&TaskLog{}).
		Select("COUNT(*) as Count, SUM(duration) as  duration, SUM(case when duration >= 1500 then 1 else 0 end) as Greate").
		Where("start_time >= ? AND start_time <= ?", utcBeijingDate, utcBeijingEndDate).
		Scan(&r)

	sec := time.Duration(r.Duration)

	s := fmt.Sprintf("今日已完成:%d 🍅 \n未完成%d 💔 \n总时长:%.fm⌛", r.Greate, r.Count-r.Greate, sec.Minutes())
	return s
}

func TodayLast(minDelta int) string {

	now := time.Now()

	// 获取北京时间当天零点的UTC时间
	beijingLocation, _ := time.LoadLocation("Asia/Shanghai")
	now = now.In(beijingLocation)
	beijingDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, beijingLocation)
	utcBeijingDate := beijingDate.UTC()

	// 获取北京时间当天23:59:59的UTC时间
	beijingEndDate := beijingDate.Add(24 * time.Hour).Add(-1 * time.Second)
	utcBeijingEndDate := beijingEndDate.UTC()

	var r TaskLog
	db.Where("end_time >= ? AND end_time <= ?", utcBeijingDate, utcBeijingEndDate).Last(&r)

	var delta time.Duration

	if r.ID != 0 {
		delta = now.Sub(r.EndTime)
	} else {
		startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		delta = now.Sub(startOfDay)
	}

	totalMinutes := int(delta.Minutes())

	var h = totalMinutes / 60
	var m = totalMinutes % 60

	var s string
	if totalMinutes >= minDelta {
		s = fmt.Sprintf("😭😭😭你已经 %d hour, %d min 没有学习了 ", h, m)
	}

	return s
}

func DurationsByDate(startDate, endDate time.Time) (dayDuration map[string]int, err error) {
	loc, _ := time.LoadLocation("Asia/Shanghai")

	var taskLogs []TaskLog

	// Get the first record in the database
	var firstRecord TaskLog
	err = db.Order("start_time ASC").First(&firstRecord).Error

	if err != nil {
		return nil, err
	}

	// Check if startDate is before the first record's date, and adjust it
	if startDate.Before(firstRecord.StartTime) {
		startDate = firstRecord.StartTime
	}

	err = db.Where("start_time >= ? AND start_time <= ?", startDate, endDate).Order("start_time ASC").Find(&taskLogs).Error
	if err != nil {
		return nil, err
	}

	durationByDate := make(map[string]int)

	// Initialize the duration of all dates between startDate and endDate to zero
	for d := startDate; !d.After(endDate); d = d.AddDate(0, 0, 1) {
		dateString := d.In(loc).Format("2006-01-02")
		durationByDate[dateString] = 0
	}

	// Update the duration of each date based on the taskLogs
	for _, record := range taskLogs {
		dateString := record.StartTime.In(loc).Format("2006-01-02")
		durationByDate[dateString] += record.Duration
	}

	return durationByDate, nil
}

var Tasks []struct {
	Task       string
	TotalHours float64
}

type TaskGroup struct {
	Task         string
	TotalHours   float64
	MinStartTime time.Time
	MaxEndTime   time.Time
}

func GetTaskGroup() ([]TaskGroup, error) {

	var list []TaskGroup

	result := db.Table("task_logs").
		Select("task, SUM(duration)/3600.0 AS total_hours,MIN(start_time) AS min_start_time, MAX(end_time) AS max_end_time").
		Group("task").
		Find(&list)

	if result.Error != nil {
		return nil, result.Error
	}

	for _, task := range Tasks {
		task.TotalHours = math.Round(task.TotalHours*100) / 100

	}

	return list, nil

}
