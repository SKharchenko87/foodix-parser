package main

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/SKharchenko87/foodix-parser/internal/config"
	"github.com/SKharchenko87/foodix-parser/internal/parser"
	"github.com/SKharchenko87/foodix-parser/internal/storage"
)

func main() {
	bootstrapLogger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// Подгружаем путь до конфига
	ConfigPath, exists := os.LookupEnv("CONFIG_PATH")
	if !exists {
		bootstrapLogger.Error("CONFIG_PATH is not exists")
		os.Exit(1)
	}

	// Конфиг
	cfg, err := config.LoadConfig(ConfigPath)
	if err != nil {
		bootstrapLogger.Error("failed to load config", "path", ConfigPath, "error", err)
		os.Exit(1)
	}

	// Logger
	logger := initLogger(cfg)
	slog.SetDefault(logger)
	logger.Info("Starting app")

	// Список источников
	sources, err := initSources(cfg.Sources)
	if err != nil {
		logger.Error("failed to initialize sources", "error", err)
		os.Exit(1)
	}

	// Хранилище данных
	store, err := storage.NewStore(cfg.Store)
	if err != nil {
		logger.Error("failed to initialize store", "error", err)
		os.Exit(1)
	}

	// Парсим и записываем
	err = run(sources, store)
	if err != nil {
		logger.Error("failed to run", "error", err)
		os.Exit(1)
	}
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

func initSources(cfgSources []config.SourceConfig) ([]parser.Parser, error) {
	result := make([]parser.Parser, 0, len(cfgSources))
	for _, source := range cfgSources {
		pars, err := parser.NewParser(source)
		if err != nil {
			return nil, err
		}
		result = append(result, pars)
	}
	return result, nil
}

func run(sources []parser.Parser, store storage.DB) error {
	for _, source := range sources {
		data, err := source.Parse()
		if err != nil {
			return fmt.Errorf("failed parsing source %s: %w", source.GetName(), err)
		}

		err = store.InsertProducts(data)
		if err != nil {
			return fmt.Errorf("failed inserting products: %w", err)
		}
	}
	return nil
}
