package bot

import (
	"aci-chatbot/apic"
	"aci-chatbot/webex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"sort"
	"strconv"
)

// Callback helpers
type Callback func(a apic.ApicInterface, m Message, wm WebexMessage) string

// Struct to represent the incomming Webex message
type WebexMessage struct {
	sender string
	roomId string
}

// Struct to represent the CLI command
type Message struct {
	cmd string
}

// Struct to save the suported CLI commands
type Command struct {
	help     string
	callback Callback
	suffix   string
	regex    string
}

// Bot definition
type Bot struct {
	wbx      webex.WebexInterface
	apic     apic.ApicInterface
	wsck     *apic.ApicWebSocket
	server   *http.Server
	router   *http.ServeMux
	url      string
	commands map[string]Command
	wsSubs   *webSocketDb
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
	bot.wsSubs = NewWsDb()

	log.Println("Adding `/info` command")
	bot.addCommand("/info", "Get Fabric Information ℹ️", "\\/info", "$", infoCommand)
	log.Println("Adding `/cpu` command")
	bot.addCommand("/cpu", "Get APIC CPU Information 💾", "\\/cpu", "$", cpuCommand)
	log.Println("Adding `/ep` command")
	bot.addCommand("/ep", "Get APIC Endpoint Information 💻. Usage <code>/ep [ep_mac] </code>", "\\/ep", " ([[:xdigit:]]{2}[:.-]?){5}[[:xdigit:]]{2}$", endpointCommand)
	log.Println("Adding `/neigh` command")
	bot.addCommand("/neigh", "Get Fabric Topology Information 🔢. Usage <code>/neigh [node_id] </code>", "\\/neigh", "( )?([0-9]{1,4})?$", neighCommand)
	log.Println("Adding `/faults` command")
	bot.addCommand("/faults", "Get Fabric latest faults ⚠️. Usage <code>/faults [count(1-10):opt] </code>", "\\/faults", "( )?([1-9]|10)( )?$", faultCommand)
	log.Println("Adding `/events` command")
	bot.addCommand("/events", "Get Fabric latest events ❎.   Usage <code>/events [user:opt] [count(1-10):opt] </code>", "\\/events", "( )?([A-Za-z]{5,10})?( )?([1-9]|10)?$", eventCommand)
	bot.addCommand("/websocket", "Subscribe to Fabric events 📩", "\\/websocket", " [A-Za-z]{1,20}( )?(rm)?$", websocketCommand(bot.wsSubs))
	log.Println("Adding `/help` command")
	log.Println("Adding `/help` command")
	bot.addCommand("/help", "Chatbot Help ❔", "\\/help", "$", helpCommand(bot.commands))
	log.Println("Setting up Webex Webhook")
	if err = bot.setupWebhook(); err != nil {
		log.Printf("could not setup the webhook. Err %s", err)
		return Bot{}, err
	}
	bot.routes()
	return bot, nil
}

