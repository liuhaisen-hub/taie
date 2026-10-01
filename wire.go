//go:build wireinject
// +build wireinject

package main

import (
	"log/slog"

	"taie/internal/agentkit"
	"taie/internal/app"
	"taie/internal/data"
	"taie/internal/services"

	"github.com/google/wire"
	"github.com/wailsapp/wails/v3/pkg/application"
)

// InitializeApp is the Wire injector. It is compiled only when running the
// `wire` tool (build tag `wireinject`); the generated implementation lives in
// wire_gen.go. Regenerate with `make wire` (or the `wire` CLI) after changing
// providers.
//
// *slog.Logger 是注入源（同 sale 的 wireApp 模式）：main 把全局 logger 传进来后，
// wire.Build 里任何构造函数声明 *slog.Logger 参数即可按需拿到它，
// 不需要的构造函数不用改签名。
func InitializeApp(log *slog.Logger) (*application.App, func(), error) {
	wire.Build(
		wire.Value(assets),
		wire.Value(data.DefaultDSN),
		agentkit.ProviderSet,
		services.ProviderSet,
		data.ProviderSet,
		app.NewWailsApp,
	)
	return nil, nil, nil
}
