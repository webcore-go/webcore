package deps

import (
	"github.com/semanggilab/webcore-go/app/core"
	// modulea "github.com/semanggilab/webcore-go/modules/modulea"
	tb "github.com/semanggilab/webcore-go/modules/tb"
	"github.com/semanggilab/webcore-go/modules/tbpubsub"
)

var APP_PACKAGES = []core.Module{
	// modulea.NewModule(),
	tb.NewModule(),
	tbpubsub.NewModule(),

	// Add your packages here
}
