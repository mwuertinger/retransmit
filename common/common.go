package common

import (
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

const (
	magic        = 0xaabb8b216c5120f4
	headerLength = 3*8 + 32
)

type Frame struct {
	Magic    uint64   // magic constant
	Length   uint64   // length of Data in bytes
	Sequence uint64   // sequence number
	Checksum [32]byte // SHA256 checksum over entire frame
	Data     []byte   // payload data
}

func MarshalFrame(w io.Writer, frame *Frame) error {
	frame.Magic = magic
	frame.Length = uint64(len(frame.Data))

	buf := make([]byte, headerLength+frame.Length)
	binary.BigEndian.PutUint64(buf[0:8], frame.Magic)
	binary.BigEndian.PutUint64(buf[8:16], frame.Length)
	binary.BigEndian.PutUint64(buf[16:24], frame.Sequence)
	// skip checksum
	if copy(buf[headerLength:], frame.Data) != int(frame.Length) {
		panic("copy failed")
	}

	checksum := sha256.Sum256(buf)
	if copy(buf[24:56], checksum[:]) != 32 {
		panic("copy failed")
	}

	n, err := w.Write(buf)
	if err != nil {
		return err
	}
	if n != len(buf) {
		return fmt.Errorf("not entire buffer was sent")
	}

	return err
}

func UnmarshalFrame(r io.Reader) (*Frame, error) {
	var err error
	var buf [headerLength]byte
	for n, offset := 0, 0; uint64(offset) < headerLength; offset += n {
		n, err = r.Read(buf[offset:])
		if err != nil {
			return nil, fmt.Errorf("read: %v", err)
		}
	}

	var frame Frame
	frame.Magic = binary.BigEndian.Uint64(buf[0:8])
	frame.Length = binary.BigEndian.Uint64(buf[8:16])
	frame.Sequence = binary.BigEndian.Uint64(buf[16:24])
	if copied := copy(frame.Checksum[:], buf[24:56]); copied != 32 {
		panic("expected 32 bytes to copy")
	}

	if frame.Magic != magic {
		return nil, fmt.Errorf("invalid magic number: %v", frame.Magic)
	}

	frame.Data = make([]byte, frame.Length)
	for n, offset := 0, 0; uint64(offset) < frame.Length; offset += n {
		n, err = r.Read(frame.Data[offset:])
		if err != nil {
			return nil, fmt.Errorf("read data: %v", err)
		}
		//log.Printf("read %d bytes", n)
	}

	frameBuf := make([]byte, headerLength+frame.Length)
	if copy(frameBuf, buf[0:3*8]) != 3*8 { // do not copy checksum
		panic("failed to copy header")
	}
	if copy(frameBuf[headerLength:], frame.Data) != int(frame.Length) {
		panic("failed to copy data")
	}

	checksum := sha256.Sum256(frameBuf)
	if checksum != frame.Checksum {
		return nil, errors.New("checksum error")
	}

	return &frame, nil
}
