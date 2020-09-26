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

func request(botChatID int64, termID string, subjectID string, courseID string, bot *tgbotapi.BotAPI) {
	var counter int

	seatFound := false

	// 1205 Spring 2020
	// 1209 Fall 2020
	// 1211 Winter 2021

	httpBase := "http://www.adm.uwaterloo.ca/cgi-bin/cgiwrap/infocour/salook.pl?level=under&"
	httpStr := httpBase + "sess=" + termID + "&subject=" + subjectID + "&cournum=" + courseID
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

		if len(lines) == 0 {
			message := fmt.Sprintf("Unable to process info on " + httpStr)
			fmt.Printf(message)
			msg := tgbotapi.NewMessage(botChatID, message)
			bot.Send(msg)
			return
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
			var message string
			if enrollCap > enrollTot {
				message = fmt.Sprintf("Seat available at section %d for %s\n", index, courseID)
				seatFound = true // set flag to exit
			} else if counter%720 == 0 {
				message = fmt.Sprintf("Seat not available at section %d for %s\n", index, courseID)
			} else {
				continue
			}
			message += "Cap: " + strconv.Itoa(enrollCap) + " Total: " + strconv.Itoa(enrollTot)
			message += " Avaiable: " + strconv.Itoa(enrollCap-enrollTot)
			fmt.Printf(message)
			msg := tgbotapi.NewMessage(botChatID, message)
			bot.Send(msg)
		}

		counter++

		if counter%720 == 0 {
			// sends a message after an hour has passed
			message := fmt.Sprintf("One hour has passed since reading on %s:) \n", courseID)
			fmt.Printf(message)
			msg := tgbotapi.NewMessage(botChatID, message)
			bot.Send(msg)
			counter = 0 // reset counter
		}

		if seatFound { // exit
			message := fmt.Sprintf("Search for %s completed\n", courseID)
			fmt.Printf(message)
			msg := tgbotapi.NewMessage(botChatID, message)
			bot.Send(msg)
			return
		}

		time.Sleep(5000 * time.Millisecond)
	}
}

func main() {

	bot, err := tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_APITOKEN"))
	if err != nil {
		panic(err) // You should add better error handling than this!
	}

	bot.Debug = false // Controls if the library display every request and response.

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	welcomeMsg := getWelcomeMsg()

	updates, err := bot.GetUpdatesChan(u)

	for update := range updates {

		if update.Message == nil { // ignore any non-Message Updates
			continue
		} else {

			botChatID := update.Message.Chat.ID
			textReceived := update.Message.Text
			var responseMsg string

			if update.Message.IsCommand() { // handles command
				responseMsg = welcomeMsg

			} else { // handles message
				log.Printf("[%s] %s", update.Message.From.UserName, textReceived)

				termID, subjectID, courseID, err := parseInput(textReceived)

				if err != nil {
					fmt.Println(err)
					continue
				}
				term, _ := strconv.Atoi(termID)
				responseMsg = "Successfully set a watch on " + subjectID + courseID + " for term " + getTermInfo(term)

				go request(botChatID, termID, subjectID, courseID, bot)
			}
			msg := tgbotapi.NewMessage(botChatID, responseMsg)
			msg.ReplyToMessageID = update.Message.MessageID
			bot.Send(msg)
		}

	}

}

func getTermInfo(term int) string {
	termMsg := strconv.Itoa(term)
	switch term % 10 {
	case 1:
		termMsg += " Winter "
	case 5:
		termMsg += " Spring "
	case 9:
		termMsg += " Fall "
	default:
		termMsg += " Unspecified "
	}

	termMsg += "20" + strconv.Itoa((term/10)%100) + "\n"
	return termMsg
}

func getWelcomeMsg() string {
	year, month, _ := time.Now().Date()

	termBase := 1000 + (year%100)*10
	terms := make([]int, 2)

	if month < 5 {
		terms[0] = termBase + 1
		terms[1] = termBase + 5
	} else if month < 9 {
		terms[0] = termBase + 5
		terms[1] = termBase + 9
	} else {
		terms[0] = termBase + 9
		terms[1] = termBase + 11
	}

	welcomeMsg := "To get update on a course, please specify\n[term] [subject] [course number]\n\ne.g.\n1205 CS 486\n\nFor terms, \n"

	for _, term := range terms {
		welcomeMsg += getTermInfo(term)
	}
	return welcomeMsg
}

func parseInput(input string) (string, string, string, error) {
	re := regexp.MustCompile(`\d[\d,]*`)

	submatchall := re.FindAllString(input, -1)
	for _, element := range submatchall {
		fmt.Println(element)
	}

	subject := input[len(submatchall[0])+1 : len(input)-len(submatchall[1])-1]
	fmt.Println(subject)
	return submatchall[0], subject, submatchall[1], nil
}
