package main

import (
	"embed"
	"fmt"
	"log"
	"os"

	"github.com/wailsapp/wails/v3/pkg/application"

	"taie/internal/app"
	"taie/internal/data"
	"taie/internal/pkg/logger"
)

// Wails uses Go's `embed` package to embed the frontend files into the binary.
// Any files in the frontend/dist folder will be embedded into the binary and
// made available to the frontend. The embed lives in the main package because
// //go:embed can only reference files under this directory; internal/app
// receives it as a parameter.
// See https://pkg.go.dev/embed for more information.

var assets embed.FS

func init() {
	// Register a custom event whose associated data type is string.
	// This is not required, but the binding generator will pick up registered events
	// and provide a strongly typed JS/TS API for them.
	application.RegisterEvent[string]("time")
}

// main is the application's entry point. All dependencies are assembled by the
// Wire injector (see wire.go / wire_gen.go); main only runs the result.
func main() {
	// `go run . migrate [dsn]` 子命令：执行数据库初始化（migrations/ 下的 SQL）。
	// 一般通过 `make migrate` 调用。
	if len(os.Args) > 1 && os.Args[1] == "migrate" {
		dsn := ""
		if len(os.Args) > 2 {
			dsn = os.Args[2]
		}
		if err := data.Migrate(dsn); err != nil {
			log.Fatal(err)
		}
		fmt.Println("数据库初始化完成")
		return
	}

	// Get() 初始化全局单例并 slog.SetDefault：之后任意包可 logger.Info(...) / slog.Info(...)。
	// 同时作为依赖传入 Wire——wire.go 里声明 *slog.Logger 参数的构造函数按需拿到它。
	wailsApp, cleanup, err := InitializeApp(logger.Get())
	if err != nil {
		log.Fatal(err)
	}
	defer cleanup()

	// Emit a "time" event every second while the app is running.
	app.StartClockEmitter(wailsApp)

	// Run the application. This blocks until the application has been exited.
	if err := wailsApp.Run(); err != nil {
		log.Fatal(err)
	}
}
