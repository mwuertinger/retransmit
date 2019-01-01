package main

import (
	"log"
	"os"

	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/mwuertinger/retransmit/recv"
	"github.com/mwuertinger/retransmit/send"
)

func main() {
	app := kingpin.New("retransmit", "")
	timeout := app.Flag("timeout", "timeout of send and receive operations in Go notation (eg. 10s)").Short('t').Default("10s").Duration()
	cmdSend := app.Command("send", "run in sending mode")
	frameSize := cmdSend.Flag("frame-size", "size of a frame in MiB").Short('F').Default("1").Uint()
	destination := cmdSend.Arg("destination", "host:port").Required().String()
	cmdRecv := app.Command("recv", "run in receiving mode")
	listen := cmdRecv.Arg("listen", "host:port").Required().String()

	cmd, err := app.Parse(os.Args[1:])
	if err != nil {
		log.Fatal(err)
	}

	switch cmd {
	case "send":
		send.Send(*destination, *timeout, *frameSize*1024*1024)
	case "recv":
		recv.Recv(*listen, *timeout)
	default:
		panic("undefined command: " + cmd)
	}
}
