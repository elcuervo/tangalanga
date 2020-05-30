package main

import (
	"flag"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"strings"
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

// added at compile time
var version string
var baked string

var color aurora.Aurora
var tangalanga *Tangalanga
var wg sync.WaitGroup
var ids chan int
var start time.Time

var token = flag.String("token", availableToken(), "zpk token to use")
var colorFlag = flag.Bool("colors", true, "enable or disable colors")
var censorFlag = flag.Bool("censor", false, "enable or disable stdout censorship")
var outputFile = flag.String("output", "", "output file for successful finds")
var debugFlag = flag.Bool("debug", false, "show error messages")
var torFlag = flag.Bool("tor", false, "connect via tor")
var proxyAddr = flag.String("proxy", "socks5://127.0.0.1:9150", "socks url to use as proxy")
var hiddenFlag = flag.Bool("hidden", false, "connect via embedded tor")
var rateCount = flag.Int("rate", runtime.NumCPU(), "worker count. defaults to CPU count")

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

		fmt.Printf("%s can also be defined in the ENV.\n", color.Red("TOKEN"))
		fmt.Println()

		info := "token can be found by sniffing the traffic trying to join any meeting\n" +
			"currently i can't find how the token is generated but any lives for ~24 hours.\n" +
			"the token can be found as part of the Cookie header.\n" +
			"there's no need for authentication, anonymous join attempt will generate the cookie."

		fmt.Printf("%s %s\n", color.Red("zpk"), color.Yellow(info))
		fmt.Println()

		os.Exit(1)
	}

	fmt.Printf("build is %s\n", color.Yellow(version))

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

func availableToken() string {
	switch {
	case os.Getenv("TOKEN") != "":
		return os.Getenv("TOKEN")
	case baked != "":
		return baked
	default:
		return ""
	}
}

func randId() int {
	min := 50000000000 // Just to avoid a sea of non existent ids
	max := 99999999999

	return rand.Intn(max-min+1) + min
}

func hide(msg string) string {
	return censor(msg, 3)
}

func censor(msg string, n int) string {
	text := []rune(msg)
	x := strings.Repeat("#", len(msg)-n)

	return string(text[0:n]) + x
}

func find(id int) {
	m, err := tangalanga.FindMeeting(id)

	if err != nil && *debugFlag {
		fmt.Printf("%s\n", err)
	}

	if err == nil {
		r := m.GetRoom()

		roomId := strconv.FormatUint(r.GetRoomId(), 10)
		roomName := r.GetRoomName()
		user := r.GetUser()
		link := r.GetLink()

		msg := "\nRoom ID: %s.\n" +
			"Room: %s.\n" +
			"Owner: %s.\n" +
			"Link: %s\n\n"

		if !*censorFlag || *outputFile != "" {
			log.WithFields(log.Fields{
				"room_id":   roomId,
				"room_name": roomName,
				"owner":     user,
				"link":      link,
			}).Info(r.GetPhoneNumbers())
		}

		if *censorFlag {
			roomId = hide(roomId)
			roomName = hide(roomName)
			user = hide(user)
			link = censor(link, 10)
		}

		fmt.Printf(msg,
			color.Green(roomId),
			color.Green(roomName),
			color.Green(user),
			color.Yellow(link),
		)
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

	fmt.Printf("searching for open meetings... %s\n", color.Yellow("please wait"))
	fmt.Println()

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

		if tangalanga.ExpiredToken {
			m := "the %s expired, you need to sniff join traffic for a new one\n"
			fmt.Printf(m, color.Red("token"))

			m = "you need to pass %s with the new token\n"
			fmt.Printf(m, color.Red("-token="))
			c <- syscall.SIGINT
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
