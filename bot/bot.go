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
type Callback func(a apic.ApicInterface, m Message, wm WebexMessage) string

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
	wbx      webex.WebexInterface
	apic     apic.ApicInterface
	server   *http.Server
	router   *http.ServeMux
	url      string
	commands map[string]Command
	info     webex.WebexPeople
}

// Bot Generator
func NewBot(wbx webex.WebexInterface, apic apic.ApicInterface, botUrl string) (Bot, error) {

	info, err := wbx.GetBotDetails()
	if err != nil {
		log.Printf("could not retrieve the bot information. Err %s", err)
		return Bot{}, err
	}
	bot := Bot{
		wbx:    wbx,
		apic:   apic,
		router: http.NewServeMux(),
		url:    botUrl,
		info:   info,
	}

	bot.commands = make(map[string]Command)
	log.Println("Adding `/cpu` command")
	bot.addCommand("/cpu", "Get APIC CPU Information", "\\/cpu", cpuCommand)
	log.Println("Adding `/ep` command")
	bot.addCommand("/ep", "Get APIC Endpoint Information. Usage /ep <ep_mac>", "\\/ep ([[:xdigit:]]{2}[:.-]?){5}[[:xdigit:]]{2}$", endpointCommand)
	log.Println("Adding `/help` command")
	bot.addCommand("/help", "Chatbot Help", "\\/help", helpCommand(bot.commands))
	log.Println("Setting up Webex Webhook")
	if err = bot.setupWebhook(); err != nil {
		log.Printf("could not setup the webhook. Err %s", err)
		return Bot{}, err
	}
	bot.routes()

	return bot, nil
}

// Command Handlers
// /ep <ep_mac> handler
func endpointCommand(c apic.ApicInterface, m Message, wm WebexMessage) string {

	res := ""
	for _, item := range c.GetEnpoint(splitEpCommand(m.cmd)["mac"]) {
		res = res + "\n- **Bridge Domain**: `" + item["bdDn"] + "`"
	}
	if res == "" {
		res = "/n Sorry " + wm.sender + "... I could not find this endpoint **mac**: `" + splitEpCommand(m.cmd)["mac"] + "`"
	}

	return fmt.Sprintf("Hi %s 🤖 , here the details of ep `%s` %s", wm.sender, splitEpCommand(m.cmd)["mac"], res)
}

// /cpu handler
func cpuCommand(c apic.ApicInterface, m Message, wm WebexMessage) string {
	res := ""
	cpu, err := c.GetProcEntity()

	if err != nil {
		log.Printf("Error while connecting to the Apic. Err: %s", err)
		return fmt.Sprintf("Hi %s 🤖 !. I could not reach the APIC... Are there any issues?", wm.sender)

	}
	for _, item := range cpu {
		res = res + "\n- **Proc**: `" + item["dn"] + "`\t💻 **CPU**: " + item["cpuPct"] + "\t💾 **Memory**: " + item["memFree"]
	}
	return fmt.Sprintf("Hi %s 🤖 !%s", wm.sender, res)
}

// /help handler
func helpCommand(cmd map[string]Command) Callback {
	return func(a apic.ApicInterface, m Message, wm WebexMessage) string {
		help := fmt.Sprintf("Hello %s, How can I help you?\n\n", wm.sender)
		for key, value := range cmd {
			help = help + "\t" + key + "->" + value.help + "\n"
		}
		return help
	}
}

// Endpoint Handlers
// /test handler
func testHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "I am alive!")
	w.WriteHeader(http.StatusOK)
}

// /about handler
func aboutMeHandler(wbx webex.WebexInterface) http.HandlerFunc {
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

// /webhook handler
// TODO: Separate by method (GET, POST, PUT)
func webhookHandler(wbx webex.WebexInterface, ap apic.ApicInterface, cmd map[string]Command, b webex.WebexPeople) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Receiving %s on /webhook URI", r.Method)
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
			sender, _ := wbx.GetPersonInformation(messages[0].PersonId)
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
				return
			}
		}
		// To differentiate Webhooks triggered from the Bot
		w.WriteHeader(http.StatusAccepted)
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
func (b *Bot) setupWebhook() error {
	// TODO: Delete exsiting webhooks with the same name

	whs, _ := b.wbx.GetWebHooks()
	for _, wh := range whs {
		if wh.Name == b.info.DisplayName {
			log.Printf("Bot already has a Webhook with name %s\n", b.info.DisplayName)
			err := b.wbx.DeleteWebhook(b.info.DisplayName, b.url+"/webhook", wh.Id)
			log.Printf("Deleting Webhooks %s\n", b.info.DisplayName)
			if err != nil {
				log.Printf("could not delete existing webhook. Err %s", err)
				return err
			}
		}
	}
	log.Printf("Creating brand new Webhook Name: %s - URL: %s\n", b.info.DisplayName, b.url)
	if err := b.wbx.CreateWebhook(b.info.DisplayName, b.url+"/webhook", "messages", "created"); err != nil {
		log.Printf("could not create brand new webhook. Err %s", err)
		return err
	}
	return nil
}
func (b *Bot) Start(addr string) error {
	// Start the http server
	b.server = &http.Server{
		Addr:    addr,
		Handler: b.router,
	}
	log.Printf("Starting Server on address %s\n", addr)
	return b.server.ListenAndServe()
}
