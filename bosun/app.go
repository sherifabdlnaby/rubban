package bosun

import (
	"fmt"
	"os"
	"time"

	"github.com/sherifabdlnaby/bosun/bosun/kibana"
	config "github.com/sherifabdlnaby/bosun/config"
	"github.com/sherifabdlnaby/bosun/log"
)

type App struct {
	Config config.Config
	Log    log.Logger
}

func Run() {
	app := Initialize()
	app.Log.Info("Starting Bosun...")

	client, err := kibana.NewKibanaClient(app.Config.Kibana, app.Log.Extend("client"))
	if err != nil {
		panic(err)
	}

	if !client.Validate(5, 10*time.Second) {
		app.Log.Fatal("Couldn't Validate Connection to Kibana API")
		return
	}
	app.Log.Info("Validated Initial Connection to Kibana API")

	semVer, err := client.GuessVersion()
	if err != nil {
		app.Log.Errorw("Couldn't Determine Kibana Version.", "error", err.Error())
		return
	}

	app.Log.Infow(fmt.Sprintf("Determined Kibana Version: %s", semVer.String()))

}

func Initialize() App {
	// Get Default Log
	logger := log.Default()

	// Load Config
	Config, err := config.Load("bosun")
	if err != nil {
		logger.Fatal("Failed to load configuration.", "error", err)
		os.Exit(1)
	}

	// Init Log
	logger = log.NewZapLoggerImpl("bosun", Config.Logging)

	// App Struct to hold common resources
	return App{Config: *Config, Log: logger}
}
