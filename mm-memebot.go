
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

type Response struct {
	ResponseType	string	`json:"response_type"`
	Text		string	`json:"text"`
}

func help(w http.ResponseWriter) {
	h := "\n  \nMattermost Meme Bot\n" +
             "**> Commands:**\n" +
             "* `/meme memename top_row;bottom_row` generate a meme image  \n" +
             "    (NOTE: memename can also be a URL to an image)\n" +
             "* `/meme list` List templates\n" +
             "* `/meme help` Shows this menu\n\n"
	responseEphemeral(w, h)
}

func responseText(w http.ResponseWriter, tp string, text string) {
	r := &Response{
		ResponseType:	tp,
		Text:		text,
	}
	b, _ := json.Marshal(r)
	w.Write(b)
}

func responseEphemeral(w http.ResponseWriter, text string) {
	responseText(w, "ephemeral", text)
}

func getTemplates(w http.ResponseWriter) bool {
	// Get JSON
	api := memegenUrl + "api/templates/"
	var body []byte
	resp, err := http.Get(api)
	if err == nil {
		body, err = ioutil.ReadAll(resp.Body)
		resp.Body.Close()
	}
	if err != nil {
		responseEphemeral(w, "error: " + err.Error() + "\n")
		return false
	}
	// Decode JSON into map
	var kv map[string]string
	if err := json.Unmarshal(body, &kv); err != nil {
		responseEphemeral(w, "error: " + api + ": " + err.Error() + "\n")
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
	responseEphemeral(w, "```\n" + strings.Join(r, "\n") + "```\n")
}

func escape(text string) string {
	if text == "" {
		return "_"
	}
	ret := ""
	for _, c := range text {
		switch c {
		case ' ':
			ret += "_"
		case '_':
			ret += "__"
		case '-':
			ret += "--"
		case '?':
			ret += "~q"
		case '%':
			ret += "~p"
		case '#':
			ret += "~h"
		case '/':
			ret += "~s"
		case '"':
			ret += "''"
		default:
			ret += string(c)
		}
	}
	log.Print("escape: " + text + " -> " + ret)
	return ret
}

func memeHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	// try to split into command<SPACE>rest or command<SEMICOLON>rest
	text := strings.Trim(r.FormValue("text"), " \t")
	p := strings.IndexAny(text, " ;")

	if p <= 0 {
		// single command.
		switch(text) {
		case "help":
			help(w)
		case "templates", "list":
			listTemplates(w)
		default:
			responseEphemeral(w, "try: /meme help\n")
		}
		return
	}

	meme := text[:p]
	s := strings.Split(text[p+1:], ";")
	for i := range s {
		s[i] = strings.Trim(s[i], " \t")
	}
	for len(s) < 2 {
		s = append(s, "")
	}
	top := escape(s[0])
	bottom := escape(s[1])

	found := false
	query := ""
	if len(meme) > 7 &&
	   (meme[:7] == "http://" || meme[:8] == "https://") {
		// custom image.
		found = true
		query = "alt=" + meme
		meme = "custom"
	} else {
		// find meme
		if len(memeTemplates) == 0 {
			if !getTemplates(w) {
				return
			}
		}
		for _, m := range memeTemplates {
			if m.Name == meme {
				found = true
				break
			}
		}
	}
	if !found {
		responseEphemeral(w, "meme not found\ntry: /meme list\n")
		return
	}

	var url *url.URL
	url, _ = url.Parse(memegenUrl)
	url.Path = meme + "/" + top + "/" + bottom + ".jpg"
	url.RawQuery = query
	sz := ""
	if imgWidth != nil && *imgWidth != "" {
		sz = " =" + *imgWidth + "x"
	}
	log.Print("url: " + url.String())
	responseText(w, "in_channel", `![image](` + url.String() + sz + `)`)
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

