package parse

import (
	"goload/conf"
	"io/ioutil"
	"testing"
)

const (
	cfgFilePathDev  = "../../conf/config.dev.yaml"
	cfgFilePathTest = "../../conf/config.test.yaml"
	cfgFilePathTmpl = "../../conf/config.template.yaml"
)

func TestParserStruct(t *testing.T) {
	c := conf.LocalConf{L: []conf.Logger{
		{
			Name:   "l1",
			Output: nil,
		},
		{
			Name:   "l2",
			Output: nil,
		},
	}}
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
