
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"log/syslog"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
)

var memegenUrl = "https://memegen.link/"

type memeTemplate struct {
	Name string
	Desc string
}
var memeTemplates = []memeTemplate{}
var authToken string
var imgWidth *string

func help(w http.ResponseWriter) {
	h := "\\nMattermost Meme Bot\\n" +
             "**> Commands:**\\n" +
             "* `/meme memename;top_row;bottom_row` generate a meme image\\n" +
             "    (NOTE: memename can also be a URL to an image)\\n" +
             "* `/meme list` List templates\\n" +
             "* `/meme help` Shows this menu\\n"
	fmt.Fprint(w, `{ "response_type": "ephemeral", "text": "` + h + `"}`)
}

func getTemplates(w http.ResponseWriter) bool {
	// Get JSON
	api := memegenUrl + "/api/templates/"
	var body []byte
	resp, err := http.Get(api)
	if err == nil {
		body, err = ioutil.ReadAll(resp.Body)
		resp.Body.Close()
	}
	if err != nil {
		fmt.Fprint(w, `{
			"response_type": "ephemeral",
			"text": "` + api + ": " + err.Error() + `"
		}`)
		return false
	}
	// Decode JSON into map
	var kv map[string]string
	if err := json.Unmarshal(body, &kv); err != nil {
		fmt.Fprint(w, `{
			"response_type": "ephemeral",
			"text": "` + api + ": " + err.Error() + `"
		}`)
		return false
	}
	// reverse map
	kv2 := make(map[string]string)
	keys := []string{}
	for k, v := range kv {
		i := strings.LastIndex(v, "/")
		if i < 0 {
			continue
		}
		kv2[v[i+1:]] = k
		keys = append(keys, v[i+1:])
	}
	sort.Strings(keys)
	// put into array
	r := []memeTemplate{}
	for _, v := range keys {
		r = append(r, memeTemplate{v, kv2[v]})
	}
	memeTemplates = r
	return true
}

func listTemplates(w http.ResponseWriter) {
	if len(memeTemplates) == 0 {
		if !getTemplates(w) {
			return
		}
	}
	r := []string{}
	for _, m := range memeTemplates {
		r = append(r, fmt.Sprintf("%-20s %s", m.Name, m.Desc))
	}
	fmt.Fprint(w, `{
		"response_type": "ephemeral",
		"text": "` + "```\\n" + strings.Join(r, "\\n") + "```" + `"
	}`)
}

func memeHandler(w http.ResponseWriter, r *http.Request) {
	text := r.FormValue("text")
	s := strings.Split(text, ";")
	for i := range s {
		s[i] = strings.Trim(s[i], " \t")
	}

	if len(s) == 1 && s[0] == "help" {
		help(w)
		return
	}
	if len(s) == 1 && (s[0] == "templates" || s[0] == "list") {
		listTemplates(w)
		return
	}
	if len(s) < 1 || s[0] == "" {
		fmt.Fprint(w, `{
			"response_type": "ephemeral",
			"text": "try: /meme help"
		}`)
		return
	}

	if len(memeTemplates) == 0 {
		if !getTemplates(w) {
			return
		}
	}

	found := false
	query := ""
	if len(s[0]) > 7 &&
	   (s[0][:7] == "http://" || s[0][:8] == "https://") {
		found = true
		query = "alt=" + s[0]
		s[0] = "custom"
	}
	for _, m := range memeTemplates {
		if m.Name == s[0] {
			found = true
			break
		}
	}
	if !found {
		fmt.Fprint(w, `{
			"response_type": "ephemeral",
			"text": "meme not found\ntry: /meme list"
		}`)
		return
	}

	for len(s) < 3 {
		s = append(s, "")
	}
	if s[1] == "" {
		s[1] = "_"
	}
	if s[2] == "" {
		s[2] = "_"
	}
	var url *url.URL
	url, _ = url.Parse(memegenUrl)
	url.Path = s[0] + "/" + s[1] + "/" + s[2] + ".jpg"
	url.RawQuery = query
	sz := ""
	if imgWidth != nil && *imgWidth != "" {
		sz = " =" + *imgWidth + "x"
	}
	fmt.Fprint(w, `{
		"response_type": "in_channel",
		"text": "![image](` + url.String() + sz + `)"
	}`)
}

func setLog(logfile string) {
	switch logfile {
	case "syslog":
		logw, err := syslog.New(syslog.LOG_NOTICE, "memebot")
		if err != nil {
			log.Fatalf("error opening syslog: %v", err)
		}
		log.SetOutput(logw)
	case "none":
		log.SetOutput(ioutil.Discard)
	case "stdout":
	default:
		f, err := os.OpenFile(logfile,
			os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Fatalf("error opening file: %v", err)
		}
		log.SetOutput(f)
	}
	log.SetFlags(0)
}

func main() {

	env_listen := os.Getenv("MEMEBOT_LISTEN")
	env_logfile := os.Getenv("MEMEBOT_LOG")
	env_imgwidth := os.Getenv("MEMEBOT_IMGWIDTH")
	authToken = os.Getenv("MEMEBOT_TOKEN")
	if env_listen == "" {
		env_listen = ":5020"
	}
	if env_logfile == "" {
		env_logfile = "stdout"
	}
	if env_imgwidth == "" {
		env_imgwidth = "250"
	}

	logfile := flag.String("logfile", env_logfile,
		"Path of logfile. Use 'syslog' for syslog, 'stdout' " +
		"for standard output, or 'none' to disable logging.")
	imgWidth = flag.String("imgwidth", env_imgwidth,
		"Width of image in pixels")
	listen := flag.String("listen", env_listen, "Server listen address")
	flag.Parse()
	if logfile != nil {
		setLog(*logfile)
	}

	http.HandleFunc("/", memeHandler)
	log.Fatal(http.ListenAndServe(*listen, nil))
}

