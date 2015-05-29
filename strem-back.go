package main

import (
	"encoding/json"
	"fmt"
	"github.com/tealeg/xlsx"
	"github.com/thoj/go-ircevent"
	"golang.org/x/net/websocket"
	"net/http"
	"regexp"
	"strings"
	"time"
)

var playChannel = make(chan int)
var dongerChannel = make(chan int)
var moveChannel = make(chan int)
var musicChannel = make(chan int)

var cards map[string][]string
var donger string
var voiceMessage string
var move string
var track string

func wsCardsHandler(ws *websocket.Conn) {
	for {
		message := ""
		websocket.Message.Receive(ws, &message)
		switch message {
		case "getCards":
			{
				websocket.JSON.Send(ws, cards)
			}
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func wsChatHandler(ws *websocket.Conn) {
	for {
		/*
			message := ""
			websocket.Message.Receive(ws, &message)

			type messageType struct {
				id      string
				payload string
			}

			switch message {
		*/
		//case "getDonger":
		select {
		case <-dongerChannel:
			{
				response := map[string]string{
					"id":      "donger",
					"payload": donger,
				}
				websocket.JSON.Send(ws, response)
			}
		//case "getMove":
		case <-moveChannel:
			{
				response := map[string]string{
					"id":      "move",
					"payload": move,
				}
				websocket.JSON.Send(ws, response)
			}
		//case "getVoiceMessage":
		case <-playChannel:
			{
				response := map[string]string{
					"id":      "playMessage",
					"payload": voiceMessage,
				}
				websocket.JSON.Send(ws, response)
			}

		case <-musicChannel:
			{
				fmt.Printf("Sending play Track" + track)
				response := map[string]string{
					"id":      "playTrack",
					"payload": track,
				}
				websocket.JSON.Send(ws, response)
			}
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func getChat() {

	ircCodes := map[string]string{
		"WELCOME": "001",
		"NAMES":   "366",
	}

	users := map[string]int{}

	server := "irc.twitch.tv:6667"
	user := "knightofdongerino"
	oauth := "oauth:pfh7jb4mkvphr2pti2yb14o81me93m"
	channel := "#nl_kripp"
	//channel := "#knightofdongerino"

	ircConn := irc.IRC(user, user)
	//ircConn.VerboseCallbackHandler = true
	//ircConn.Debug = true
	ircConn.Password = oauth

	err := ircConn.Connect(server)

	if err != nil {
		fmt.Printf(err.Error())
		fmt.Printf("Can't connect to freenode.")
	} else {
		fmt.Printf("Connected")
	}

	ircConn.AddCallback(ircCodes["WELCOME"], func(e *irc.Event) { ircConn.Join(channel) })

	con2ok := false

	//message := `TO - Twktch Support SUBJECT - Cannot use PJSalt emoticon MESSAGE - Dear Twitch support, I can't seem to be able to emote with the PJSalt emoticon. Instead I receive the message "We are sorry, but user nl_Kripp exausted all the world salt reserves." Please advise.`
	message := ` <:::::::::::[=¤ԅ╏ ˵ ⊚ ◡ ⊚ ˵ ╏┐ I'm the knight of dongerino. Kind moderino ໒( • ͜ʖ • )७, don't mind me, I come in peace ᕕ( ՞ ᗜ ՞ )ᕗ`

	if false {
		ircConn.AddCallback(ircCodes["NAMES"], func(e *irc.Event) {
			t := time.NewTicker(31 * time.Second)
			i := 30
			for {
				<-t.C
				//ircConn.Privmsgf(channel, "Spamming every %d seconds\n", i)
				ircConn.Privmsgf(channel, message)
				if con2ok {
					i -= 1
				}
				if i == 0 {
					t.Stop()
					ircConn.Quit()
				}
			}
		})

	}
	regDonger := regexp.MustCompile("(.+)")
	regPlay := regexp.MustCompile("play (.+)")
	regMove := regexp.MustCompile("move (.+)")
	regMusic := regexp.MustCompile("music (.+)")

	ircConn.AddCallback("PRIVMSG", func(e *irc.Event) {
		if _, ok := users[e.User]; ok {
			users[e.User]++

		} else {
			users[e.User] = 1
		}

		arrayString := regDonger.FindStringSubmatch(e.Message())
		if arrayString != nil {
			//fmt.Printf("USER: %s N_MESSAGES: %s \n", e.User, arrayString)
			donger = fmt.Sprintf("%s: %s", e.User, arrayString[len(arrayString)-1])
			if len(donger) > 70 {
				donger = donger[0:70]
			}
			fmt.Printf("User: %s Donger : %s \n", e.User, donger)
			dongerChannel <- 1
		}

		arrayString = regMove.FindStringSubmatch(e.Message())
		if arrayString != nil {
			//fmt.Printf("USER: %s N_MESSAGES: %s \n", e.User, arrayString)
			move = fmt.Sprintf("%s: %s", e.User, arrayString[len(arrayString)-1])
			if len(move) > 70 {
				move = move[0:70]
			}
			fmt.Printf("User: %s Move: %s \n", e.User, move)
			moveChannel <- 1
		}

		arrayString = regPlay.FindStringSubmatch(e.Message())
		if arrayString != nil {
			voiceMessage = fmt.Sprintf("%s", arrayString[len(arrayString)-1])
			if len(voiceMessage) > 70 {
				voiceMessage = voiceMessage[0:70]
			}
			fmt.Printf("User: %s Message : %s \n", e.User, voiceMessage)
			playChannel <- 1
		}

		arrayString = regMusic.FindStringSubmatch(e.Message())
		if arrayString != nil {
			track = fmt.Sprintf("%s", arrayString[len(arrayString)-1])
			if len(track) > 70 {
				track = track[0:70]
			}
			fmt.Printf("User: %s Track: %s \n", e.User, track)
			musicChannel <- 1
		}
		//fmt.Printf("USER: %s N_MESSAGES: %s \n", e.User, arrayString)
		//donger = fmt.Sprintf("%s: %s", e.User, e.Message())

		//fmt.Printf("USER: %s N_MESSAGES: %d \n", e.User, users[e.User])
		//fmt.Printf("USER: %s MESSAGE: %s \n", e.User, e.Message())
	})

	ircConn.Loop()
}

func setCardWebSocket() {
	fmt.Printf("Start card  ws\n")
	http.Handle("/cards", websocket.Handler(wsCardsHandler))
	err := http.ListenAndServe(":12345", nil)
	if err != nil {
		panic("ListenAndServe: " + err.Error())
	}
	fmt.Printf("end card  ws\n")

}

func setChatWebSocket() {
	fmt.Printf("Start chat ws\n")
	http.Handle("/chat", websocket.Handler(wsChatHandler))
	err := http.ListenAndServe(":23456", nil)
	if err != nil {
		panic("ListenAndServe: " + err.Error())
	}
	fmt.Printf("end chat ws\n")
}

func testWebSocketServer() {
	origin := "http://localhost/"
	url := "ws://localhost:12345/cards"
	ws, err := websocket.Dial(url, "", origin)
	fmt.Printf("Start client\n")

	if err != nil {
		fmt.Printf(err.Error())
	}

	//for {
	//
	//}
	t := time.NewTicker(3 * time.Second)
	for {
		<-t.C
		if err := websocket.Message.Send(ws, []byte("getCards")); err != nil {
			fmt.Printf(err.Error())
		}
		var jsonCards []byte
		websocket.JSON.Receive(ws, &jsonCards)

		var cards map[string][]string
		json.Unmarshal(jsonCards, &cards)
		for key, value := range cards {
			fmt.Printf("%s > %s\n", key, value)
		}
	}

}

func parseExcelFile(fileName string) map[string][]string {

	cards := map[string][]string{}
	xlFile, _ := xlsx.OpenFile(fileName)

	sheet := xlFile.Sheet["Card List"]
	nameIndex := 3
	for _, row := range sheet.Rows {
		cardName := strings.ToLower(row.Cells[nameIndex].String())
		cards[cardName] = strings.Fields(cardName)
	}
	return cards
}

func main() {
	cards = parseExcelFile(`c:\Users\vince\workspace\go\src\github.com\vincenzoauteri\hello\hs.xlsx`)
	donger = "<:::::::::::[=¤ԅ╏ ˵ ⊚ ◡ ⊚ ˵ ╏┐"
	move = "power"
	go getChat()
	go setChatWebSocket()
	go setCardWebSocket()
	for {
		time.Sleep(1000 * time.Millisecond)
	}
}
