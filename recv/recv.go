package recv

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"sync"
	"time"

	"github.com/mwuertinger/retransmit/common"
)

const (
	frameBufLen = 16
)

type sequence struct {
	mu  sync.Mutex
	num uint64
}

func Recv(listen string, timeout time.Duration) error {
	frameChan := make(chan *common.Frame, frameBufLen)

	listener, err := net.Listen("tcp", listen)
	if err != nil {
		return fmt.Errorf("listen: %v", err)
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		for frame := range frameChan {
			_, err := os.Stdout.Write(frame.Data)
			if err != nil {
				log.Fatal("write to stdout: ", err)
			}
		}
		wg.Done()
	}()

	var sequence sequence

	for {
		conn, err := listener.Accept()
		if err != nil {
			return fmt.Errorf("accept: %v", err)
		}
		err = handleConn(conn, frameChan, timeout, &sequence)
		if err == io.EOF {
			listener.Close()
			break
		}
		if err != nil {
			log.Print("handleConn: ", err)
		}
		time.Sleep(1 * time.Second)
	}

	// Wait for go routine to write everything to stdout
	wg.Wait()

	return nil
}

func handleConn(conn net.Conn, frameChan chan<- *common.Frame, timeout time.Duration, sequence *sequence) error {
	defer conn.Close()
	log.Print("New incoming connection from ", conn.RemoteAddr())
	for {
		err := handleFrame(conn, frameChan, timeout, sequence)
		if err == io.EOF {
			return io.EOF
		}
		if err != nil {
			return fmt.Errorf("handleFrame: %v", err)
		}
	}
}

func handleFrame(conn net.Conn, frameChan chan<- *common.Frame, timeout time.Duration, sequence *sequence) error {
	if err := conn.SetDeadline(time.Now().Add(timeout)); err != nil {
		return fmt.Errorf("SetDeadline: %v", err)
	}

	frame, err := common.UnmarshalFrame(conn)
	if err != nil {
		return fmt.Errorf("UnmarshalFrame: %v", err)
	}

	log.Print("Got frame ", frame.Sequence)

	sequence.mu.Lock()
	defer sequence.mu.Unlock()

	if sequence.num < frame.Sequence {
		return fmt.Errorf("gap in sequence numbers: expected=%d, got=%d", sequence.num, frame.Sequence)
	}

	err = common.MarshalFrame(conn, &common.Frame{Sequence: frame.Sequence})
	if err != nil {
		return fmt.Errorf("sending ack failed: %v", err)
	}

	if sequence.num > frame.Sequence {
		log.Printf("Discarding duplicate frame %d, expected %d", frame.Sequence, sequence.num)
		return nil
	}

	sequence.num++
	frameChan <- frame

	if frame.Length == 0 {
		log.Printf("Received final frame, closing channel")
		close(frameChan)
		return io.EOF
	}

	return nil
}