// Command Handlers
// /websocket handler
func websocketCommand(wsDb *webSocketDb) Callback {
	return func(c apic.ApicInterface, m Message, wm WebexMessage) string {
		class := splitWebsocketCommand(m.cmd)["class"]
		operation := splitWebsocketCommand(m.cmd)["op"]

		// /websocket list -> Return the list of subscriptions
		if class == "list" {
			res := "<ul>"
			classes := wsDb.getClassesbyRoomId(wm.roomId)
			for _, class := range classes {
				res += fmt.Sprintf("<li><code>%s</code></li>", class)
			}
			res += "</ul>"
			if len(classes) == 0 {
				return fmt.Sprintf("Hi %s 🤖 !\n You are no subscribed to any class", wm.sender)
			} else {
				return fmt.Sprintf("Hi %s 🤖 !\n Here the list of subcribed classes:\n %s", wm.sender, res)
			}
		}
		// /websocket xxxx rm -> Remove subscirption to this Room
		if operation == "rm" {
			if wsDb.checkSubsciption(class, wm.roomId) {
				wsDb.removeSubcription(class, wm.roomId)
				return fmt.Sprintf("Hi %s 🤖 !\n\n Websocket subscription to MO/Class <code>%s</code> deleted 🔧 !", wm.sender, class)
			} else {
				return fmt.Sprintf("Hi %s 🤖 !\n\n You are not subscribed to MO/Class <code>%s</code>", wm.sender, class)
			}
			// /websocket xxxx -> Add subscirption to this Room
		} else {
			if !wsDb.checkSubsciption(class, wm.roomId) {
				id, err := c.SubscribeClassWebSocket(class)
				if err != nil {
					return fmt.Sprintf("Hi %s 🤖 !\n Sorry... I could not subscribe to the class <code>%s</code>", wm.sender, class)
				}
				wsDb.addSubcription(class, id, wm.roomId)
				return fmt.Sprintf("Hi %s 🤖 !\n\n Websocket subscription to MO/Class <code>%s</code> configured 🔧 !", wm.sender, class)
			} else {
				return fmt.Sprintf("Hi %s 🤖 !\n\n You are already subscribed to MO/Class <code>%s</code>", wm.sender, class)
			}

		}
	}
}

// /event handler
func eventCommand(c apic.ApicInterface, m Message, wm WebexMessage) string {
	res := ""
	indMap := map[string]string{"creation": "❇️", "modification": "🔄", "deletion": "🗑"}
	events := splitFaultsAndEnvents(m.cmd)

	var err error
	var info []apic.ApicMoAttributes

	if user, ok := events["user"]; ok {
		info, err = c.GetLatestEvents(events["count"], user)
	} else {
		info, err = c.GetLatestEvents(events["count"])
	}

	if err != nil {
		log.Printf("Error while connecting to the Apic. Err: %s", err)
		return fmt.Sprintf("Hi %s 🤖 !. I could not reach the APIC... Are there any issues?", wm.sender)
	}

	if len(info) == 0 {
		return fmt.Sprintf("Hi %s 🤖 !. There are no events for username <code>%s</code>", wm.sender, events["user"])
	}

	res += fmt.Sprintf("\nThese are the latest %s events in the the Fabric : \n\n", events["count"])

	res += "<ul>"
	for _, f := range info {
		res += fmt.Sprintf("<li><strong>%s</strong> - <em>%s</em>", f["code"], f["affected"])
		res += "<ul>"
		res += fmt.Sprintf("<li>%s</li>", f["descr"])
		res += fmt.Sprintf("<li><strong>User</strong>: %s</li>", f["user"])
		res += fmt.Sprintf("<li><strong>Type</strong>: %s %s</li>", f["ind"], indMap[f["ind"]])
		res += fmt.Sprintf("<li><strong>Created</strong>: %s</li>", f["created"])
		res += "</ul>"
	}
	res += "</ul>"
	return fmt.Sprintf("Hi %s 🤖 !\n\n%s", wm.sender, res)
}

// /fault handler
func faultCommand(c apic.ApicInterface, m Message, wm WebexMessage) string {
	res := ""
	sevMap := map[string]string{"critical": "📛", "major": "☢️", "minor": "⚠️", "warning": "🌀", "cleared": "❎"}
	lcMap := map[string]string{"soaking": "♻️", "retaining": "✅", "raised": "❌", "soaking-clearing": "♻️", "raised-clearing": "♻️"}
	faults := splitFaultsAndEnvents(m.cmd)

	info, err := c.GetLatestFaults(faults["count"])

	if err != nil {
		log.Printf("Error while connecting to the Apic. Err: %s", err)
		return fmt.Sprintf("Hi %s 🤖 !. I could not reach the APIC... Are there any issues?", wm.sender)
	}

	res += fmt.Sprintf("\nThese are the latest %s faults in the the Fabric : \n\n", faults)

	res += "<ul>"
	for _, f := range info {
		res += fmt.Sprintf("<li><strong>%s</strong> - <em>%s</em>", f["code"], f["dn"])
		res += "<ul>"
		res += fmt.Sprintf("<li>%s</li>", f["descr"])
		res += fmt.Sprintf("<li><strong>Severity</strong>: %s %s</li>", f["severity"], sevMap[f["severity"]])
		res += fmt.Sprintf("<li><strong>Current Lyfecycle</strong>: %s %s</li>", f["lc"], lcMap[f["lc"]])
		res += fmt.Sprintf("<li><strong>Type</strong>: %s</li>", f["type"])
		res += fmt.Sprintf("<li><strong>Created</strong>: %s</li>", f["created"])
		res += "</ul>"
	}
	res += "</ul>"
	return fmt.Sprintf("Hi %s 🤖 !\n\n%s", wm.sender, res)
}

