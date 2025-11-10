package logger

import (
	"log/slog"
	"os"

	"github.com/rodney-b/swish-test-publisher/internal/pkg/config"
)

var root *slog.Logger

func Initialize(cp config.ConfigProvider) {
	handlerOptions := slog.HandlerOptions{}

	if cp.IsDevelopment() {
		handlerOptions.Level = slog.LevelDebug
	}

	root = slog.New(slog.NewJSONHandler(os.Stdout, &handlerOptions))
}

func New(name string) *slog.Logger {
	return root.With("package", name)
}
