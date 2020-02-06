package bosun

import (
	"time"

	"github.com/sherifabdlnaby/bosun/bosun/kibana"
	config "github.com/sherifabdlnaby/bosun/config"
	"github.com/sherifabdlnaby/bosun/log"
)

type App struct {
	Config config.Config
	Logger log.Logger
}

func Run() {
	app := Initialize()

	client, err := kibana.NewKibanaClient(app.Config.Kibana, app.Logger.Extend("client"))
	if err != nil {
		panic(err)
	}

	if !client.Validate(5, 10*time.Second) {
		app.Logger.Panicf("couldn't validate connection to Kibana")
	}

	ver, err := client.GuessVersion()
	if err != nil {
		panic(err)
	}
	app.Logger.Infof("Determined Kibana Version: %s", ver.String())

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
