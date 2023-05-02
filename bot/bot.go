package bot

import (
	"context"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/atony2099/time_manager/db"
	"github.com/atony2099/time_manager/elapse"
	"github.com/atony2099/time_manager/img"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/wcharczuk/go-chart/v2"
)

var bot *tgbotapi.BotAPI
var chatID int64

func SendBotMsg(message string) {
	msg := tgbotapi.NewMessage(chatID, message)
	bot.Send(msg)
}

func RunBot(token string, chat int64, ctx context.Context) {
	var err error
	bot, err = tgbotapi.NewBotAPI(token)
	chatID = chat
	if err != nil {
		log.Fatal(err)
	}
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 10
	updates := bot.GetUpdatesChan(u)
	for {
		select {
		case update, ok := <-updates:
			if !ok {
				log.Println("Updates channel closed")
				return
			}

			if update.Message != nil { // If we got a message
				handleUpdate(update)
			}

		case <-ctx.Done():
			log.Println("RunBot: stopping bot gracefully")
			return
		}
	}
}

func handleUpdate(update tgbotapi.Update) {
	if command := update.Message.Command(); command != "" {
		switch command {
		case "pass":
			handlePassCommand()
		case "d":
			// handleDCommand(update.Message.Text)
		case "group":
			handleGroupCommand()
		}
	}
}

func handlePassCommand() {
	msg := tgbotapi.NewMessage(chatID, elapse.Combine())
	bot.Send(msg)
}

// func handleDCommand(text string) {
// 	start, end, _ := parseDetailInput(text)
// 	d, _ := db.DurationsByDate(start, end)

// 	str, a := detailWithDurationByDate(d)
// 	b := img.Image(d, a)
// 	msg := tgbotapi.NewMessage(chatID, str)
// 	photo := tgbotapi.NewPhoto(chatID, tgbotapi.FileBytes{Name: "data", Bytes: b})
// 	bot.Send(msg)
// 	bot.Send(photo)
// }

func handleGroupCommand() {
	list, err := db.GetTaskGroup()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	values, str := prepareGroupOutput(list)

	b, err := img.Pip(values)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	msg := tgbotapi.NewMessage(chatID, str)
	bot.Send(msg)
	photo := tgbotapi.NewPhoto(chatID, tgbotapi.FileBytes{Name: "data", Bytes: b})
	bot.Send(photo)
}

func prepareGroupOutput(list []db.TaskGroup) ([]chart.Value, string) {
	var values []chart.Value
	var str string

	for _, v := range list {
		task := fmt.Sprintf("%s (%.2f)", v.Task, v.TotalHours)
		values = append(values, chart.Value{Value: v.TotalHours, Label: task})
		s := fmt.Sprintf("task:%s	start:%s, end:%s, totalH: %.2f\n", v.Task, v.MinStartTime.Format("2006-01-02"), v.MaxEndTime.Format("2006-01-02"), v.TotalHours)
		str += s
	}

	return values, str
}

func parseDetailInput(input string) (time.Time, time.Time, int) {

	strList := strings.Fields(input)

	var text string
	if len(strList) == 1 {
		text = "1"
	}

	if len(strList) >= 2 {
		text = strList[1]
	}

	fmt.Println(text, "go")

	// extract start and end times from the input
	re := regexp.MustCompile(`\d+\.\d+-\d+\.\d+`)
	match := re.FindString(text)
	fmt.Println(match, "xx")

	loc, _ := time.LoadLocation("Asia/Shanghai")
	current := time.Now().In(loc)

	var days = 1
	start := time.Date(current.Year(), current.Month(), current.Day(), 0, 0, 0, 0, loc).UTC()
	end := start.Add(24 * time.Hour).Add(-1 * time.Second)

	if match != "" {
		times := regexp.MustCompile(`\d+\.\d+`).FindAllString(match, -1)
		fmt.Println(times)
		start, _ = time.Parse("1.2", times[0])
		end, _ = time.Parse("1.2", times[1])
		start = time.Date(current.Year(), start.Month(), start.Day(), 0, 0, 0, 0, loc).UTC()
		end = time.Date(current.Year(), end.Month(), end.Day(), 23, 59, 59, 0, loc).UTC()
		days = int(end.Sub(start).Hours()/24) + 1

	} else {
		re = regexp.MustCompile(`\d+`)
		match = re.FindString(text)
		if match != "" {
			n, err := fmt.Sscan(match, &days)
			fmt.Println(n, err)
			if days <= 0 {
				days = 1
			}
			start = start.Add(-time.Duration(days-1) * 24 * time.Hour)

		}

	}
	return start, end, days
}

func detailWithDurationByDate(durationByDate map[string]time.Duration) (string, float64) {

	var startDate, endDate time.Time
	var days int
	for dateString := range durationByDate {
		date, _ := time.Parse("2006-01-02", dateString)
		if startDate.IsZero() || date.Before(startDate) {
			startDate = date
		}
		if endDate.IsZero() || date.After(endDate) {
			endDate = date
		}
		days++
	}

	var total time.Duration
	for _, duration := range durationByDate {
		total += duration
	}

	average := total.Hours() / float64(days)

	var s string
	summary := fmt.Sprintf("summary:  total day: %d, total hour: %.0fh, total min: %.fmin, average: %.2fh\n\n", days, total.Hours(), total.Minutes(), average)
	s += summary
	for i := 0; i < days; i++ {
		date := endDate.AddDate(0, 0, -i)
		dateStr := date.Format("2006-01-02")
		dateWeekday := date.Weekday().String()
		totalDuration := durationByDate[dateStr]
		totalMinutes := totalDuration.Minutes()
		totalHours := totalMinutes / 60
		s += fmt.Sprintf("\n%s, %s, ", dateStr, dateWeekday)

		empty := "🈳🈳🈳\n"
		if totalMinutes == 0 {
			s += empty
		} else {
			s += fmt.Sprintf("%.fmin; %.2fh\n", totalMinutes, totalHours)
		}
	}

	return s, average
}
