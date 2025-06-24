package sci

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
)

var (
	httpMethods = []string{
		"GET",
		"POST",
		"PATCH",
		"DELETE",
		"PUT",
		"HEAD",
		"OPTIONS",
		"CONNECT",
		"TRACE",
	}
	maskHeader = strings.ToLower("X-Auth-Token:")
)

type logger struct {
	svc string
}

func (l logger) Printf(format string, args ...any) {
	if len(format) == 0 || format[len(format)-1] != '\n' {
		format += "\n"
	}
	fmt.Fprintf(os.Stderr, format, args...)
}

func (l logger) Debugf(format string, args ...any) {
	for _, arg := range args {
		v, ok := arg.(string)
		if !ok {
			continue
		}
		str := deleteEmpty(strings.Split(v, "\n"))
		cycle := "Response"
		if len(str) > 0 {
			for _, method := range httpMethods {
				if strings.HasPrefix(str[0], method) {
					cycle = "Request"
					break
				}
			}
		}
		printed := false

		for i, s := range str {
			if i == 0 && cycle == "Request" {
				v := strings.SplitN(s, " ", 3)
				if len(v) > 1 {
					log.Printf("[DEBUG] %s %s URL: %s %s", l.svc, cycle, v[0], v[1])
				}
			} else if i == 0 && cycle == "Response" {
				v := strings.SplitN(s, " ", 2)
				if len(v) > 1 {
					log.Printf("[DEBUG] %s %s Code: %s", l.svc, cycle, v[1])
				}
			} else if i == len(str)-1 {
				debugInfo, err := formatJSON([]byte(s))
				if err != nil {
					printHeaders(l.svc, cycle, &printed)
					log.Print(s)
				} else {
					log.Printf("[DEBUG] %s %s Body: %s", l.svc, cycle, debugInfo)
				}
			} else if strings.HasPrefix(strings.ToLower(s), maskHeader) {
				printHeaders(l.svc, cycle, &printed)
				v := strings.SplitN(s, ":", 2)
				log.Printf("%s: ***", v[0])
			} else {
				printHeaders(l.svc, cycle, &printed)
				log.Print(s)
			}
		}
	}
}

func deleteEmpty(s []string) []string {
	var r []string
	for _, str := range s {
		if strings.TrimSpace(str) != "" {
			r = append(r, str)
		}
	}
	return r
}

func printHeaders(name, cycle string, printed *bool) {
	if !*printed {
		log.Printf("[DEBUG] %s %s Headers:", name, cycle)
		*printed = true
	}
}

// formatJSON is a function to pretty-format a JSON body.
// It will also mask known fields which contain sensitive information.
func formatJSON(raw []byte) (string, error) {
	var rawData any

	err := json.Unmarshal(raw, &rawData)
	if err != nil {
		return string(raw), fmt.Errorf("unable to parse OpenStack JSON: %s", err)
	}

	data, ok := rawData.(map[string]any)
	if !ok {
		pretty, err := json.MarshalIndent(rawData, "", "  ")
		if err != nil {
			return string(raw), fmt.Errorf("unable to re-marshal OpenStack JSON: %s", err)
		}

		return string(pretty), nil
	}

	// Strip kubeconfig
	if _, ok := data["kubeconfig"].(string); ok {
		data["kubeconfig"] = "***"
	}

	pretty, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return string(raw), fmt.Errorf("unable to re-marshal OpenStack JSON: %s", err)
	}

	return string(pretty), nil
}
