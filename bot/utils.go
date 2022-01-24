package bot

import (
	"aci-chatbot/webex"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strconv"
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

func splitNeighCommand(s string) map[string]string {
	w := strings.Split(s, " ")
	if len(w) == 1 {
		return map[string]string{"neigh": "all"}
	} else {
		return map[string]string{"neigh": w[1]}
	}
}

func splitFaultsAndEnvents(s string) map[string]string {
	w := strings.Split(s, " ")
	switch len(w) {
	case 3:
		return map[string]string{"user": w[1], "count": w[2]}
	case 2:
		if c, _ := strconv.Atoi(w[1]); c != 0 {
			return map[string]string{"count": w[1]}
		} else {
			return map[string]string{"user": w[1], "count": "10"}
		}
	case 1:
		return map[string]string{"count": "10"}
	}
	return map[string]string{}
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
