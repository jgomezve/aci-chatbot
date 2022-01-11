package bot

import (
	"aci-chatbot/webex"
	"encoding/json"
	"io/ioutil"
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
	if err != nil {
		return err
	}
	if err = json.Unmarshal(body, &wh); err != nil {
		return err
	}
	return nil
}
