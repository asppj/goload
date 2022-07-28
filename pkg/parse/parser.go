package parse

import (
	"io"
	"log"
	"os"
)

type parser struct {
	logger          *log.Logger
	tagOpt          *TagOption
	source          interface{}
	fields          []*parseField
	allFields       []*parseField
	anonymousFields []*parseField // 匿名嵌套结构体
}

type Parser interface {
	InspectStruct(interface{}) error
	Load(closer io.ReadCloser) error
	LoadEnv() error // 环境变量加载
	LoadCmd() error // os.cmd加载

}

func newDefaultParse() *parser {
	return &parser{
		tagOpt: NewDefaultTagOpt(),
		logger: log.New(os.Stdout, "", log.Llongfile),
	}
}

func NewParser(opts ...SetOpt) Parser {
	p := newDefaultParse()
	// apply options
	for _, opt := range opts {
		opt(p)
	}
	return p
}
