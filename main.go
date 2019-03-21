package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// Link is a HATEOAS links
type Link struct {
	Rel  string `json:"rel"`
	Href string `json:"href"`
	Type string `json:"type"`
}

// HandlerEntry ist the spec to a http handler
type HandlerEntry struct {
	Link    Link
	Handler func(w http.ResponseWriter, r *http.Request)
}

// Probe is a testing struct
type Probe struct {
	ID    int64  `json:"id"`
	Text  string `json:"text"`
	Links []Link `json:"links"`
}

var handlerMap = make(map[string]HandlerEntry)

func main() {
	addHandler(Link{Href: "/", Rel: "index", Type: "GET"}, index)
	http.ListenAndServe(":8080", nil)
}

func index(w http.ResponseWriter, r *http.Request) {
	p := Probe{ID: 123, Text: "Text"}
	p.Links = append(p.Links, createLink("self", index, r))
	response, err := json.Marshal(p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(response)
}

func addHandler(link Link,
	handler func(w http.ResponseWriter, r *http.Request)) {
	handlerMap[link.Rel] = HandlerEntry{Link: link, Handler: handler}
	http.HandleFunc(link.Href, handler)
}

func createLink(rel string, handler func(w http.ResponseWriter, r *http.Request), r *http.Request) Link {
	var host string
	var port string
	spl := strings.Split(r.Host, ":")
	if len(spl) == 1 {
		host = r.Host
	} else {
		host = spl[0]
		port = ":" + spl[1]
	}
	proto := "http"
	if r.TLS != nil {
		proto = "https"
	}
	if xhost := r.Header.Get("x-forwarded-host"); xhost != "" {
		host = xhost
	}
	if xport := r.Header.Get("x-forwarded-port"); xport != "" {
		port = ":" + xport
	}
	if xproto := r.Header.Get("x-forwarded-proto"); xproto != "" {
		proto = xproto
	}
	base := proto + "://" + host + port
	h := fmt.Sprintf("%v", handler)
	for _, v := range handlerMap {
		if h == fmt.Sprintf("%v", v.Handler) {
			return Link{Rel: rel, Href: base + v.Link.Href, Type: v.Link.Type}
		}
	}
	return Link{Rel: rel}
}
