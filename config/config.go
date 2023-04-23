package config

import (
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	TelegramToken string
	DatabaseDSN   string
	ChatID        int64
	Port          string
}

func LoadEnv() {

	viper.AutomaticEnv()

}

func GetConfig() *Config {
	token := viper.GetString("TOKEN")
	dsn := viper.GetString("DB")
	chatid := viper.GetInt64("CHATID")
	port := viper.GetString("PORT")
	if port == "" {
		port = "8080"
	}

	if token == "" || dsn == "" || chatid == 0 {
		log.Fatalf("env error:token: %s, dsn: %s, chatid: %d ", token, dsn, chatid)
	}

	// log.Println("config loaded", token, dsn, chatid, port)

	return &Config{
		TelegramToken: token,
		DatabaseDSN:   dsn,
		ChatID:        chatid,
		Port:          port,
	}
}
