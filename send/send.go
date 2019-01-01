package send

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"time"

	"github.com/mwuertinger/retransmit/common"
)

func Send(destination string, timeout time.Duration, frameSize uint) error {
	frameChan := make(chan *common.Frame)

	// Go routine reads from stdin and sends frames of data to channel.
	go func() {
		var seq uint64

		for {
			buf := make([]byte, frameSize)
			n, err := os.Stdin.Read(buf)
			if err == io.EOF {
				// End of transmission indicated by length 0 frame
				frameChan <- &common.Frame{
					Sequence: seq,
					Length:   0,
				}
				close(frameChan)
				return
			}
			// if an error occurs reading from stdin all we can do is give up
			if err != nil {
				log.Fatal(err)
			}

			// send the frame to the channel
			frameChan <- &common.Frame{
				Sequence: seq,
				Data:     buf[0:n],
			}

			seq++
		}
	}()

	var curFrame *common.Frame

	// Try to connect forever
	for {
		log.Print("Connecting to ", destination)
		conn, err := net.DialTimeout("tcp", destination, timeout)
		if err != nil {
			log.Print("Connecting failed: ", err)
			time.Sleep(1 * time.Second)
			continue
		}

		err = handleConn(conn, timeout, frameChan, &curFrame)
		if err == io.EOF {
			return nil
		}
		if err != nil {
			log.Print("handleConn: ", err)
		}
		time.Sleep(1 * time.Second)
	}

	return nil
}

func handleConn(conn net.Conn, timeout time.Duration, frameChan <-chan *common.Frame, curFrame **common.Frame) error {
	defer conn.Close()
	log.Print("Connection opened")

	for {
		if err := conn.SetDeadline(time.Now().Add(timeout)); err != nil {
			log.Printf("SetDeadline: %v", err)
			return nil
		}

		if *curFrame == nil {
			*curFrame = <-frameChan
		}

		if err := common.MarshalFrame(conn, *curFrame); err != nil {
			return err
		}
		log.Print("Sent frame ", (*curFrame).Sequence)

		ack, err := common.UnmarshalFrame(conn)
		if err != nil {
			return err
		}
		log.Print("Received ack ", ack.Sequence)

		if ack.Sequence != (*curFrame).Sequence {
			return fmt.Errorf("unexpected sequence number: expected=%v, got=%v", (*curFrame).Sequence, ack.Sequence)
		}

		// last frame has length 0
		if (*curFrame).Length == 0 {
			return io.EOF
		}

		*curFrame = nil
	}
}
