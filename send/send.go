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
	frameChan := make(chan *common.Frame, 16)

	// Go routine reads from stdin and sends frames of data to channel.
	go func() {
		var seq uint64
		eof := false

		// r := bufio.NewReaderSize(os.Stdin, 2*int(frameSize))
		r := os.Stdin

		for {
			buf := make([]byte, frameSize)
			offset := 0
			for n := 0; offset < int(frameSize); offset += n {
				var err error
				n, err = r.Read(buf)
				if err == io.EOF {
					eof = true
					break
				}
				if err != nil {
					// if an error occurs reading from stdin all we can do is give up
					log.Fatal(err)
				}
			}

			if !eof && offset != int(frameSize) {
				log.Printf("non-full frame: %d", offset)
			}

			// send the frame to the channel
			frameChan <- &common.Frame{
				Sequence: seq,
				Data:     buf[0:offset],
			}

			seq++

			if eof {
				// if we hit EOF and the last frame was not empty we have to send a final empty frame
				if offset > 0 {
					frameChan <- &common.Frame{
						Sequence: seq,
					}
				}

				return
			}
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
		log.Printf("Sent frame %d (%d bytes)", (*curFrame).Sequence, (*curFrame).Length)

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
