package parse

import (
	"io/fs"
	"io/ioutil"
)

func (p *parser) Load(content []byte) error {
	if err := p.decoder(content, p.source); err != nil {
		return err
	}
	return p.InspectStruct(p.source)
}

func (p *parser) ImportFile(fileName string) error {
	content, err := ioutil.ReadFile(fileName)
	if err != nil {
		return err
	}
	return p.decoder(content, p.source)
}

func (p *parser) ExportFile(filePath string) error {
	content, err := p.encoder(p.source)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filePath, content, fs.FileMode(0600))
}

func (p *parser) LoadEnv() error {
	// TODO implement me
	panic("implement me")
}

func (p *parser) LoadCmd() error {
	// TODO implement me
	panic("implement me")
}
