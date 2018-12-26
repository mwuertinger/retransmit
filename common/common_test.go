package common

import (
	"bytes"
	"testing"
)

func Test(t *testing.T) {
	frame0 := Frame{
		Sequence: 17,
		Data:     []byte{17, 23, 42},
	}

	var buf bytes.Buffer
	err := MarshalFrame(&buf, &frame0)
	if err != nil {
		t.Error(err)
	}

	frame1, err := UnmarshalFrame(&buf)
	if err != nil {
		t.Error(err)
	}

	if frame0.Sequence != frame1.Sequence {
		t.Error("sequence differs")
	}
	if bytes.Compare(frame0.Data, frame1.Data) != 0 {
		t.Error("data differs")
	}
}
