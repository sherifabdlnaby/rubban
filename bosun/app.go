package bosun

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/sherifabdlnaby/bosun/bosun/kibana"
	config "github.com/sherifabdlnaby/bosun/config"
	"github.com/sherifabdlnaby/bosun/log"
)

type Bosun struct {
	config           config.Config
	logger           log.Logger
	client           *kibana.Client
	semVer           semver.Version
	api              kibana.API
	autoIndexPattern AutoIndexPattern
}

func Main() {
	bosun := Bosun{}
	err := bosun.Initialize()
	if err != nil {
		panic("Failed to Initalize Bosun. Error: " + err.Error())
	}

	bosun.AutoIndexPattern()
}

func (b *Bosun) Initialize() error {

	var err error

	// Get Default logger
	logger := log.Default()

	// Load config
	b.config, err = config.Load("bosun")
	if err != nil {
		logger.Fatal("Failed to load configuration.", "error", err)
		os.Exit(1)
	}

	// Init logger
	b.logger = log.NewZapLoggerImpl("bosun", b.config.Logging)
	b.logger.Info("Starting Bosun...")

	// Init Kibana API Client
	b.logger.Info("Initializing Kibana API Client...")
	b.client, err = kibana.NewKibanaClient(b.config.Kibana, b.logger.Extend("client"))
	if err != nil {
		b.logger.Fatalw("Could not Initialize Kibana API Client", "error", err.Error())
	}

	// Validate Connection
	if !b.client.Validate(5, 10*time.Second) {
		err = fmt.Errorf("couldn't validate connection to Kibana API")
		b.logger.Fatal("Cannot Initialize Bosun without an Initial Connection to Kibana API")
		return err
	}
	b.logger.Info("Validated Initial Connection to Kibana API")

	// Get Kibana Version (To Determine which set of APIs to use later)
	b.semVer, err = b.client.GuessVersion()
	if err != nil {
		err = fmt.Errorf("couldn't determine kibana version")
		b.logger.Fatal(strings.ToTitle(err.Error()))
		return err
	}
	b.logger.Infow(fmt.Sprintf("Determined Kibana Version: %s", b.semVer.String()))

	// Determine API
	// TODO for now bosun only support API V7
	b.api = kibana.NewApiVer7(b.client)

	// Init AutoIndexPattern
	b.autoIndexPattern = *NewAutoIndexPattern(b.config.AutoIndexPattern)

	return nil
}
