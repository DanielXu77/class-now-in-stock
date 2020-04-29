package main

import (
	"bufio"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"
)

func main() {

	for {
		resp, err := http.Get("http://www.adm.uwaterloo.ca/cgi-bin/cgiwrap/infocour/salook.pl?level=under&sess=1205&subject=CS&cournum=486")
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
			if numSlice[4] < numSlice[5] {
				fmt.Printf("Seat available at section %d\n", index)
			} else {
				fmt.Printf("No available seat at section %d\n", index)
			}
		}

		time.Sleep(5000 * time.Millisecond)
	}

}
