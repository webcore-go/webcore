package yaml

import (
	"github.com/webcore-go/webcore/adapter/authstore/store"
	"github.com/webcore-go/webcore/infra/config"
	"github.com/webcore-go/webcore/port"
)

type YamlLoader struct {
	name string
}

func (a *YamlLoader) SetName(name string) {
	a.name = name
}

func (a *YamlLoader) Name() string {
	return a.name
}

func (l *YamlLoader) Init(args ...any) (port.Library, error) {
	config := args[1].(config.AuthConfig)
	backend, err := YamlBackend(config.Control, config.Directory)
	if err != nil {
		return nil, err
	}

	store := &store.AuthStore{}
	store.SetBackend(backend)
	err = store.Install(args...)
	if err != nil {
		return nil, err
	}

	return store, nil
}
