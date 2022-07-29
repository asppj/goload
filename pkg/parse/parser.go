package parse

import (
	"bytes"
	"encoding/json"
	"log"
	"os"

	"github.com/pelletier/go-toml/v2"
	"gopkg.in/yaml.v3"
)

type parser struct {
	logger          *log.Logger
	tagOpt          *TagOption
	source          interface{}
	fields          []*parseField
	allFields       []*parseField
	anonymousFields []*parseField // 匿名嵌套结构体
	encoder         MarshalFunc
	decoder         UnmarshalFunc
}

type Parser interface {
	InspectStruct(interface{}) error
	Load(readCloser []byte) error     // load from reader
	ImportFile(filePath string) error // import cfg from file
	ExportFile(filePath string) error // export cfg to file
	LoadEnv() error                   // load from env
	LoadCmd() error                   // load from os.args

}

type UnmarshalFunc func(in []byte, out interface{}) error
type MarshalFunc func(any) ([]byte, error)
type EncoderFunc func(any) error

// encoder
var (
	JSONEncoder = func(a any) ([]byte, error) {
		buf := bytes.Buffer{}
		enc := json.NewEncoder(&buf)
		enc.SetIndent("", "  ")
		err := enc.Encode(a)
		if err != nil {
			return nil, err
		}
		return buf.Bytes(), nil
	}
	YAMLEncoder = func(a any) ([]byte, error) {
		buf := bytes.Buffer{}
		enc := yaml.NewEncoder(&buf)
		enc.SetIndent(2)
		err := enc.Encode(a)
		if err != nil {
			return nil, err
		}
		return buf.Bytes(), nil
	}
	TOMLEncoder = func(a any) ([]byte, error) {
		buf := bytes.Buffer{}
		enc := toml.NewEncoder(&buf)
		err := enc.Encode(a)
		if err != nil {
			return nil, err
		}
		return buf.Bytes(), nil
	}
)

// decoder
var (
	JSONDecoder = json.Unmarshal
	YAMLDecoder = yaml.Unmarshal
	TOMLDecoder = toml.Unmarshal
)

func newDefaultParse() *parser {
	return &parser{
		tagOpt:  NewDefaultTagOpt(),
		logger:  log.New(os.Stdout, "", log.Llongfile),
		encoder: JSONEncoder,
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
