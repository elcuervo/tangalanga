package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	pb "github.com/elcuervo/tangalanga/proto"
	"github.com/golang/protobuf/proto"
	log "github.com/sirupsen/logrus"
)

type Option func(*Tangalanga)

func WithTransport(transport *http.Transport) Option {
	return func(t *Tangalanga) {
		t.client = &http.Client{Transport: transport}
	}
}

type Tangalanga struct {
	client       *http.Client
	ErrorCounter int
}

func (t *Tangalanga) Close() {
}

func NewTangalanga(opts ...Option) (*Tangalanga, error) {
	c := &Tangalanga{ErrorCounter: 0}

	for _, opt := range opts {
		opt(c)
	}

	return c, nil
}

func (t *Tangalanga) FindMeeting(id int) (*pb.Meeting, error) {
	meetId := strconv.Itoa(id)
	p := url.Values{"cv": {"5.0.25694.0524"}, "mn": {meetId}, "uname": {"tangalanga"}}

	req, _ := http.NewRequest("POST", zoomUrl, strings.NewReader(p.Encode()))
	cookie := fmt.Sprintf("zpk=%s", *token)

	req.Header.Add("Cookie", cookie)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := t.client.Do(req)

	if err != nil {
		fmt.Printf("%s\nerror: %s\n", color.Red("can't connect to Zoom!!"), err.Error())
		os.Exit(1)
	}

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return nil, nil
	}

	m := &pb.Meeting{}
	err = proto.Unmarshal(body, m)
	if err != nil {
		log.Panic("err: ", err)
	}

	missing := m.GetError() != 0

	if missing {
		info := m.GetInformation()

		if m.GetError() == 124 {
			fmt.Println(color.Red("token expired"))
			//			os.Exit(1)
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
