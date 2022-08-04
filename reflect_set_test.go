package goload

import (
	"testing"

	"github.com/asppj/goload/conf"
)

func TestLoader(t *testing.T) {
	c := conf.LocalConf{L: []conf.Logger{
		{Name: "testLoader", Output: []string{"1"}},
	}}
	if err := LoadStruct(&c, "default"); err != nil {
		t.Fatal(err)
	}
	t.Log("success")
}
