package elapse

import (
	"fmt"
	"strings"
	"time"
)

func Combine() string {
	s := LoadTime()
	m := LoadM()
	y := LoadY()
	var list = []string{s, m, y}
	str := strings.Join(list, "\n\n")
	return str
}

func LoadTime() string {

	pass, total := getPercent()
	bar := bar(pass, total, 93)
	reaminStr := fmt.Sprintf("🚗今天还有%dm, %.2f小时可以挥霍", total-pass, float64(total-pass)/60)
	timeDesc := bar + "\n" + reaminStr
	return timeDesc

}

func LoadM() string {

	bj, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		panic(err)
	}

	// Get the current time in Beijing timezone
	now := time.Now().In(bj)
	pass := now.Day()

	year, month, _ := now.Date()
	total := time.Date(year+1, month+1, 0, 0, 0, 0, 0, bj).Day()
	remain := total - pass

	bar := bar(pass, total, 2)

	reaminStr := fmt.Sprintf("\n📱本月还剩%d天", remain)

	return bar + reaminStr

}

func LoadY() string {

	bj, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		panic(err)
	}

	// Get the current time in Beijing timezone
	now := time.Now().In(bj)
	passDay := now.YearDay()

	year, _, _ := now.Date()
	total := time.Date(year+1, 0, 0, 0, 0, 0, 0, bj).YearDay()
	remain := total - passDay
	// fmt.Println(passDay, total)

	bar := bar(passDay, total, 22)

	reaminStr := fmt.Sprintf("\n ⏰今年还剩%d天", remain)

	return bar + reaminStr

}

func bar(pass, total, rate int) string {

	progress := strings.Repeat("█", pass/rate)

	remain := strings.Repeat("▓", (total-pass)/rate)
	return fmt.Sprintf("%s%s %d/%d - %.f%%", progress, remain, pass, total, 100*float64(pass)/float64(total))

	// return fmt.Sprintf("[%-"+fmt.Sprintf("%d", total/rate)+"s] %d/%d", progress, pass, total)

}

func getPercent() (int, int) {

	// Load the Asia/Shanghai timezone location
	bj, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		panic(err)
	}

	// Get the current time in Beijing timezone
	now := time.Now().In(bj)

	// Calculate the elapsed time in minutes since the start of the day
	var total = 24 * 60
	elapsedMinutes := int(now.Sub(time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, bj)).Minutes())

	// percent := float64(elapsedMinutes) / float64(total)
	// value := int(math.Round(percent * 100))

	return elapsedMinutes, total
}
