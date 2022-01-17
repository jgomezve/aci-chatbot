package apic

import (
	"fmt"
	"strings"
)

func getApicManagedObjects(p map[string]interface{}, c string) []ApicMoAttributes {
	mos := []ApicMoAttributes{}

	for _, item := range p["imdata"].([]interface{}) {
		att := make(map[string]string)
		for k, v := range item.(map[string]interface{})[c].(map[string]interface{})["attributes"].(map[string]interface{}) {
			att[k] = v.(string)
		}
		mos = append(mos, att)
	}
	return mos
}

func getApicManagedObjectsChildren(p map[string]interface{}, parent string, children string) []ApicMoAttributes {
	mos := []ApicMoAttributes{}

	for _, p := range p["imdata"].([]interface{}) {
		ch := p.(map[string]interface{})[parent].(map[string]interface{})["children"]
		if ch != nil {
			for _, c := range ch.([]interface{}) {
				att := make(map[string]string)
				for k, v := range c.(map[string]interface{})[children].(map[string]interface{})["attributes"].(map[string]interface{}) {
					att[k] = v.(string)
				}
				mos = append(mos, att)
			}
		}
	}
	return mos
}

func GetRn(dn string, rnId string) string {
	fSplit := strings.Split(dn, "/")
	var sSplit []string
	start := 0
	joining := false
	for idx, item := range fSplit {
		if strings.Contains(item, "[") && strings.Contains(item, "]") {
			sSplit = append(sSplit, item)
		}
		if strings.Contains(item, "[") {
			start = idx
			joining = true
		} else if strings.Contains(item, "]") {
			sSplit = append(sSplit, strings.Join(fSplit[start:idx+1], "/"))
			joining = false
		} else if !joining {
			sSplit = append(sSplit, item)
		}

	}
	for _, rn := range sSplit {
		id := strings.Split(rn, "-")[0]
		if id == rnId {
			return strings.Join(strings.Split(rn, "-")[1:], "-")
		}
	}
	return ""
}

func getPath(tdn string) map[string]string {
	path := make(map[string]string)
	fmt.Println(tdn)
	if strings.Contains(tdn, "protpaths") && !strings.Contains(tdn, "tunnel") {
		path["pod"] = GetRn(tdn, "pod")
		path["type"] = "vPC"
		path["nodes"] = GetRn(tdn, "protpaths")
		path["port"] = GetRn(tdn, "pathep")
		return path
	} else if strings.Contains(tdn, "pathep") && !strings.Contains(tdn, "tunnel") {
		pathEp := GetRn(tdn, "pathep")
		fmt.Println(pathEp)
		if !strings.Contains(pathEp, "tunnel") && strings.Contains(pathEp, "/") {
			path["pod"] = GetRn(tdn, "pod")
			path["type"] = "Access"
			path["nodes"] = GetRn(tdn, "paths")
			path["port"] = GetRn(tdn, "pathep")
		} else if !strings.Contains(pathEp, "tunnel") {
			path["pod"] = GetRn(tdn, "pod")
			path["type"] = "PC"
			path["nodes"] = GetRn(tdn, "paths")
			path["port"] = GetRn(tdn, "pathep")
		}
		return path
	}
	return nil
}
