package ConfVersion

import (
	"sync"

	"github.com/spf13/viper"
)

type loader struct {
	vp     map[string]*viper.Viper
	locker sync.RWMutex
}

func NewLoader() *loader {
	return &loader{vp: make(map[string]*viper.Viper), locker: sync.RWMutex{}}
}

func (l *loader) Load(userValue string, configURL string) error {
	return nil
}
