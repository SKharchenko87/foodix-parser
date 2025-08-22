package main

import (
	"flag"
	"log"
	"log/slog"
	"os"
	"strings"

	"github.com/SKharchenko87/foodix-parser/internal/config"
	"github.com/SKharchenko87/foodix-parser/internal/parser"
	"github.com/SKharchenko87/foodix-parser/internal/storage"
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

	// Logger
	logger := initLogger(cfg)
	slog.SetDefault(logger)
	logger.Info("Starting app")

	// Список источников
	sources, err := initSources(cfg.Sources)
	if err != nil {
		bootstrapLogger.Error("failed to initialize sources", "error", err)
	}

	// Хранилище данных
	store, err := storage.NewStore(cfg.Store)
	if err != nil {
		bootstrapLogger.Error("failed to initialize store", "error", err)
	}
	run(sources, store)
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

func run(sources []parser.Parser, store storage.DB) {
	for _, source := range sources {
		data, err := source.Parse()
		if err != nil {
			log.Fatal(err)
			return
		}

		err = store.InsertProducts(data)
		if err != nil {
			log.Fatal(err)
			return
		}
	}
}
