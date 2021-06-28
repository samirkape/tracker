// main package for getting vaccine slot availability information from cowin.org and if more than one slots are available to book, then sending message to appropriate user by a means of telegram bot.
package tracker

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// driver function
func Track() {
	for {
		err := GetSlotInfo()
		if err != nil {
			fmt.Println(err)
		}
		time.Sleep(time.Duration(WaitTime)) // follow 100 requests per 5 minutes limit by cowin.gov
	}
}

// 1. construct and parse url
// 2. fetch by http.Get + json decode
// 3. sending msg to bot
func GetSlotInfo() error {
	gsi_err := "GetSessionInfo: failed to get vaccine info %v"
	url, err := BuildQuery() // parse url
	if err != nil {
		return fmt.Errorf(gsi_err, err)
	}
	data, err := FetchV3(url) // fetch url + decode json
	if err != nil {
		return fmt.Errorf(gsi_err, err)
	}
	if reflect.ValueOf(data).IsZero() {
		return fmt.Errorf("no vaccines available for %s", GetDate())
	} else {
		FilterDist(data) // discard unnecessary data
		// err := MessageHandler(msg, ) // send msg if the vaccines are available to book
		if err != nil {
			return fmt.Errorf(gsi_err, err)
		}
	}
	return nil
}

// initialize and validate bot
func getTBot() (*tgbotapi.BotAPI, error) {
	BotToken := os.Getenv("TOKEN")
	if len(BotToken) == 0 {
		return nil, errors.New("getTBot: could not find bot token")
	}
	bot, err := tgbotapi.NewBotAPI(BotToken)
	//bot.Debug = true
	if err != nil {
		return nil, fmt.Errorf("getTBot: error initializing bot: %v", err)
	}
	return bot, err
}

// sends message to registered id
func SendMessage(Info string) error {
	if StopFlag {
		fmt.Println("warning! bot has recieved an ack to stop")
		return nil
	}
	msg := tgbotapi.NewMessage(SAMIRCID, Info)
	url := tgbotapi.NewMessage(SAMIRCID, "https://selfregistration.cowin.gov.in")
	//msg1 := tgbotapi.NewMessage(GROUPID, Info)
	msg.ParseMode = "markdown"
	_, err := Bot.Send(msg)
	Bot.Send(url)
	//_, err = bot.Send(msg1)
	if err != nil {
		return fmt.Errorf("sendmessage: message sending failed: %v", err)
	}
	time.Sleep(30 * time.Second)
	return nil
}

// // checks for any msg from bot
// func ACK(bot *tgbotapi.BotAPI) {
// 	if StopFlag {
// 		return
// 	}
// 	u := tgbotapi.NewUpdate(0)
// 	u.Timeout = 60
// 	updates, err := bot.GetUpdatesChan(u)
// 	if err != nil {
// 		StopFlag = false
// 	}
// 	for update := range updates {
// 		msg := update.Message.Text
// 		if msg == "skip" {
// 			StopFlag = true
// 		} else if msg == "stop" {
// 			StopFlag = true
// 		} else if strings.HasPrefix(msg, "Date=") {
// 			Date, err = strconv.Atoi(msg[5:])
// 		}
// 	}
// }

// custom http.Get for changing header + decode json response
func FetchV3(url string) (SlotInfo, error) {
	CoMeta := &SlotInfo{}
	client := &http.Client{}
	request, err := http.NewRequest("GET", url, nil)
	log.Println("https://selfregistration.cowin.gov.in")
	if err != nil {
		return *CoMeta, fmt.Errorf("fetch v2: unable to fetch url: %v", err)
	}

	// faking browser
	request.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.1 Safari/605.1.15")
	resp, err := client.Do(request)

	if resp == nil {
		return *CoMeta, err
	}

	if err != nil || resp.StatusCode != http.StatusOK {
		log.Println(err, resp.StatusCode)
		return *CoMeta, fmt.Errorf("fetch v2: unable to fetch url: %v", err)
	}

	defer resp.Body.Close()
	err = json.NewDecoder(resp.Body).Decode(&CoMeta)

	if err != nil {
		return *CoMeta, fmt.Errorf("fetch v2: unable to decode json response %v", err)
	}
	return *CoMeta, nil
}

func FilterDist(data SlotInfo) {
	if SentOnce {
		time.Sleep(30 * time.Second)
		SentOnce = false
	} else {
		for _, session := range data.Sessions {
			c := session.AvailableCapacityDose2 > 0 && session.MinAgeLimit == 18
			if c {
				msg := CreateMessage(session)
				go SendMessage(msg)
				SentOnce = true
			}
		}
	}
}

