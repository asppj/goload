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

func TestRedis(t *testing.T) {
	c := conf.Redis{DB: 4}
	if err := LoadStruct(&c, "default"); err != nil {
		t.Fatal(err)
	}
	t.Logf("%+v", c)
	t.Log("success")
}
