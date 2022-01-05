package bot

import (
	"chatbot/apic"
	"chatbot/webex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

// Callback helpers
type Callback func(a *apic.ApicClient, m Message, wm WebexMessage) string

// struc to represent the incomming Webex message

type WebexMessage struct {
	sender string
}

// struct to represent the CLI command
type Message struct {
	cmd string
}

type Command struct {
	help     string
	callback Callback
	regex    string
}

// Bot definition
type Bot struct {
	wbx      *webex.WebexClient
	apic     *apic.ApicClient
	server   *http.Server
	router   *http.ServeMux
	url      string
	commands map[string]Command
}

// Bot Generator
func NewBot(wbx *webex.WebexClient, apic *apic.ApicClient, botUrl string) Bot {

	bot := Bot{
		wbx:    wbx,
		apic:   apic,
		router: http.NewServeMux(),
		url:    botUrl,
	}

	bot.commands = make(map[string]Command)
	bot.addCommand("/cpu", "Get APIC CPU Information", "\\/cpu", cpuCommand)
	bot.addCommand("/ep", "Get APIC Endpoint Information. Usage /ep <ep_mac>", "\\/ep ([[:xdigit:]]{2}[:.-]?){5}[[:xdigit:]]{2}$", endpointCommand)
	bot.addCommand("/help", "Chatbot Help", "\\/help", helpCommand(bot.commands))
	bot.setupWebhook()
	bot.routes()
	return bot
}

// Command Handlers
func endpointCommand(c *apic.ApicClient, m Message, wm WebexMessage) string {

	res := ""
	for _, item := range c.GetEnpoint(splitEpCommand(m.cmd)["mac"]) {
		res = res + "\n- **Bridge Domain**: `" + item["bdDn"] + "`"
	}
	if res == "" {
		res = "/n Sorry " + wm.sender + "... I could not find this endpoint **mac**: `" + splitEpCommand(m.cmd)["mac"] + "`"
	}

	return fmt.Sprintf("Hi %s 🤖 !%s", wm.sender, res)
}

func cpuCommand(c *apic.ApicClient, m Message, wm WebexMessage) string {
	res := ""
	for _, item := range c.GetProcEntity() {
		res = res + "\n- **Proc**: `" + item["dn"] + "`\t💻 **CPU**: " + item["cpuPct"] + "\t💾 **Memory**: " + item["memFree"]
	}
	return fmt.Sprintf("Hi %s 🤖 !%s", wm.sender, res)
}

func helpCommand(cmd map[string]Command) Callback {
	return func(a *apic.ApicClient, m Message, wm WebexMessage) string {
		help := fmt.Sprintf("Hello %s, How can I help you?\n\n", wm.sender)
		for key, value := range cmd {
			help = help + "\t" + key + "->" + value.help + "\n"
		}
		return help
	}
}

// Endpoint Handlers
func echoHandler(wbx *webex.WebexClient) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("Hi!, I can access this client %s", wbx.GetBaseUrl())
	})
}
func testHandler(wbx *webex.WebexClient) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wbx.SendMessageToRoom("Did you call me?", "Y2lzY29zcGFyazovL3VzL1JPT00vZjRmZWZjZDAtNjI3NS0xMWVjLThiMTQtMDEyYWYxZGQ1M2Vl")
	})
}
func webhookHandler(wbx *webex.WebexClient, ap *apic.ApicClient, cmd map[string]Command) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload webex.WebexWebhook
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Println("Error: ", err)
		}

		if err = json.Unmarshal(body, &payload); err != nil {
			log.Println("Error: ", err)
		}
		messages, _ := wbx.GetMessages(payload.Data.RoomId, 1)
		botInfo, _ := wbx.GetBotDetails()
		if messages[0].PersonId != botInfo.Id {
			sender, err := wbx.GetPersonInfromation(messages[0].PersonId)
			if err != nil {
				sender.NickName = "Joe Doe"
			}
			found := false
			for _, element := range cmd {
				if MatchCommand(messages[0].Text, element.regex) {
					wbx.SendMessageToRoom(element.callback(ap, Message{cmd: messages[0].Text}, WebexMessage{sender: sender.NickName}), payload.Data.RoomId)
					found = true
				}
			}
			if !found {
				wbx.SendMessageToRoom(cmd["/help"].callback(ap, Message{cmd: messages[0].Text}, WebexMessage{sender: sender.NickName}), payload.Data.RoomId)
			}
		}
	})
}

// Bot Methods
func (b *Bot) routes() {

	b.router.HandleFunc("/echo", echoHandler(b.wbx))
	b.router.HandleFunc("/test", testHandler(b.wbx))
	b.router.HandleFunc("/webhook", webhookHandler(b.wbx, b.apic, b.commands))
}
func (b *Bot) addCommand(cmd string, h string, re string, call Callback) {

	b.commands[cmd] = Command{
		help:     h,
		callback: call,
		regex:    re,
	}
}
func (b *Bot) setupWebhook() {
	// TODO: Handle error updating and deleting existing webhook
	b.wbx.CreateWebhook("the-webhook-1", b.url+"/webhook", "messages", "created")
}
func (b *Bot) Start(addr string) error {

	b.server = &http.Server{
		Addr:    addr,
		Handler: b.router,
	}

	return b.server.ListenAndServe()
}
