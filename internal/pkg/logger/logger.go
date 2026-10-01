package logger

import (
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"gopkg.in/natefinch/lumberjack.v2"
)

func getLogFilePath() string {
	// 日志写到可执行文件所在目录的 logs/ 下
	if exePath, err := os.Executable(); err == nil {
		// go run / go test 时可执行文件在临时构建目录里，此时退回到工作目录下的 logs/
		if !strings.HasPrefix(exePath, os.TempDir()) {
			return filepath.Join(filepath.Dir(exePath), "logs", "app.log")
		}
	}
	return filepath.Join("logs", "app.log")
}

func NewLogger() *slog.Logger {

	lumberWriter := &lumberjack.Logger{
		Filename:   getLogFilePath(),
		MaxSize:    10,
		MaxBackups: 5,
		MaxAge:     10,
		Compress:   true,
		LocalTime:  true,
	}
	handler := slog.NewTextHandler(lumberWriter, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	return slog.New(handler)
}

// 全局单例：任意包 import 后通过 Get() / 便捷函数使用，首次调用时初始化。
var (
	once   sync.Once
	global *slog.Logger
)

// Get 返回全局 logger，并在首次调用时将其设为 slog 的默认 logger
// （之后 slog.Info 等原生函数也会写入同一个日志文件）。
func Get() *slog.Logger {
	once.Do(func() {
		global = NewLogger()
		slog.SetDefault(global)
	})
	return global
}

// 便捷函数：业务代码直接 logger.Info("msg", "key", val) 即可，无需持有实例。
func Debug(msg string, args ...any) { Get().Debug(msg, args...) }
func Info(msg string, args ...any)  { Get().Info(msg, args...) }
func Warn(msg string, args ...any)  { Get().Warn(msg, args...) }
func Error(msg string, args ...any) { Get().Error(msg, args...) }
