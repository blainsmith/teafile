package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/blainsmith/teafile"
)

type myDecoderType struct {
	Timestamp time.Time `teafiles:"ts,timestamp"`
	Volume    int8      `teafiles:"volume"`
}

func main() {
	f, err := os.Open("./AA.tea")
	if err != nil {
		log.Fatal(err)
	}

	dec, err := teafile.NewDecoder(f)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%+v\n", dec.Header())

	var det myDecoderType
	if err := dec.Decode(&det); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%+v\n", det)
}
