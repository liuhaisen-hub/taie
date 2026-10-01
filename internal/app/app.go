package app

import (
	"embed"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"

	"taie/internal/pkg/events"
	"taie/internal/services"
)

// NewWailsApp is the Wire provider for the Wails application. It receives the
// embedded frontend assets (the frontend/dist tree lives at the repo root, so
// it is embedded in the main package and passed in here) and every service
// that should be bound to the frontend as parameters, so adding a new service
// means adding a provider and listing it in wire.go.
func NewWailsApp(assets embed.FS, modelService *services.ModelServices, toolsService *services.ToolsServices, systemService *services.SystemServices, sessionService *services.SessionServices, tokenUseService *services.TokenUseServices, ags *services.AgentServices) *application.App {
	// Create a new Wails application by providing the necessary options.
	// Variables 'Name' and 'Description' are for application metadata.
	// 'Assets' configures the asset server with 'assets' pointing to the frontend files.
	// 'Services' is a list of Go struct instances. The frontend has access to the methods of these instances.
	// 'Mac' options tailor the application when running on macOS.
	wailsApp := application.New(application.Options{
		Name:        "taie",
		Description: "A demo of using raw HTML & CSS",
		Services: []application.Service{
			application.NewService(modelService),
			application.NewService(toolsService),
			application.NewService(systemService),
			application.NewService(sessionService),
			application.NewService(tokenUseService),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: true,
		},
	})
	// 注册 AI 问答流式通道。
	// 协议：前端连接后发送一条 {"user_input": "..."}，Go 持续回推 ChatEvent
	// 直到回答完成（done/error）后关闭连接；前端提前 close 会取消推理。
	wailsApp.HandleStream("agent/chat", func(c *application.StreamConn) {
		defer c.Close()

		// 读取用户输入；连接关闭/窗口刷新会在这里返回错误。
		var req services.ChatRequest
		if err := c.ReceiveJSON(&req); err != nil {
			return
		}

		// c.Context() 在前端断开时自动取消，Ask 内部的模型调用随之中断。
		// Ask 出错时已通过 onEvent 推送 error 事件，这里无需重复处理。
		ags.ChatWithAgent(c.Context(), &req, func(e events.Event) error {
			return c.SendJSON(e)
		})
	})
	// 注册审批回调通道，协议与 agent/chat 对称：
	// 前端收到 approval 事件弹卡片，用户裁决后发一条 {"session_id","interrupt_id","approved","reason"}，
	// Go 从 checkpoint 续跑并回推同一套事件，直到 done/error/approval(多工具连续审批)。
	wailsApp.HandleStream("agent/approval", func(c *application.StreamConn) {
		defer c.Close()

		var req services.ApprovalRequest
		if err := c.ReceiveJSON(&req); err != nil {
			return
		}

		ags.ApprovalWithAgent(c.Context(), &req, func(e events.Event) error {
			return c.SendJSON(e)
		})
	})
	// Create a new window with the necessary options.
	// 'Title' is the title of the window.
	// 'Mac' options tailor the window when running on macOS.
	// 'BackgroundColour' is the background colour of the window.
	// 'URL' is the URL that will be loaded into the webview.
	wailsApp.Window.NewWithOptions(application.WebviewWindowOptions{
		Title: "Window 1",
		// Window sized to the golden ratio (1200 / 742 ≈ 1.618).
		Width:  1200,
		Height: 742,
		// Minimum size is 80% of the default, keeping the same ratio
		// (960 / 594 ≈ 1.618).
		MinWidth:  960,
		MinHeight: 594,
		Mac: application.MacWindow{
			InvisibleTitleBarHeight: 50,
			Backdrop:                application.MacBackdropTranslucent,
			TitleBar:                application.MacTitleBarHiddenInset,
		},
		BackgroundColour: application.NewRGB(6, 7, 15),
		URL:              "/",
	})

	return wailsApp
}

// StartClockEmitter starts a goroutine that emits an event containing the
// current time every second. The frontend can listen to this event and update
// the UI accordingly.
func StartClockEmitter(wailsApp *application.App) {
	go func() {
		for {
			now := time.Now().Format(time.RFC1123)
			wailsApp.Event.Emit("time", now)
			time.Sleep(time.Second)
		}
	}()
}
