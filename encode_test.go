package teafile_test

import (
	"bytes"
	"log"
	"testing"
	"time"

	"github.com/blainsmith/teafiles"
)

type myEncoderType struct {
	Timestamp time.Time `teafiles:"ts,timestamp"`
	Volume    int8      `teafiles:"volume"`
}

func TestEncode(t *testing.T) {
	var h teafiles.Header
	met := myEncoderType{
		Volume: 10,
	}

	buf := bytes.NewBuffer(nil)

	enc, err := teafiles.NewEncoder(buf, &h)
	if err != nil {
		t.Error(err)
	}

	if err := enc.Encode(met); err != nil {
		t.Error(err)
	}

	met.Timestamp = time.Now()
	if err := enc.Encode(met); err != nil {
		t.Error(err)
	}

	log.Fatalf("%x", buf.String())
}
