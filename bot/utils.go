package bot

import (
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