func CreateMessage(data DistSessions) string {
	msg := []string{"Name: %s\n",
		"Address: %s\n",
		"Pincode: %d\n",
		"FeeType: %s\n",
		"Fee: %s\n",
		"Date: %s\n",
		"AvailableCapacity: %d\n",
		"AvailableCapacityDose1: %d\n",
		"AvailableCapacityDose2: %d\n",
		"MinAgeLimit: %d\n",
		"Vaccine: %s\n"}
	var BuildSlot strings.Builder
	BuildSlot.WriteString(fmt.Sprintf(msg[0], data.Name))
	BuildSlot.WriteString(fmt.Sprintf(msg[1], data.Address))
	BuildSlot.WriteString(fmt.Sprintf(msg[2], data.Pincode))
	BuildSlot.WriteString(fmt.Sprintf(msg[3], data.FeeType))
	BuildSlot.WriteString(fmt.Sprintf(msg[4], data.Fee))
	BuildSlot.WriteString(fmt.Sprintf(msg[5], data.Date))
	BuildSlot.WriteString(fmt.Sprintf(msg[6], data.AvailableCapacity))
	BuildSlot.WriteString(fmt.Sprintf(msg[7], data.AvailableCapacityDose1))
	BuildSlot.WriteString(fmt.Sprintf(msg[8], data.AvailableCapacityDose2))
	BuildSlot.WriteString(fmt.Sprintf(msg[9], data.MinAgeLimit))
	BuildSlot.WriteString(fmt.Sprintf(msg[10], data.Vaccine))
	return BuildSlot.String()
}

// simple http.Get + decode json response
func Fetch(url string) (Meta, error) {
	var CoMeta Meta
	resp, err := http.Get(url)
	if err != nil || resp.StatusCode != http.StatusOK {
		log.Println(err, resp.StatusCode)
		return Meta{}, fmt.Errorf("fetch v1: unable to fetch URL: %v", err)
	}
	defer resp.Body.Close()
	err = json.NewDecoder(resp.Body).Decode(&CoMeta)
	if err != nil || resp.StatusCode != http.StatusOK {
		log.Println(err, resp.StatusCode)
		return Meta{}, fmt.Errorf("fetch v1: unable to decode json response: %v", err)

	}
	return CoMeta, nil
}

// custom http.Get for changing header + decode json response
func FetchV2(url string) (SlotInfo, error) {
	CoMeta := &SlotInfo{}
	client := &http.Client{}
	request, err := http.NewRequest("GET", url, nil)
	log.Println("https://selfregistration.cowin.gov.in")
	if err != nil {
		return *CoMeta, fmt.Errorf("fetch v2: unable to fetch url: %v", err)
	}

	// faking browser
	request.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.1 Safari/605.1.15")
	resp, err := client.Do(request)

	if resp == nil {
		return *CoMeta, err
	}

	if err != nil || resp.StatusCode != http.StatusOK {
		log.Println(err, resp.StatusCode)
		return *CoMeta, fmt.Errorf("fetch v2: unable to fetch url: %v", err)
	}

	defer resp.Body.Close()
	err = json.NewDecoder(resp.Body).Decode(&CoMeta)

	if err != nil {
		return *CoMeta, fmt.Errorf("fetch v2: unable to decode json response %v", err)
	}
	return *CoMeta, nil
}

// dummy input for testing
func dummyJson() *bytes.Reader {
	input := []byte(`{"centers": [{"center_id": 116271, "name": "Sanvatsar PHC", "address": "Sanvatsar", "state_name": "Maharashtra", "district_name": "Ahmednagar", "block_name": "Kopargaon", "pincode": 423601, "lat": 19, "long": 74, "from": "09:00:00", "to": "17:00:00", "fee_type": "Free", "sessions": [{"session_id": "ab670e27-4e05-487b-b282-bde3a8904061", "date": "08-05-2021", "available_capacity": 0, "min_age_limit": 45, "vaccine": "COVISHIELD", "slots": ["09:00AM-11:00AM", "11:00AM-01:00PM", "01:00PM-03:00PM", "03:00PM-05:00PM"]}]}]}`)
	r := bytes.NewReader(input)
	return r
}

// date needed for query
func GetDate() string {
	year, month, day := time.Now().Date()
	if Date != -1 {
		day = Date
	} else {
		day++
	}
	return fmt.Sprintf(strconv.Itoa(day) + "-" + strconv.Itoa(int(month)) + "-" + strconv.Itoa(year))
}

// form query by paramters
func BuildQuery() (string, error) {
	date := GetDate()
	base, err := url.Parse(URL)
	if err != nil {
		return "", fmt.Errorf("buildquery: unable to parse url: %v", err)
	}
	base.Path += URLPATH
	params := url.Values{}
	params.Add(DATEQUERY, date)
	params.Add(DISTQUERY, DISTID)
	base.RawQuery = params.Encode()
	url := base.String()
	return url, nil
}
