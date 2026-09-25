//go:build wireinject
// +build wireinject

package main

import (
	"github.com/google/wire"
	"github.com/wailsapp/wails/v3/pkg/application"

	"taie/internal/app"
	"taie/internal/data"
	"taie/internal/services"
)

// InitializeApp is the Wire injector. It is compiled only when running the
// `wire` tool (build tag `wireinject`); the generated implementation lives in
// wire_gen.go. Regenerate with `make wire` (or the `wire` CLI) after changing
// providers.
func InitializeApp() (*application.App, error) {
	wire.Build(
		wire.Value(assets),
		wire.Value(data.DefaultDSN),
		services.ProviderSet,
		data.ProviderSet,
		app.NewWailsApp,
	)
	return nil, nil
}
