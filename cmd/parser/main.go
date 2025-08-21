package main

import (
	"flag"
	"log"
	"log/slog"
	"os"
	"strings"

	"github.com/SKharchenko87/foodix-parser/internal/config"
	"github.com/SKharchenko87/foodix-parser/internal/models"
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

	run(cfg)
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

func run(cfg config.Config) {
	var data []models.Product
	for _, source := range cfg.Sources {
		if source.Name == "calorizator" {
			pars := parser.NewCalorizator(source)
			var err error
			data, err = pars.Parse()
			if err != nil {
				log.Fatal(err)
				return
			}
		}

		var store storage.DB
		var err error
		if cfg.Store.Name == "postgres" {
			store, err = storage.NewPostgres(cfg.Store)
			if err != nil {
				log.Fatal(err)
				return
			}
		}

		err = store.InsertProducts(data)
		if err != nil {
			log.Fatal(err)
			return
		}
	}
}
