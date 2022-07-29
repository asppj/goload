package parse

import (
	"reflect"
	"strings"
)

type SetOpt func(p *parser)

func SetIdent(ident string) SetOpt {
	return func(p *parser) {
		p.tagOpt.IdentTag = ident
		switch ident {
		case JSON:
			p.encoder = JSONEncoder
			p.decoder = JSONDecoder
		case YAML:
			p.encoder = YAMLEncoder
			p.decoder = YAMLDecoder
		case TOML:
			p.encoder = TOMLEncoder
			p.decoder = TOMLDecoder
		}
	}
}

func SetValidTag(tag string) SetOpt {
	return func(p *parser) {
		p.tagOpt.ValidTag = tag
	}
}

func SetDescTag(tag string) SetOpt {
	return func(p *parser) {
		p.tagOpt.DescTag = tag
	}
}

func SetDefaultTag(tag string) SetOpt {
	return func(p *parser) {
		p.tagOpt.DefaultTag = tag
	}
}

type (
	// TagOption 选项
	TagOption struct {
		IdentTag   string // yaml,json,toml,id等
		DefaultTag string // 默认值，会被更高优先级覆盖
		OptionTag  string // 选项,
		DescTag    string // 描述，html显示;
		// Option     string // 选项，只能选择其中某些值 html显示 Usage: oneof=red green \n oneof=5 7 9
		ValidTag string // 验证 github.com/go-playground/validator/v10
	}
	// TagValue 值
	TagValue struct {
		Ident      string `json:"ident"`
		Default    string `json:"default"`
		DefaultSet bool   // true: 设置了默认值
		Option     string `json:"option"`
		Describe   string `json:"describe"`
		Valid      string `json:"valid"`
	}
)

func NewDefaultTagOpt() *TagOption {
	return &TagOption{
		IdentTag:   ID,
		DefaultTag: DefaultTag,
		DescTag:    DescTag,
		ValidTag:   ValidTag,
	}
}

// parseField 对应字段
type parseField struct {
	value        reflect.Value
	defaultValue reflect.Value
	subFields    []*parseField // nested children
	fullIDParts  []string      // full ID of the option with all its parents
	isParent     bool          // is nested and has children
	isMap        bool          // is a map type
	isSlice      bool          // is a slice
	canSet       bool          //
	tagValue     TagValue      // struct.tag
}

// fullID returns the full ID of the option consisting of all IDs of its parents
// joined by dots.
func (o parseField) fullID() string {
	return strings.Join(o.fullIDParts, ".")
}
