package teafile_test

import (
	"os"
	"testing"
	"time"

	"github.com/blainsmith/teafiles"
)

type myDecoderType struct {
	Timestamp time.Time `teafiles:"ts,timestamp"`
	Volume    int8      `teafiles:"volume"`
}

func TestDecode(t *testing.T) {
	var h teafiles.Header
	f, _ := os.Open("./AA.tea")

	t.Run("nil rows", func(t *testing.T) {
		err := teafiles.Decode(f, &h, nil)
		if err != teafiles.ErrNonPointer {
			t.Error(err)
		}
	})

	t.Run("non-pointer rows", func(t *testing.T) {
		var mdt myDecoderType
		err := teafiles.Decode(f, &h, mdt)
		if err != teafiles.ErrNonPointer {
			t.Error(err)
		}
	})

	t.Run("pointer to unsupported type rows", func(t *testing.T) {
		var mdt myDecoderType
		err := teafiles.Decode(f, &h, &mdt)
		if v, ok := err.(*teafiles.UnsupportedTypeError); !ok {
			t.Error(v)
		}
	})
}

func TestDecoder(t *testing.T) {
	var h teafiles.Header
	f, _ := os.Open("./AA.tea")

	dec, err := teafiles.NewDecoder(f, &h)
	if err != nil {
		t.Error(err)
	}

	var det myDecoderType
	if err := dec.Decode(&det); err != nil {
		t.Error(err)
	}
}
