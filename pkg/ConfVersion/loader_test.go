package ConfVersion

import (
	"testing"

	"github.com/asppj/goload/conf"
	"github.com/asppj/goload/pkg/parse"
)

func TestLoader(t *testing.T) {
	c := conf.LocalConf{L: []conf.Logger{
		{Name: "testLoader", Output: []string{"1"}},
	}}
	if err := parse.LoadStruct(&c, parse.NewDefaultTagOpt()); err != nil {
		t.Fatal(err)
	}
	t.Log("success")
}
