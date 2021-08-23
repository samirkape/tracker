// main package for getting vaccine slot availability information from cowin.org and if more than one slots are available to book, then sending message to appropriate user by a means of telegram bot.
package tracker

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/tidwall/buntdb"
)

// driver function
func Track() {
	db := getDB()
	defer db.Close()
	for {
		data, err := slotInfoProc()
		if err != nil {
			fmt.Println(err)
			time.Sleep(time.Duration(WaitTime)) // follow 100 requests per 5 minutes limit by cowin.gov
			continue
		}
		if reflect.ValueOf(data).IsZero() {
			log.Printf("no vaccines available for %s", getDate())
		} else {
			go filterData(data, db) // discard unnecessary data
		}
		time.Sleep(time.Duration(WaitTime)) // follow 100 requests per 5 minutes limit by cowin.gov
	}
}

func getDB() *buntdb.DB {
	db, err := buntdb.Open("msg.db")
	if err != nil {
		log.Fatal(err)
	}
	return db
}

// 1. Construct and parse url
// 2. Fetch by http.Get + response JSON decode
// 3. Sending msg to bot on meeting certain conditions
func slotInfoProc() (SlotInfo, error) {
	gsi_err := "slotInfoProc: failed to get vaccine info %v"
	url, err := buildQuery() // parse url
	if err != nil {
		return SlotInfo{}, fmt.Errorf(gsi_err, err)
	}
	log.Printf("query building successful %s\n", url)
	data, err := fetchURL(url) // http get on cowin api + decode json
	if err != nil {
		return SlotInfo{}, fmt.Errorf(gsi_err, err)
	}
	log.Println("http get and response decode successful")
	return data, nil
}

func filterData(data SlotInfo, db *buntdb.DB) {
	for _, session := range data.Sessions {
		// Poll for Dose1 and for age below 45
		if session.AvailableCapacityDose1 > 1 || session.AvailableCapacityDose2 > 1 {
			err := db.View(func(tx *buntdb.Tx) error {
				val, err := tx.Get(session.Name)
				if err != nil {
					return err
				}
				log.Println("key is already there, wait for timeout to send the message\n", val)
				return nil
			})

			if err == buntdb.ErrNotFound {
				if session.FeeType == "Paid" {
					db.Update(func(tx *buntdb.Tx) error {
						tx.Set(session.Name, "", &buntdb.SetOptions{Expires: true, TTL: time.Hour * MessageTimeout})
						return nil
					})
				} else {
					db.Update(func(tx *buntdb.Tx) error {
						tx.Set(session.Name, "", &buntdb.SetOptions{Expires: true, TTL: time.Hour * MessageTimeout})
						return nil
					})
				}
				t := time.Now()
				if t.Hour() < 18 && t.Hour() > 7 {
					msg := createMessage(session)
					SendMessage(msg, MYID)
				}
			}
		}
	}
}

// simple http.Get + decode json response
func fetchURLv1(url string) (Meta, error) {
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
func fetchURL(url string) (SlotInfo, error) {
	CoMeta := &SlotInfo{}
	client := &http.Client{}
	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return *CoMeta, fmt.Errorf("fetchURL: unable to fetch url: %v", err)
	}
	// faking browser
	request.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.1 Safari/605.1.15")
	resp, err := client.Do(request)

	if resp == nil {
		return *CoMeta, err
	}

	if err != nil || resp.StatusCode != http.StatusOK {
		log.Println(err, resp.StatusCode)
		return *CoMeta, fmt.Errorf("fetchURL: unable to fetch url: %v", err)
	}

	defer resp.Body.Close()
	err = json.NewDecoder(resp.Body).Decode(&CoMeta)

	if err != nil {
		return *CoMeta, fmt.Errorf("fetchURL: unable to decode json response %v", err)
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
func getDate() string {
	year, month, day := time.Now().Date()
	t := time.Now()
	if t.Hour() > 15 {
		day++
	}
	return fmt.Sprintf(strconv.Itoa(day) + "-" + strconv.Itoa(int(month)) + "-" + strconv.Itoa(year))
}

// form query by paramters
func buildQuery() (string, error) {
	date := getDate()
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

func createMessage(data DistSessions) string {
	msg := []string{
		"\nName: %s\n",
		"Pincode:  %d\n",
		"Type:  %s\n",
		"Fee:  %s\n",
		"Date:  %s\n",
		"Dose1:  *%d*\n",
		"Age Limit:  %d\n",
		"Vaccine:  *%s*\n",
		"Dose2:  *%d*\n",
		"Area:  %s\n",
	}

	var BuildSlot strings.Builder
	name := strings.Split(data.Name, " ")
	addr := strings.Split(data.Address, " ")

	if len(name) > 2 {
		BuildSlot.WriteString(fmt.Sprintf(msg[0], strings.Join(name[:len(name)/2], " ")))
		BuildSlot.WriteString(strings.Join(name[len(name)/2:], " "))
		BuildSlot.WriteString("\n\n")
	} else {
		BuildSlot.WriteString(fmt.Sprintf(msg[0], data.Name))
		BuildSlot.WriteString("\n")
	}

	if len(addr) > int(2) {
		BuildSlot.WriteString(fmt.Sprintf(msg[9], strings.Join(addr[:len(addr)/2], " ")))
		BuildSlot.WriteString(strings.Join(addr[len(addr)/2:], " "))
		BuildSlot.WriteString("\n\n")
	} else {
		BuildSlot.WriteString(fmt.Sprintf(msg[9], data.Address))
	}

	BuildSlot.WriteString(fmt.Sprintf(msg[1], data.Pincode))
	BuildSlot.WriteString(fmt.Sprintf(msg[2], data.FeeType))
	if data.FeeType != "Free" {
		BuildSlot.WriteString(fmt.Sprintf(msg[3], data.Fee))
	}
	BuildSlot.WriteString(fmt.Sprintf(msg[4], data.Date))
	BuildSlot.WriteString(fmt.Sprintf(msg[6], data.MinAgeLimit))
	BuildSlot.WriteString(fmt.Sprintf(msg[7], data.Vaccine))

	BuildSlot.WriteString("\n")
	BuildSlot.WriteString(fmt.Sprintf(msg[5], data.AvailableCapacityDose1))
	if data.AvailableCapacityDose2 > 1 {
		BuildSlot.WriteString(fmt.Sprintf(msg[8], data.AvailableCapacityDose2))
	}
	return BuildSlot.String()
}
