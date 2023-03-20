package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/viper"
)

func main() {
	viper.AutomaticEnv()

	chatIDs, err := strconv.ParseInt(strings.Split(viper.GetString("CHATID"), ","), 10, 64)
	if err != nil {
		fmt.Println("Error parsing CHAT_IDS:", err)
		os.Exit(1)
	}

	fmt.Println(chatIDs) // prints [123 456]
}
