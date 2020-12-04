package server

import "github.com/mrpc/codec"


const MagicNumber = 0x3bef5c

var DefaultOption = &Option{
	MagicNumber: MagicNumber,
	CodecType:   codec.GobType,
}

type Option struct {
	MagicNumber int //
	CodecType   codec.Type
}
