package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	pb "github.com/elcuervo/tangalanga/proto"
	"github.com/golang/protobuf/proto"

	"github.com/briandowns/spinner"
	"github.com/cretz/bine/tor"
	"github.com/ipsn/go-libtor"
)

const logo = `
▄▄▄▄▄ ▄▄▄·  ▐ ▄  ▄▄ •  ▄▄▄· ▄▄▌   ▄▄▄·  ▐ ▄  ▄▄ •  ▄▄▄·
•██  ▐█ ▀█ •█▌▐█▐█ ▀ ▪▐█ ▀█ ██•  ▐█ ▀█ •█▌▐█▐█ ▀ ▪▐█ ▀█
 ▐█.▪▄█▀▀█ ▐█▐▐▌▄█ ▀█▄▄█▀▀█ ██▪  ▄█▀▀█ ▐█▐▐▌▄█ ▀█▄▄█▀▀█
 ▐█▌·▐█ ▪▐▌██▐█▌▐█▄▪▐█▐█ ▪▐▌▐█▌▐▌▐█ ▪▐▌██▐█▌▐█▄▪▐█▐█ ▪▐▌
 ▀▀▀  ▀  ▀ ▀▀ █▪·▀▀▀▀  ▀  ▀ .▀▀▀  ▀  ▀ ▀▀ █▪·▀▀▀▀  ▀  ▀
`

const zoomUrl = "https://www3.zoom.us/conf/j"
const token = "zpk=lLAcbIV3Irl4YDEFWUtWejg3pFcIuRjWSQjITMXFKIk%3D.BwcAAAFyVi2fUAAAqMAkQTU5NDQxMzMtQTNDNi00RjEwLTk5NUUtOEE1QkYyMzgyMzE3AAAIZWxjdWVydm9mAAAAAAD%2FAAAA;"

func init() {
	rand.Seed(time.Now().UnixNano())
	log.Println(logo)
}

func debugReq(req *http.Request) {
	requestDump, err := httputil.DumpRequest(req, true)
	if err != nil {
		fmt.Println(err)
	}

	log.Println(string(requestDump))
}

func randomMeetingId() int {
	//     88392789130
	min := 80000000000
	max := 99999999999

	return rand.Intn(max-min+1) + min
}

type Tangalanga struct {
	client *http.Client
}

func (t *Tangalanga) FindMeeting(id int) (*pb.Meeting, error) {
	p := url.Values{"cv": {"5.0.25694.0524"}, "mn": {strconv.Itoa(id)}, "uname": {"tangalanga"}}

	req, _ := http.NewRequest("POST", zoomUrl, strings.NewReader(p.Encode()))

	req.Header.Add("Cookie", token)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, _ := t.client.Do(req)
	body, _ := ioutil.ReadAll(resp.Body)

	m := &pb.Meeting{}
	err := proto.Unmarshal(body, m)
	if err != nil {
		log.Panic("err: ", err)
	}

	missing := m.GetMissing()

	if missing {
		return nil, fmt.Errorf(m.GetInformation())
	}

	return m, nil
}

func main() {
	client := &http.Client{}

	tangalanga := &Tangalanga{
		client: client,
	}

	for i := 0; i < 100; i++ {
		id := randomMeetingId()

		m, err := tangalanga.FindMeeting(id)

		if err != nil {
			log.Println(err)
		} else {
			room := m.GetRoom()
			log.Printf("Found ID: %s Hello: %s", room.GetRoomId(), room.GetRoomName())
		}
	}
}

func main2() {
	s := spinner.New(spinner.CharSets[4], 100*time.Millisecond)
	c := make(chan os.Signal, 2)

	s.Suffix = " Connecting to the TOR network."
	s.Start()

	go func() {
		<-c
		os.Exit(0)
	}()

	t, err := tor.Start(nil, &tor.StartConf{ProcessCreator: libtor.Creator})

	if err != nil {
		fmt.Errorf("Unable to start Tor: %v", err)
	}

	defer t.Close()

	dialCtx, dialCancel := context.WithTimeout(context.Background(), 3*time.Minute)

	dialer, err := t.Dialer(dialCtx, nil)

	httpClient := &http.Client{Transport: &http.Transport{DialContext: dialer.DialContext}}

	tangalanga := &Tangalanga{
		client: httpClient,
	}

	tangalanga.FindMeeting(1)

	defer dialCancel()

	s.Stop()
}
