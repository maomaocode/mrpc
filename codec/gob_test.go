package codec

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
	"testing"
)

type P struct {
	X, Y, Z int
	Name    string
}

type Q struct {
	X, Y *int32
	Name string
}

func TestGob(t *testing.T) {
	var network bytes.Buffer

	enc := gob.NewEncoder(&network)
	dec := gob.NewDecoder(&network)

	_ = enc.Encode(P{3, 4, 5, "mmm"})

	_ = enc.Encode(P{1, 1, 1, "ggg"})


	var q Q
	_ = dec.Decode(&q)
	fmt.Printf("%q: {%d, %d}\n", q.Name, *q.X, *q.Y)

	dec.Decode(&q)
	fmt.Printf("%q: {%d, %d}\n", q.Name, *q.X, *q.Y)
}


type Vector struct {
	x, y, z int
}

func (v Vector) MarshalBinary() ([]byte, error) {
	// A simple encoding: plain text.
	var b bytes.Buffer
	_, _ = fmt.Fprintln(&b, v.x, v.y, v.z)
	return b.Bytes(), nil
}

// UnmarshalBinary modifies the receiver so it must take a pointer receiver.
func (v *Vector) UnmarshalBinary(data []byte) error {
	// A simple encoding: plain text.
	b := bytes.NewBuffer(data)
	_, err := fmt.Fscanln(b, &v.x, &v.y, &v.z)
	return err
}

func TestGob2(t *testing.T) {


	// The Vector type has unexported fields, which the package cannot access.
	// We therefore write a BinaryMarshal/BinaryUnmarshal method pair to allow us
	// to send and receive the type with the gob package. These interfaces are
	// defined in the "encoding" package.
	// We could equivalently use the locally defined GobEncode/GobDecoder
	// interfaces.

	var network bytes.Buffer // Stand-in for the network.

	// Create an encoder and send a value.
	enc := gob.NewEncoder(&network)
	err := enc.Encode(&Vector{3, 4, 5})
	if err != nil {
		log.Fatal("encode:", err)
	}

	// Create a decoder and receive a value.
	dec := gob.NewDecoder(&network)
	var v Vector
	err = dec.Decode(&v)
	if err != nil {
		log.Fatal("decode:", err)
	}
	fmt.Println(v)



}