package main

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func main() {

	bot, err := tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_APITOKEN"))
	if err != nil {
		panic(err) // You should add better error handling than this!
	}

	bot.Debug = false // Controls if the library display every request and response.

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	var botChatID int64

	var courseID string

	for update := range updates {
		if update.Message == nil { // ignore any non-Message Updates
			continue
		}

		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

		botChatID = update.Message.Chat.ID

		courseID = update.Message.Text

		if err != nil {
			fmt.Println(err)
			continue
		}

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "successfully linked to "+courseID)
		msg.ReplyToMessageID = update.Message.MessageID

		bot.Send(msg)
		break
	}

	var counter int
	httpStr := "http://www.adm.uwaterloo.ca/cgi-bin/cgiwrap/infocour/salook.pl?level=under&sess=1205&subject=CS&cournum=" + courseID

	fmt.Printf(httpStr + "\n")

	for {

		resp, err := http.Get(httpStr)
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()

		fmt.Println("Response status:", resp.Status)

		var lines []string

		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			tempString := scanner.Text()
			if strings.Contains(tempString, "LEC") {
				lines = append(lines, tempString)
			}
		}
		if err := scanner.Err(); err != nil {
			fmt.Println("Error in reading response body.")
			panic(err)
		}

		for index, val := range lines {
			re := regexp.MustCompile("[0-9]+")
			numSlice := re.FindAllString(val, -1)
			enrollCap, err := strconv.Atoi(numSlice[4])
			if err != nil {
				fmt.Println("Error reading enrollment cap")
			}
			enrollTot, err := strconv.Atoi(numSlice[5])
			if err != nil {
				fmt.Println("Error reading enrollment total")
			}
			if enrollCap > enrollTot {
				message := fmt.Sprintf("Seat available at section %d\n", index)
				fmt.Printf(message)
				msg := tgbotapi.NewMessage(botChatID, message)
				bot.Send(msg)
			}
		}

		if counter%720 == 0 {
			// sends a message after an hour has passed
			message := fmt.Sprintf("One hour has passed :)")
			fmt.Printf(message)
			msg := tgbotapi.NewMessage(botChatID, message)
			bot.Send(msg)
			counter = 0 // reset counter
		}

		counter++

		time.Sleep(5000 * time.Millisecond)
	}

}
