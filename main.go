package main

import (
	"chatbot/apic"
	"chatbot/bot"
	"chatbot/webex"
	"errors"
	"os"
)

type Requirements struct {
	webexToken string
	botUrl     string
	apicUrl    string
	apicUsr    string
	apicPsw    string
}

func checkRequirements() (*Requirements, error) {
	r := Requirements{}
	if os.Getenv("WEBEX_TOKEN") == "" {
		return nil, errors.New("WEBEX_TOKEN not set")
	}
	r.webexToken = os.Getenv("WEBEX_TOKEN")
	if os.Getenv("BOT_URL") == "" {
		return nil, errors.New("BOT_URL not set")
	}
	r.botUrl = os.Getenv("BOT_URL")
	if os.Getenv("APIC_URL") == "" {
		return nil, errors.New("APIC_URL not set")
	}
	r.apicUrl = os.Getenv("APIC_URL")
	if os.Getenv("APIC_USERNAME") == "" {
		return nil, errors.New("APIC_USERNAME not set")
	}
	r.apicUsr = os.Getenv("APIC_USERNAME")
	if os.Getenv("APIC_PASSWORD") == "" {
		return nil, errors.New("APIC_PASSWORD not set")
	}
	r.apicPsw = os.Getenv("APIC_PASSWORD")
	return &r, nil
}

func main() {

	// Check requirements
	r, err := checkRequirements()
	if err != nil {
		panic(err)
	}
	// Set up Webex Client
	wbx := webex.NewWebexClient(r.webexToken)
	// Set up APIC Client
	apic, err := apic.NewApicClient(r.apicUrl, r.apicUsr, r.apicPsw, apic.SetTimeout(5))
	if err != nil {
		panic("APIC connection failed")
	}
	// Configure and start Bot server
	bot := bot.NewBot(&wbx, &apic, r.botUrl)
	if err = bot.Start(":7001"); err != nil {
		panic("Bot failed to start")
	}
}
