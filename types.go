// all the required data structures are maintained here
package tracker

import (
	"fmt"
	"os"
	"time"

	env "github.com/caarlos0/env/v6"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

var Bot *tgbotapi.BotAPI

// bot constructor
func init() {
	bot, err := getTBot()
	Bot = bot
	if err != nil {
		fmt.Println("bot initialization failed")
	}
}

const (
	MYID           = 1346530914
	GROUPID        = -557832891
	URL            = "https://cdn-api.co-vin.in/"
	URLPATH        = "api/v2/appointment/sessions/public/findByDistrict"
	DISTQUERY      = "district_id"
	DISTID         = "391"
	PINQUERY       = "pincode"
	PINCODE        = "423601"
	DATEQUERY      = "date"
	HostAddress    = "8081"
	WaitTime       = float64((time.Minute * 5) / 100)
	MessageTimeout = 30
)

const (
	AvailableCapacity = "Available Capacity: "
	MinAge            = "Minimum Age: "
	Vaccine           = "Vaccine: "
	Name              = "Name: "
	Available         = "Available"
	Session           = "Session"
	Slot              = "Slot"
	SessionCount      = "SessionCount"
	SlotCount         = "SlotCount"
)

type Meta struct {
	Center []Centers `json:"centers,omitempty"`
}

type Centers struct {
	Name     string     `json:"name,omitempty"`
	Address  string     `json:"address,omitempty"`
	PinCode  int        `json:"pincode,omitempty"`
	From     string     `json:"from,omitempty"`
	To       string     `json:"to,omitempty"`
	Capacity int        `json:"available_capacity,omitempty"`
	AgeLimit int        `json:"min_age_limit,omitempty"`
	Session  []Sessions `json:"sessions"`
}

type Sessions struct {
	Date              string   `json:"date,omitempty"`
	AvailableCapacity int      `json:"available_capacity,omitempty"`
	MinAge            int      `json:"min_age_limit,omitempty"`
	Vaccine           string   `json:"vaccine,omitempty"`
	Slots             []string `json:"slots,omitempty"`
}

type Needed struct {
	NumberOfSessions  int
	NumberOfSlots     int
	Name              string
	AvailableCapacity int
	MinAge            int
	Vaccine           string
	Slots             []string
}

type BotConfig struct {
	Token string `env:"TOKEN"`
}

func init() {
	cfg := BotConfig{}
	if err := env.Parse(&cfg); err != nil {
		fmt.Printf("%+v\n", err)
	}
}

var BotToken = os.Getenv("TOKEN")
var FinalMsg map[string]map[string]string

var (
	StopFlag = false
	SkipFlag = false
	Date     = -1
	SentOnce = false
)

type SlotInfo struct {
	Sessions []DistSessions `json:"sessions"`
}

type DistSessions struct {
	Name                   string `json:"name"`
	Address                string `json:"address"`
	Pincode                int    `json:"pincode"`
	FeeType                string `json:"fee_type"`
	Fee                    string `json:"fee"`
	Date                   string `json:"date"`
	AvailableCapacity      int    `json:"available_capacity"`
	AvailableCapacityDose1 int    `json:"available_capacity_dose1"`
	AvailableCapacityDose2 int    `json:"available_capacity_dose2"`
	MinAgeLimit            int    `json:"min_age_limit"`
	Vaccine                string `json:"vaccine"`
}
