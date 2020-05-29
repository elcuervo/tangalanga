// +build linux

package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/cretz/bine/tor"
	"github.com/ipsn/go-libtor"
)

func (t *Transport) torDialer() (*tor.Dialer, error) {
	tor, err := tor.Start(nil, &tor.StartConf{ProcessCreator: libtor.Creator})

	if err != nil {
		fmt.Errorf("Unable to start Tor: %v", err)
	}

	dialCtx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)

	t.Close = func() {
		cancel()
		tor.Close()
	}

	return tor.Dialer(dialCtx, nil)
}

func (t *Transport) InteralTOR() *http.Transport {
	d, _ := t.torDialer()
	return &http.Transport{DialContext: d.DialContext}
}
