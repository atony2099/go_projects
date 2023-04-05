package db

import (
	"fmt"
	"log"
	"math"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type TaskLog struct {
	gorm.Model
	StartTime time.Time
	EndTime   time.Time
	Duration  int
	Task      string
	Project   string
}

var db *gorm.DB

var max = 1500

func Init(dsn string) {

	var err error
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("dns is %s db err: %v", dsn, err)
	}

}

func CreateTask(startTime time.Time, endTime time.Time, duration int, project string, task string) error {
	// Create a new task record
	taskR := TaskLog{
		StartTime: startTime,
		EndTime:   endTime,
		Duration:  duration,
		Task:      task,
		Project:   project,
	}

	// Save the task to the database
	return db.Create(&taskR).Error
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

func Detail(startDate, endDate time.Time, days int) (string, map[string]time.Duration, float64) {

	loc, _ := time.LoadLocation("Asia/Shanghai")

	var taskLogs []TaskLog
	db.Where("start_time >= ? AND start_time <= ?", startDate, endDate).Order("start_time ASC").Find(&taskLogs)

	durationByDate := make(map[string]time.Duration)
	var total time.Duration
	for _, record := range taskLogs {
		dateString := record.StartTime.In(loc).Format("2006-01-02")
		delta := record.EndTime.Sub(record.StartTime)
		durationByDate[dateString] += delta
		total += delta
	}
	fmt.Println(total, durationByDate, "total")

	hours := total.Hours()
	minutes := total.Minutes()
	average := total.Hours() / float64(days)

	var s string
	summary := fmt.Sprintf("summary:  total day: %d, total hour: %.0fh, total min: %.fmin, average: %.2fh\n\n", days, hours, minutes, average)
	s += summary
	for i := 0; i < days; i++ {
		beijingDate := endDate.In(loc).AddDate(0, 0, -i)
		dateString := beijingDate.In(loc).Format("2006-01-02")
		dateWeekday := beijingDate.Weekday().String()
		totalDuration := durationByDate[dateString]
		totalMinutes := totalDuration.Minutes()
		totalHours := totalMinutes / 60
		s += fmt.Sprintf("\n%s, %s, ", dateString, dateWeekday)

		empty := "🈳🈳🈳\n"
		if totalMinutes == 0 {
			durationByDate[dateString] = 0
			s += empty
		} else {
			s += fmt.Sprintf("%.fmin; %.2fh\n", totalMinutes, totalHours)
			for _, record := range taskLogs {
				if record.StartTime.In(loc).Format("2006-01-02") == dateString {
					start := record.StartTime.In(loc).Format(time.Kitchen)
					end := record.EndTime.In(loc).Format(time.Kitchen)
					pomo := ""
					if record.Duration >= max {
						pomo = " 🍅"
					}

					min := float64(record.Duration) / 60
					s += fmt.Sprintf("%s -- %s, duration: %.fm  %s\n", start, end, min, pomo)
				}
			}

		}

	}

	return s, durationByDate, average
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
