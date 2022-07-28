package parse

import (
	"goload/conf"
	"testing"
)

func TestParserStruct(t *testing.T) {
	c := conf.LocalConf{L: conf.LocalConf{}}
	p := NewParser(
		SetIdent(JSON),
		SetValidTag(ValidTag),
		SetDefaultTag(DefaultTag),
		SetDescTag(DescTag),
	)
	if err := p.InspectStruct(&c); err != nil {
		t.Fatal(err)
	}
	t.Log("success")
}
