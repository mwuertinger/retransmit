package main

import (
	"gopkg.in/alecthomas/kingpin.v2"
	"log"
	"os"
)

func main() {
	app := kingpin.New("retransmit", "")
	cmdSend := app.Command("send", "")
	destination := cmdSend.Arg("destination", "host:port").String()
	cmdRecv := app.Command("recv", "")
	listen := cmdRecv.Arg("listen", "host:port").String()

	cmd, err := app.Parse(os.Args)
	if err != nil {
		log.Fatal(err)
	}

	switch cmd {
	case "send":
		send(*destination)
	case "recv":
		recv(*listen)
	default:
		panic("undefined command: " + cmd)
	}
}

func send(destination string) {

}
