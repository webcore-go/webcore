package basic

import (
	"github.com/webcore-go/webcore/adapter/auth/authn"
	"github.com/webcore-go/webcore/port"
)

type BasicAuthLoader struct {
	name string
}

func (a *BasicAuthLoader) SetName(name string) {
	a.name = name
}

func (a *BasicAuthLoader) Name() string {
	return a.name
}

func (a *BasicAuthLoader) Init(args ...any) (port.Library, error) {

	authn := &authn.AuthN{}
	authn.SetValidator(&BasicAuthValidator{})
	err := authn.Install(args...)
	if err != nil {
		return nil, err
	}

	return authn, nil
}
