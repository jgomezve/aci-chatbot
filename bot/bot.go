package bot

import (
	"aci-chatbot/apic"
	"aci-chatbot/webex"
	"encoding/json"
	"fmt"
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

type webexInterface interface {
	SendMessageToRoom(m string, roomId string) error
	GetBotDetails() (webex.WebexPeople, error)
	GetWebHooks() ([]webex.WebexWebhook, error)
	DeleteWebhook(name, tUrl, id string) error
	CreateWebhook(name, url, resource, event string) error
	GetPersonInfromation(id string) (webex.WebexPeople, error)
	GetMessages(roomId string, max int) ([]webex.WebexMessage, error)
}

// Bot definition
type Bot struct {
	wbx      webexInterface
	apic     *apic.ApicClient
	server   *http.Server
	router   *http.ServeMux
	url      string
	commands map[string]Command
	info     webex.WebexPeople
}

// Bot Generator
func NewBot(wbx webexInterface, apic *apic.ApicClient, botUrl string) Bot {

	info, err := wbx.GetBotDetails()
	if err != nil {
		log.Printf("could not retrieve the bot information. Err %s", err)
	}
	bot := Bot{
		wbx:    wbx,
		apic:   apic,
		router: http.NewServeMux(),
		url:    botUrl,
		info:   info,
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

	return fmt.Sprintf("Hi %s 🤖 , here the details of ep `%s` %s", wm.sender, splitEpCommand(m.cmd)["mac"], res)
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
func testHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "I am alive!")
	w.WriteHeader(http.StatusOK)
}
func aboutMeHandler(wbx webexInterface) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		info, err := wbx.GetBotDetails()
		if err != nil {
			log.Printf("could not retrieve the bot information. Err %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		jp, err := json.Marshal(info)
		if err != nil {
			log.Printf("could not parse webex response. Error %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(jp)
		w.WriteHeader(http.StatusOK)
	})
}
func webhookHandler(wbx webexInterface, ap *apic.ApicClient, cmd map[string]Command, b webex.WebexPeople) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Parse incoming webhook. From which room does it come  from?
		wh := webex.WebexWebhook{}
		if err := parseWebHook(&wh, r); err != nil {
			log.Printf("failed to parse incoming webhook. Error %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		// Retrieve the last message, it should not have been written by the bot
		messages, err := wbx.GetMessages(wh.Data.RoomId, 1)
		if err != nil {
			log.Printf("failed trying to retreive the last message. Error %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		// Is the message send from someone who is not the bot
		if messages[0].PersonId != b.Id {
			// Get sender personal information
			sender, _ := wbx.GetPersonInfromation(messages[0].PersonId)
			found := false
			// Check which command was sent in the webex room
			for _, element := range cmd {
				if MatchCommand(messages[0].Text, element.regex) {
					// Send message back the text is returned from the commandHandler
					wbx.SendMessageToRoom(element.callback(ap, Message{cmd: messages[0].Text}, WebexMessage{sender: sender.NickName}), wh.Data.RoomId)
					found = true
					w.WriteHeader(http.StatusOK)
					return
				}
			}
			// If command sent does not match anything, send back the help menu
			if !found {
				wbx.SendMessageToRoom(cmd["/help"].callback(ap, Message{cmd: messages[0].Text}, WebexMessage{sender: sender.NickName}), wh.Data.RoomId)
				w.WriteHeader(http.StatusOK)
			}
		}
	})
}

// Bot Methods
func (b *Bot) routes() {
	// TODO: is this fine?
	b.router.HandleFunc("/about", aboutMeHandler(b.wbx))
	b.router.HandleFunc("/test", testHandler)
	b.router.HandleFunc("/webhook", webhookHandler(b.wbx, b.apic, b.commands, b.info))
}
func (b *Bot) addCommand(cmd string, h string, re string, call Callback) {
	// add item to the dispatch table
	b.commands[cmd] = Command{
		help:     h,
		callback: call,
		regex:    re,
	}
}
func (b *Bot) setupWebhook() {
	// TODO: Delete exsiting webhooks with the same name

	whs, _ := b.wbx.GetWebHooks()
	for _, wh := range whs {
		if wh.Name == b.info.DisplayName {
			b.wbx.DeleteWebhook(b.info.DisplayName, b.url+"/webhook", wh.Id)
		}
	}
	if err := b.wbx.CreateWebhook(b.info.DisplayName, b.url+"/webhook", "messages", "created"); err != nil {
		log.Printf("Getting current webhooks")
	}
}
func (b *Bot) Start(addr string) error {
	// Start the http server
	b.server = &http.Server{
		Addr:    addr,
		Handler: b.router,
	}
	log.Println("Starting Server")
	return b.server.ListenAndServe()
}
