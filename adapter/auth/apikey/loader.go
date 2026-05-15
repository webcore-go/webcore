package apikey

import (
	"github.com/webcore-go/webcore/adapter/auth/authn"
	"github.com/webcore-go/webcore/infra/config"
	"github.com/webcore-go/webcore/port"
)

type ApiKeyLoader struct {
	name string
}

func (a *ApiKeyLoader) SetName(name string) {
	a.name = name
}

func (a *ApiKeyLoader) Name() string {
	return a.name
}

func (a *ApiKeyLoader) Init(args ...any) (port.Library, error) {
	config := args[1].(config.AuthConfig)
	authn := &authn.AuthN{}
	authn.SetValidator(NewApiKeyValidator(config))
	err := authn.Install(args...)
	if err != nil {
		return nil, err
	}

	return authn, nil
}
