package cache

import (
	"github.com/webcore-go/webcore/adapter/authsession/session"
	"github.com/webcore-go/webcore/app/core"
	"github.com/webcore-go/webcore/infra/config"
	"github.com/webcore-go/webcore/port"
)

type MemoryCacheSessionLoader struct {
	name string
}

func (a *MemoryCacheSessionLoader) SetName(name string) {
	a.name = name
}

func (a *MemoryCacheSessionLoader) Name() string {
	return a.name
}

func (l *MemoryCacheSessionLoader) Init(args ...any) (port.Library, error) {
	context := args[0].(*core.AppContext)
	config := args[1].(config.AuthConfig)

	backend, err := MemoryCacheSessionBackend(config, context)
	if err != nil {
		return nil, err
	}

	session := &session.AuthSession{}
	session.SetSessionStore(backend)
	err = session.Install(args...)
	if err != nil {
		return nil, err
	}

	return session, nil
}
