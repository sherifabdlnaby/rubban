package bosun

import (
	"fmt"

	config "github.com/sherifabdlnaby/bosun/config"
	"github.com/sherifabdlnaby/bosun/log"
)

type App struct {
	Config config.Config
	Logger log.Logger
}

func Run() {
	app := Initialize()
}

func Initialize() App {
	// Get Default Logger
	logger := log.Default()

	// Load Config
	Config, err := config.Load("bosun")
	if err != nil {
		logger.Infow("failed to load configuration.", "error", err)
	}

	// Init Logger
	logger = log.NewZapLoggerImpl("bosun", Config.Logging)
	logger.Info("Hello, World!")

	// App Struct to hold common resources
	return App{Config: *Config, Logger: logger}
}
