package deps

import (
	"github.com/semanggilab/webcore-go/app/core"
	modulea "github.com/semanggilab/webcore-go/modules/modulea"
	tb "github.com/semanggilab/webcore-go/modules/tb"
)

var APP_PACKAGES = []core.Module{
	modulea.NewModule(),
	tb.NewModule(),

	// Add your packages here
}
