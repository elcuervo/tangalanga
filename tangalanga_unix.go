// +build linux

package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/cretz/bine/tor"
	pb "github.com/elcuervo/tangalanga/proto"
	"github.com/golang/protobuf/proto"
	"github.com/ipsn/go-libtor"
	log "github.com/sirupsen/logrus"
)

type Tangalanga struct {
	client       *http.Client
	tor          *tor.Tor
	ErrorCounter int
}

func (t *Tangalanga) torDialer() (*tor.Dialer, error) {
	tor, err := tor.Start(nil, &tor.StartConf{ProcessCreator: libtor.Creator})

	if err != nil {
		fmt.Errorf("Unable to start Tor: %v", err)
	}

	t.tor = tor

	dialCtx, _ := context.WithTimeout(context.Background(), 3*time.Minute)

	return tor.Dialer(dialCtx, nil)
}

func (t *Tangalanga) Close() {
	defer t.tor.Close()
}

func (t *Tangalanga) NewHTTPClient() {
	t.client = &http.Client{}
}

func (t *Tangalanga) NewHTTPTORClient() {
	fmt.Printf("connecting to the TOR network... %s\n", color.Yellow("please wait"))
	dialer, err := t.torDialer()

	if err != nil {
		fmt.Errorf("Unable to start Tor: %v", err)
	}

	fmt.Printf("connection via TOR %s\n", color.Green("successful!"))
	t.client = &http.Client{Transport: &http.Transport{DialContext: dialer.DialContext}}
}

func NewTangalanga() *Tangalanga {
	t := new(Tangalanga)
	t.ErrorCounter = 0

	if *useTor {
		t.NewHTTPTORClient()
	} else {
		t.NewHTTPClient()
	}

	return t
}

func (t *Tangalanga) FindMeeting(id int) (*pb.Meeting, error) {
	meetId := strconv.Itoa(id)
	p := url.Values{"cv": {"5.0.25694.0524"}, "mn": {meetId}, "uname": {"tangalanga"}}

	req, _ := http.NewRequest("POST", zoomUrl, strings.NewReader(p.Encode()))
	cookie := fmt.Sprintf("zpk=%s", *token)

	req.Header.Add("Cookie", cookie)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, _ := t.client.Do(req)
	body, _ := ioutil.ReadAll(resp.Body)

	m := &pb.Meeting{}
	err := proto.Unmarshal(body, m)
	if err != nil {
		log.Panic("err: ", err)
	}

	missing := m.GetError() != 0

	if missing {
		info := m.GetInformation()

		if m.GetError() == 124 {
			fmt.Println(color.Red("Token Expired"))
			os.Exit(1)
		}

		// Not found
		if info == "Meeting not existed." {
			t.ErrorCounter++
		} else {
			t.ErrorCounter = 0
		}

		return nil, fmt.Errorf("%s", color.Red(info))
	}

	return m, nil
}
