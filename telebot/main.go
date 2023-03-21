package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/atony2099/go_project/telebot/bot"
	"github.com/atony2099/go_project/telebot/db"
	"github.com/atony2099/go_project/telebot/elapse"
	"github.com/go-co-op/gocron"
	"github.com/spf13/viper"
)

var TodayPassRemainInterva = 60
var NoStudyrRemainInterva = 30

func main() {

	viper.AutomaticEnv()
	token := viper.GetString("TOKEN")
	dsn := viper.GetString("DB")
	chatid := viper.GetInt64("CHATID")

	if token == "" || dsn == "" || chatid == 0 {
		log.Fatalf("env error:token: %s, dsn: %s, chatid: %d ", token, dsn, chatid)
	}

	db.Init(dsn)
	bot.NewBot(token, chatid)
	go bot.HandleCommand()
	go doCron()

	s := make(chan os.Signal, 1)
	signal.Notify(s, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	fmt.Println("start success")
	<-s
	fmt.Println("clean")

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
