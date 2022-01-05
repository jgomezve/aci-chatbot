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
type Callback func(a *apic.ApicClient) string

type Command struct {
	help     string
	callback Callback
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
	bot.addCommand("/cpu", "APIC CPU Information", cpuCommand)
	bot.addCommand("/help", "Chatbot Help", helpCommand(bot.commands))
	bot.setupWebhook()
	bot.routes()
	return bot
}

// Command Handlers
func cpuCommand(c *apic.ApicClient) string {
	res := ""
	for _, item := range c.GetProcEntity() {
		res = res + "\n- **Proc** " + item.Dn + "\t💻 **CPU**: " + item.CpuPct + "\t💾 **Memory**: " + item.MemFree
	}
	return fmt.Sprintf("Hi 🤖 !%s", res)
}

func helpCommand(cmd map[string]Command) Callback {
	return func(a *apic.ApicClient) string {
		help := "Hello Sir, How can I help you?\n\n"
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
func webhookHandler(wbx *webex.WebexClient, a *apic.ApicClient, cmd map[string]Command) http.HandlerFunc {
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
		if messages[0].PersonEmail != "cx-germany-bot@webex.bot" {
			for key, element := range cmd {
				if messages[0].Text == key {
					wbx.SendMessageToRoom(element.callback(a), payload.Data.RoomId)
				}
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
func (b *Bot) addCommand(cmd string, h string, call Callback) {

	b.commands[cmd] = Command{
		help:     h,
		callback: call,
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
