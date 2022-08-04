package parse_test

import (
	"fmt"
	"testing"

	"github.com/asppj/goload/pkg/parse"

	"gopkg.in/yaml.v3"
)

type (
	RedisConf struct {
		Host **string `yaml:"host" default:"localhost"`
		Port int      `yaml:"port" default:"6379"`
		DB   int      `yaml:"db" default:"5"`
	}
	Jwt struct {
		// Secret    string `yaml:"secret" default:"3.1415926"`
		*ArrayStr `json:"array_str" default:"127.0.0.1,localhost"`
	}
	Appserver[Other any] struct {
		// Versions []string  `yaml:"versions" default:"1,2,3" desc:"兼容的版本"`
		// model    string    `yaml:"model" default:"debug" option:"debug,test,prod"`
		// Redis    RedisConf `yaml:"redis"`
		// Other    Other     `yaml:"other"`
		// KV       map[string]ArrayStr `yaml:"kv" default:"white,black,grey"`
		Jwts map[string]Jwt `yaml:"jwts" default:"white,black,grey"`
		// Jwts2    map[string]*Jwt `yaml:"jwts2" default:"white,black,grey"`
		Jwts3 []Jwt `json:"jwts_3" default:"1,2,3"`
		// Jwts4    []****Jwt       `json:"jwts_4" default:"1,2,3"`
		// ArrayStr `json:"array_str" default:"127.0.0.1,localhost"`
	}
	ArrayStr []string
)

func TestLoadConf(t *testing.T) {
	c := &Appserver[Jwt]{Jwts: map[string]Jwt{"1234556": {}}}
	b, err := yaml.Marshal(&c)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%s\n", string(b))
	if err := parse.LoadStruct(c, parse.NewDefaultTagOpt()); err != nil {
		t.Fatal(err)
	}
	t.Logf("success\n")
}

func TestInspectStruct(t *testing.T) {
	c := &Appserver[Jwt]{Jwts: map[string]Jwt{"1": {}}}

	p := parse.NewParser(
		parse.SetIdent(parse.YAML),
		parse.SetValidTag(parse.ValidTag),
		parse.SetDefaultTag(parse.DefaultTag),
		parse.SetDescTag(parse.DescTag),
	)
	if err := p.InspectStruct(c); err != nil {
		t.Fatal(err)
	}
	t.Logf("success\n")
}
