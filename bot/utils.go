package bot

import (
	"aci-chatbot/webex"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strings"
)

func MatchCommand(s string, re string) bool {
	r, _ := regexp.Compile(re)
	return r.MatchString(s)
}

func splitEpCommand(s string) map[string]string {
	w := strings.Split(s, " ")
	cmd := make(map[string]string)
	cmd["mac"] = w[1]
	return cmd
}

func parseWebHook(wh *webex.WebexWebhook, r *http.Request) error {
	body, err := ioutil.ReadAll(r.Body)
	log.Printf("Parsing Webhook Payload\n")
	if err != nil {
		return err
	}
	if err = json.Unmarshal(body, &wh); err != nil {
		return err
	}
	if wh.Name == "" {
		return errors.New("unknown webhok payload")
	}
	return nil
}

func cleanCommand(name string, text string) string {
	var cleaned []string
	for _, w := range strings.Split(strings.TrimSpace(text), " ") {
		if w != name && w != "" {
			cleaned = append(cleaned, w)
		}
	}

	return strings.Join(cleaned, " ")
}
