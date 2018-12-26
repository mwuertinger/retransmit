package recv

import (
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/mwuertinger/retransmit/common"
)

const (
	timeout = 10 * time.Second
)

var (
	mu                 sync.Mutex
	nextSequenceNumber uint64
)

func Recv(listen string) {
	frameChan := make(chan *common.Frame)
	for {
		recvInternal(listen, frameChan)
		time.Sleep(time.Second)
	}
}

func recvInternal(listen string, frameChan chan *common.Frame) error {
	listener, err := net.Listen("tcp", listen)
	if err != nil {
		return fmt.Errorf("listen: %v", err)
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			return fmt.Errorf("accept: %v", err)
		}

		go func() {
			defer conn.Close()
			if err := conn.SetDeadline(time.Now().Add(timeout)); err != nil {
				log.Printf("SetDeadline: %v", err)
				return
			}

			frame, err := common.UnmarshalFrame(conn)
			if err != nil {
				log.Printf("ParseFrame: %v", err)
				return
			}

			mu.Lock()
			defer mu.Unlock()
			if nextSequenceNumber != frame.Sequence {
				log.Printf("unexpected sequence number: %d", frame.Sequence)
				return
			}

			if err := common.SendAck(conn, frame); err != nil {
				log.Print("SendAck: ", err)
				return
			}

			nextSequenceNumber++

			frameChan <- frame
		}()
	}
}
