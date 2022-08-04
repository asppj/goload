package parse

import (
	"io/ioutil"
	"testing"

	"github.com/asppj/goload/conf"
)

const (
	cfgFilePathDev  = "../../conf/config.dev.yaml"
	cfgFilePathTest = "../../conf/config.test.yaml"
	cfgFilePathTmpl = "../../conf/config.template.yaml"
)

func TestParserStruct(t *testing.T) {
	// a, b := "l1", "l2"
	// c := conf.LocalConf{L: []conf.Logger{
	// 	{
	// 		Name:   &a,
	// 		Output: nil,
	// 	},
	// 	{
	// 		Name:   &b,
	// 		Output: nil,
	// 	},
	// }}
	c := conf.LocalConf{}
	p := NewParser(
		SetIdent(JSON),
		SetValidTag(ValidTag),
		SetDefaultTag(DefaultTag),
		SetDescTag(DescTag),
	)
	if err := p.InspectStruct(&c); err != nil {
		t.Fatal(err)
	}
	if err := p.ExportFile(cfgFilePathTmpl); err != nil {
		t.Fatal(err)
	}
	content, err := ioutil.ReadFile(cfgFilePathDev)
	if err != nil {
		t.Fatal(err)
	}
	if err = p.Load(content); err != nil {
		t.Fatal(err)
	}
	if err = p.ExportFile(cfgFilePathTest); err != nil {
		t.Fatal(err)
	}
	t.Log("success")
}
