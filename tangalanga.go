package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	pb "github.com/elcuervo/tangalanga/proto"
	"github.com/golang/protobuf/proto"
)

type Option func(*Tangalanga)

func WithTransport(transport *http.Transport) Option {
	return func(t *Tangalanga) {
		t.client = &http.Client{
			Transport: transport,
			Timeout:   5 * time.Second,
		}
	}
}

type Tangalanga struct {
	Found      int
	Missing    int
	Suspicious int

	client *http.Client
}

func (t *Tangalanga) Close() {
}

func NewTangalanga(opts ...Option) (*Tangalanga, error) {
	c := &Tangalanga{Found: 0}

	for _, opt := range opts {
		opt(c)
	}

	return c, nil
}

func (t *Tangalanga) FindMeeting(id int) (*pb.Meeting, error) {
	meetId := strconv.Itoa(id)
	p := url.Values{"cv": {"5.0.25694.0524"}, "mn": {meetId}, "uname": {"tangalanga"}}

	req, err := http.NewRequest("POST", zoomUrl, strings.NewReader(p.Encode()))

	if err != nil {
		if *debugFlag {
			fmt.Printf("%s\nerror: %s\n", color.Red("bad request"), err.Error())
		}
		return nil, err
	}

	cookie := fmt.Sprintf("zpk=%s", *token)

	req.Header.Add("Cookie", cookie)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := t.client.Do(req)

	if err != nil {
		if *debugFlag {
			fmt.Printf("%s\nerror: %s\n", color.Red("can't connect to Zoom!!"), err.Error())
		}
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		if *debugFlag {
			fmt.Printf("%s\nerror: %s\n", color.Red("bad body"), err.Error())
		}
		return nil, err
	}

	m := &pb.Meeting{}
	err = proto.Unmarshal(body, m)

	if err != nil {
		if *debugFlag {
			fmt.Printf("%s\nerror: %s\n", color.Red("can't unpack protobuf"), err.Error())
		}
		return nil, err
	}

	missing := m.GetError() != 0

	if missing {
		t.Missing++

		info := m.GetInformation()

		if m.GetError() == 124 {
			fmt.Println(color.Red("token expired"))
		}

		// suspicious not found when there are too many
		if info == "Meeting not existed." {
			t.Suspicious++
		} else {
			t.Suspicious = 0
		}

		return nil, fmt.Errorf("%s: %s", color.Blue("zoom"), color.Red(info))
	} else {
		// Reset suspicious counter when found
		t.Suspicious = 0
		t.Found++
	}

	return m, nil
}