// /neigh handler
func neighCommand(c apic.ApicInterface, m Message, wm WebexMessage) string {
	res := ""
	neighId := splitNeighCommand(m.cmd)
	info, err := c.GetFabricNeighbors(neighId["neigh"])

	// Sort by Neigh Name
	keys := make([]string, 0, len(info))
	for k := range info {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	if err != nil {
		log.Printf("Error while connecting to the Apic. Err: %s", err)
		return fmt.Sprintf("Hi %s 🤖 !. I could not reach the APIC... Are there any issues?", wm.sender)
	}

	if len(info) == 0 && neighId["neigh"] != "all" {
		return fmt.Sprintf("Hi %s 🤖 !\n It seems there are no Neighbors for Node <code>%s</code>", wm.sender, neighId["neigh"])
	} else if len(info) == 0 && neighId["neigh"] == "all" {
		return fmt.Sprintf("Hi %s 🤖 !\n Sorry.. I could not discover the Topology of the Fabric", wm.sender)
	}

	if neighId["neigh"] == "all" {
		res += "\nThis is the Topology information of the Fabric : \n\n"
	} else {
		res += fmt.Sprintf("\nThese are the Neighbors of the Node <code>%s</code>: \n\n", neighId["neigh"])
	}
	res += "<ul>"

	for _, k := range keys {
		res += fmt.Sprintf("<li><strong>%s</strong>:\t", k)
		for _, n := range info[k] {
			res += fmt.Sprintf("%s   ", n)
		}
		res += "</li>"
	}
	res += "</ul>"
	return fmt.Sprintf("Hi %s 🤖 !\n\n%s", wm.sender, res)
}

// /ep <ep_mac> handler
func endpointCommand(c apic.ApicInterface, m Message, wm WebexMessage) string {

	res := ""
	info, err := c.GetEndpointInformation(splitEpCommand(m.cmd)["mac"])
	if err != nil {
		log.Printf("Error while connecting to the Apic. Err: %s", err)
		return fmt.Sprintf("Hi %s 🤖 !. I could not reach the APIC... Are there any issues?", wm.sender)

	}
	res = res + fmt.Sprintf("\nThis is the information for the Endpoint <code>%s</code>", splitEpCommand(m.cmd)["mac"])
	res = res + "<ul>"
	for _, item := range info {
		res = res + fmt.Sprintf("<li><strong>Tenant</strong>: %s</li>", item.Tenant)
		res = res + fmt.Sprintf("<li><strong>Application Profile</strong>: %s</li>", item.App)
		res = res + fmt.Sprintf("<li><strong>EPG</strong>: %s</li>", item.Epg)
		for idx, path := range item.Location {
			res = res + fmt.Sprintf("<li><strong>Location %d</strong>: </li>", idx+1)
			res = res + "<ul>"
			res = res + fmt.Sprintf("<li><strong>Pod</strong>: %s", path["pod"])
			res = res + fmt.Sprintf("  <strong>Node</strong>: %s", path["nodes"])
			res = res + fmt.Sprintf("  <strong>Type</strong>: %s", path["type"])
			res = res + fmt.Sprintf("  <strong>Port</strong>: %s</li>", path["port"])
			res = res + "</ul>"
		}
		if len(item.Ips) > 0 {
			res = res + "<li><strong>IPs</strong>: </li>"
			res = res + "<ul>"
			for _, path := range item.Ips {
				res = res + fmt.Sprintf("<li><strong>IP</strong>: %s</li>", path)
			}
			res = res + "</ul>"
		}
	}
	res = res + "</ul>"
	return fmt.Sprintf("Hi %s 🤖 !\n\n%s", wm.sender, res)
}

// /info handler
func infoCommand(c apic.ApicInterface, m Message, wm WebexMessage) string {
	res := ""
	info, err := c.GetFabricInformation()

	if err != nil {
		log.Printf("Error while connecting to the Apic. Err: %s", err)
		return fmt.Sprintf("Hi %s 🤖 !. I could not reach the APIC... Are there any issues?", wm.sender)

	}
	res = res + fmt.Sprintf("\nThis is the general information of the Fabric <code>%s</code> (%s): \n\n", info.Name, info.Url)
	res = res + fmt.Sprintf("<ul><li>Current Health Score: <strong>%s</strong></li>", info.Health)
	res = res + "<li><strong>APIC Controllers</strong><ul>"
	for _, item := range info.Apics {
		res = res + "<li>" + item["name"] + " (<strong>" + item["version"] + "</strong>)</li>"
	}
	res = res + "</ul><li><strong>Pods</strong><ul>"
	for _, item := range info.Pods {
		res = res + "<li>Pod" + item["id"] + " <em>" + item["type"] + "</em></li>"
	}
	res = res + "</ul><li></strong>Switches</strong><ul>"
	res = res + fmt.Sprintf("<li># of Spines : <strong>%d</strong></li>", len(info.Spines))
	res = res + fmt.Sprintf("<li># of Leafs : <strong>%d</strong></li>", len(info.Leafs))
	res = res + "</ul></ul></li>"
	return fmt.Sprintf("Hi %s 🤖 !\n\n%s", wm.sender, res)
}

// /cpu handler
func cpuCommand(c apic.ApicInterface, m Message, wm WebexMessage) string {
	res := ""
	cpu, err := c.GetProcEntity()

	if err != nil {
		log.Printf("Error while connecting to the Apic. Err: %s", err)
		return fmt.Sprintf("Hi %s 🤖 !. I could not reach the APIC... Are there any issues?", wm.sender)
	}
	res = res + "\nThis is the CPU information of the controllers: \n\n"
	res = res + "<ul>"

	for _, item := range cpu {
		memFree, _ := strconv.ParseFloat(item["memFree"], 32)
		memMax, _ := strconv.ParseFloat(item["maxMemAlloc"], 32)
		res = res + fmt.Sprintf("<li><code>APIC %s</code> -> \t💻 <strong>CPU: </strong>%s\t💾 <strong>Memory %%: </strong> %f</li>", apic.GetRn(item["dn"], "node"), item["cpuPct"], 100.0*memFree/memMax)
	}
	res = res + "</ul>"
	return fmt.Sprintf("Hi %s 🤖 !\n\n	%s", wm.sender, res)
}

// /help handler
func helpCommand(cmd map[string]Command) Callback {
	return func(a apic.ApicInterface, m Message, wm WebexMessage) string {

		keys := make([]string, 0, len(cmd))
		for k := range cmd {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		res := fmt.Sprintf("Hello %s, How can I help you?\n\n", wm.sender)
		res = res + "<ul>"
		for _, k := range keys {
			res = res + fmt.Sprintf("<li><code>%s</code>\t->\t%s</li>", k, cmd[k].help)
		}
		res = res + "<ul>"
		return res
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
		message, err := wbx.GetMessageById(wh.Data.Id)
		if err != nil {
			log.Printf("failed trying to retrieve the last message. Error %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		// Is the message send from someone who is not the bot
		if message.PersonId != b.Id {
			// Get sender personal information
			sender, _ := wbx.GetPersonInformation(message.PersonId)
			found := false
			// Check which command was sent in the webex room
			messageText := cleanCommand(b.DisplayName, message.Text)
			for cli, element := range cmd {
				if MatchCommand(messageText, element.regex) {
					// Send message back the text is returned from the commandHandler
					wbx.SendMessageToRoom(element.callback(ap, Message{cmd: messageText}, WebexMessage{sender: sender.NickName, roomId: message.RoomId}), wh.Data.RoomId)
					found = true
					w.WriteHeader(http.StatusOK)
					return
				}
				// Matches the first word but the arguments does not fit. Send back the usage
				if MatchCommand(messageText, element.suffix) {
					wbx.SendMessageToRoom(fmt.Sprintf("Hi %s 🤖 \n I could not fully understand the input\n Please check the usage of the <code>%s</code> command:\n <ul><li>%s</ul></li>\n", sender.NickName, cli, element.help), wh.Data.RoomId)
					found = true
					w.WriteHeader(http.StatusOK)
					return
				}
			}
			// If command sent does not match anything, send back the help menu
			if !found {
				wbx.SendMessageToRoom(cmd["/help"].callback(ap, Message{cmd: messageText}, WebexMessage{sender: sender.NickName, roomId: message.RoomId}), wh.Data.RoomId)
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
	// b.router.HandleFunc("/webscket", websocketHandler(b.wbx, b.apic, b.info))
}
func (b *Bot) addCommand(cmd string, help string, suf string, re string, call Callback) {
	// add item to the dispatch table
	b.commands[cmd] = Command{
		help:     help,
		callback: call,
		suffix:   suf,
		regex:    suf + re,
	}
}
func (b *Bot) setupWebhook() error {
	// TODO: Delete exsiting webhooks with the same name

	whs, _ := b.wbx.GetWebHooks()
	for _, wh := range whs {
		if wh.Name == b.info.DisplayName {
			log.Printf("Bot already has a Webhook with name %s\n", b.info.DisplayName)
			err := b.wbx.DeleteWebhook(wh.Id)
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

func readWebsocket(b *Bot) {
	statusMap := map[string]string{"deleted": "❌", "created": "✅", "modified": "✏️"}
	for {

		subId, events := b.wsck.ReadSocketEvent()
		className := b.wsSubs.getClassNamebySubId(subId)

		msg := "<ul>"
		for _, event := range events {
			msg += fmt.Sprintf("<li>The object <code>%s</code> has been <strong>%s</strong></li> %s", event["dn"], event["status"], statusMap[event["status"].(string)])
		}
		msg += "</ul>"

		for _, room := range b.wsSubs.getRoomsIdbyClass(className) {
			roomInfo, _ := b.wbx.GetRoomById(room)
			res := fmt.Sprintf("Hi %s 🤖\n Notification from the APIC %s \n %s", roomInfo.Title, b.apic.GetIp(), msg)
			b.wbx.SendMessageToRoom(res, room)
		}
	}
}

func refreshApicClient(b *Bot, t int) {
	tickerToken := time.NewTicker(time.Duration(t) * time.Second)
	tickerWs := time.NewTicker(60 * time.Second)
	defer tickerToken.Stop()
	defer tickerWs.Stop()
	for {

		select {
		case <-tickerToken.C:
			log.Printf("Refreshing REST APIC Token\n")
			b.apic.Login()
			log.Printf("Refreshing Websocket APIC Token")
			b.wsck.NewDial(b.apic.GetToken())

		case <-tickerWs.C:
			for class, subId := range b.wsSubs.getActiveSubscriptions() {
				log.Printf("Refreshing subscription %s - %s", class, subId)
				b.apic.RefreshSubscriptionWebSocket(subId)
			}
		}
	}
}

func (b *Bot) SetupWebSocket() error {

	b.wsck, _ = apic.NewApicWebSClient(b.apic.GetIp(), b.apic.GetToken())
	go refreshApicClient(b, 180)
	go readWebsocket(b)
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
