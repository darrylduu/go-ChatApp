package main

import (
	"flag"
	"fmt"
	"github.com/manishmeganathan/peerchat/src"
	"github.com/sirupsen/logrus"
	"os"
	"time"
)

const APPFLAG = "TT4_Tuesday"

func init() {
	logrus.SetFormatter(&logrus.TextFormatter{
		ForceColors:     true,
		FullTimestamp:   true,
		TimestampFormat: time.RFC822,
	})
	logrus.SetOutput(os.Stdout)
}

func main() {
	username := flag.String("User", "", "Username to use in the chatRoom")
	chatRoom := flag.String("Room", "", "ChatRoom to Join")
	logLevel := flag.String("log", "", "Level of logs to print")
	discovery := flag.String("discover", "", "Method to use for discovery")
	flag.Parse()

	switch *logLevel {
	case "panic", "PANIC":
		logrus.SetLevel(logrus.PanicLevel)
	case "fatal", "FATAL":
		logrus.SetLevel(logrus.FatalLevel)
	case "error", "ERROR":
		logrus.SetLevel(logrus.ErrorLevel)
	case "warn", "WARN":
		logrus.SetLevel(logrus.WarnLevel)
	case "info", "INFO":
		logrus.SetLevel(logrus.InfoLevel)
	case "debug", "DEBUG":
		logrus.SetLevel(logrus.DebugLevel)
	case "trace", "TRACE":
		logrus.SetLevel(logrus.TraceLevel)
	default:
		logrus.SetLevel(logrus.InfoLevel)
	}
	fmt.Println(APPFLAG)
	fmt.Println("The Chat Application is starting...")
	fmt.Println("Please be waiting...This might take 3*10 seconds...Raf is Working on it...")
	fmt.Println()

	p2phost := src.NewP2P()
	logrus.Infoln("Completed P2P Setup...Raf is doing a great job...")

	switch *discovery {
	case "announce":
		p2phost.AnnounceConnect()
	case "advertise":
		p2phost.AdvertiseConnect()
	default:
		p2phost.AdvertiseConnect()
	}
	logrus.Infoln("Connected to Service...")
	chatapp, _ := src.JoinChatRoom(p2phost, *username, *chatRoom)
	logrus.Infoln("Joined the '%s' chatroom as '%s'", chatapp.RoomName, chatapp.UserName)
	time.Sleep(time.Second * 5)
	ui := src.NewUI(chatapp)
	ui.Run()
}
