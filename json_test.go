package goload

import (
	"encoding/json"
	"fmt"
	"testing"
)

type LocalConf struct {
	LogMap **string `json:"logMap" desc:"日志" default:"default,app,server" option:"default"`
}

func TestDemoJson(t *testing.T) {
	logger := "defaultLoggerDem"
	l := &logger
	cases := LocalConf{LogMap: &l}
	bb, err := json.Marshal(cases)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%s\n", string(bb))
	data := LocalConf{}
	err = json.Unmarshal(bb, &data)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%+v\n", data)
}
