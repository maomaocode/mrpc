package codec

import "io"

// 请求需要 服务名 方法名 参数
// 响应需要 返回值 错误
// 参数 和 返回值 抽象为body
type Header struct {
	ServiceMethod string // 服务方法
	Seq           uint64 // 请求的id
	Error         string
}

type Codec interface {
	io.Closer
	ReadHeader(*Header) error
	ReadBody(interface{}) error
	Write(*Header, interface{}) error
}

type NewCodecFunc func(closer io.ReadWriteCloser) Codec

type Type string

const (
	GobType  Type = "application/gob"
	JsonType Type = "application/json" //
)

var NewCodecFuncMap map[Type]NewCodecFunc

func init() {
	NewCodecFuncMap = make(map[Type]NewCodecFunc)
	NewCodecFuncMap[GobType] = NewGobCodec
}