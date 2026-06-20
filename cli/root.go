// Package cli assembles the cmo command tree.
package cli

import (
	"github.com/tamnd/any-cli/kit"
	"github.com/tamnd/mathchina-cli/mathchina"
)

// Build metadata, set via -ldflags at release time.
var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)

// NewApp assembles the kit application. Commands are added as kit
// escape-hatch commands and call the mathchina client.
func NewApp() *kit.App {
	id := mathchina.Domain{}.Info().Identity
	id.Version = Version

	app := kit.New(id)
	(mathchina.Domain{}).Register(app)

	app.AddCommand(newListCmd())
	app.AddCommand(newProblemCmd())
	app.AddCommand(newSearchCmd())
	app.AddCommand(newForumsCmd())
	app.AddCommand(newVersionCmd())

	return app
}
