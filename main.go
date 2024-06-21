package main

import (
	_ "embed"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gempir/go-twitch-irc/v4"
)

const (
	loadListingsInterval = time.Minute * 10
	loadItemsInterval    = time.Minute * 60
)

func main() {
	_ = captains.loadFromFile()
	_ = listings.loadFromFile()
	_ = items.loadFromFile()

	go loadCaptainsAsync()
	go loadMLBCardsAsync()
	go loadListingsasync()

	log.Println(os.Getenv("TWITCH_CHANNELS"))

	channels := strings.Split(os.Getenv("TWITCH_CHANNELS"), ",")

	clientID := os.Getenv("CLIENT_ID")
	log.Println("Get an access token\n",
		"https://id.twitch.tv/oauth2/authorize?client_id="+
			clientID+
			"&redirect_uri=http://localhost&response_type=token&scope=chat:read%20chat:edit")

	client := twitch.NewClient("joshq00", os.Getenv("TWITCH_OAUTH_TOKEN"))

	client.OnConnect(func() {
		slog.Info("connected")
	})

	client.OnPrivateMessage(pmhandler(client))

	client.OnWhisperMessage(func(msg twitch.WhisperMessage) {})
	client.OnClearChatMessage(func(msg twitch.ClearChatMessage) {})
	client.OnClearMessage(func(msg twitch.ClearMessage) {
		log.Printf("[DELETED] #%s @%s : %s\n", msg.Channel, msg.Login, msg.Message)
	})
	client.OnRoomStateMessage(func(msg twitch.RoomStateMessage) {})
	client.OnUserNoticeMessage(func(msg twitch.UserNoticeMessage) {})
	client.OnUserStateMessage(func(msg twitch.UserStateMessage) {})
	client.OnGlobalUserStateMessage(func(msg twitch.GlobalUserStateMessage) {
		gusm = msg
	})
	client.OnNoticeMessage(func(msg twitch.NoticeMessage) {})
	client.OnUserJoinMessage(func(msg twitch.UserJoinMessage) {})
	client.OnUserPartMessage(func(msg twitch.UserPartMessage) {})

	client.Join(channels...)

	go func() {
		err := client.Connect()
		if err != nil {
			panic(err)
		}
	}()
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	for _, v := range channels {
		client.Depart(v)
	}
	log.Println(client.Disconnect())
	time.Sleep(time.Second)
}
