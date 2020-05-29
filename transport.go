package main

import (
	"net/http"
	"net/url"

	log "github.com/sirupsen/logrus"
)

type Transport struct {
	Close func()
}

func (t *Transport) Default() *http.Transport {
	return &http.Transport{}
}

func (t *Transport) Proxy(addr string) *http.Transport {
	url, err := url.Parse(addr)

	if err != nil {
		log.Panic(err)
	}

	return &http.Transport{Proxy: http.ProxyURL(url)}
}
