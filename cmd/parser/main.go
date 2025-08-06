package main

import (
	"flag"
	"github.com/SKharchenko87/foodix-parser/internal/config"
	"log/slog"
	"os"
	"strings"
)

type Flags struct {
	ConfigPath string
}

func main() {
	// Аргументы запуска
	flags := initFlags()

	// Конфиг
	bootstrapLogger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	cfg, err := config.LoadConfig(flags.ConfigPath)
	if err != nil {
		bootstrapLogger.Error("failed to load config", "path", flags.ConfigPath, "error", err)
		os.Exit(1)
	}

	// Логер
	logger := initLogger(cfg)
	slog.SetDefault(logger)
	logger.Info("Starting app")

	//ToDo
	println(cfg.Source)
}

func initFlags() Flags {
	flags := Flags{}
	flag.StringVar(&flags.ConfigPath, "config", "configs/config.yaml", "Path to config file")
	flag.Parse()
	return flags
}

func initLogger(cfg config.Config) *slog.Logger {

	levelMap := map[string]slog.Level{
		"debug": slog.LevelDebug,
		"info":  slog.LevelInfo,
		"warn":  slog.LevelWarn,
		"error": slog.LevelError,
	}

	level, ok := levelMap[strings.ToLower(cfg.Log.Level)]
	if !ok {
		level = slog.LevelInfo
	}
	levelVar := new(slog.LevelVar)
	levelVar.Set(level)

	var handler slog.Handler
	opts := &slog.HandlerOptions{Level: levelVar}

	switch cfg.Log.Format {
	case "json":
		handler = slog.NewJSONHandler(os.Stdout, opts)
	default:
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	return slog.New(handler)
}
