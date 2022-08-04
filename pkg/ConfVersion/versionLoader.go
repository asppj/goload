package ConfVersion

import (
	"fmt"
)

type ConfValid interface {
	Valid() (bool, error)
}

type confLoader[T ConfValid] struct {
	version      int       // 当前版本
	mod          int       // 保留最近几个版本
	versionCache map[int]T // 版本
}

type ConfLoader interface {
}

func NewConfLoader[T ConfValid](mod int) ConfLoader {
	return &confLoader[T]{
		version:      0,
		mod:          mod,
		versionCache: make(map[int]T),
	}
}

func (c *confLoader[T]) Load(newConf T) error {
	valid, err := newConf.Valid()
	if err != nil {
		return err
	}
	if !valid {
		return fmt.Errorf("valid err:%v", err)
	}
	if len(c.versionCache) != 0 {
		c.version++
	}
	c.versionCache[c.version] = newConf
	if c.mod > 1 && len(c.versionCache) > c.mod { // 0,1,2; 2
		delete(c.versionCache, c.version-c.mod) // 删除老版本
	}
	return nil
}

func (c *confLoader[T]) Get() T {
	return c.versionCache[c.version]
}
