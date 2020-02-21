package rubban

import (
	"context"
	"fmt"
	"os"

	"github.com/Masterminds/semver/v3"
	"github.com/sherifabdlnaby/rubban/config"
	"github.com/sherifabdlnaby/rubban/log"
	"github.com/sherifabdlnaby/rubban/rubban/autoindexpattern"
	"github.com/sherifabdlnaby/rubban/rubban/kibana"
)

//Rubban App Structure
type Rubban struct {
	config           *config.Config
	logger           log.Logger
	semVer           semver.Version
	api              kibana.API
	scheduler        scheduler
	autoIndexPattern autoindexpattern.AutoIndexPattern
	mainCtx          context.Context
	cancel           context.CancelFunc
}

//New Create new App structure
func New() *Rubban {
	rootCtx, cancel := context.WithCancel(context.Background())
	return &Rubban{mainCtx: rootCtx, cancel: cancel, logger: log.Default()}
}

var errFailedToInitialize = fmt.Errorf("failed to Initialize application")

//Initialize Initialize Application after Loading Configuration
func (r *Rubban) Initialize() error {

	var err error

	// Load config
	r.config, err = config.Load("Rubban")
	if err != nil {
		r.logger.Fatalw("Failed to load configuration.", "error", err)
		os.Exit(1)
	}

	// Init logger
	r.logger = log.NewZapLoggerImpl("Rubban", r.config.Logging)

	r.logger.Info("Successfully Loaded Configuration")

	// Create scheduler
	r.scheduler = *newScheduler(r.mainCtx, r.logger.Extend("scheduler"))

	// Init Kibana API client
	err = r.initKibanaClient(r.mainCtx)
	if err != nil {
		r.logger.Fatalw("Failed to initialize Kibana API", "error", err)
	}

	// Init Tasks
	r.initTasks()

	// Register Tasks
	err = r.registerTasks()
	if err != nil {
		r.logger.Fatalw("Failed to initialize scheduler", "error", err)
	}

	return nil
}

//Start Start Rubban
func (r *Rubban) Start() {

	r.logger.Infof("Starting Rubban...")

	// Start scheduler
	r.scheduler.Start()
}

//Stop Rubban (Will wait for everything to finish)
func (r *Rubban) Stop() {
	r.logger.Infof("Rubban is Stopping...")

	// Cancel Main Context
	r.cancel()

	// Stop and Wait for all running jobs to finish
	r.scheduler.Stop()

	r.logger.Infof("Stopped.")
	r.logger.Infof("Goodbye <3")
	_ = r.logger.Sync()
}

func (r *Rubban) initTasks() {

	if r.config.AutoIndexPattern.Enabled {
		r.autoIndexPattern = *autoindexpattern.NewAutoIndexPattern(r.config.AutoIndexPattern, r.api, r.logger.Extend("autoIndexPattern"))
		r.logger.Infof("Enabled %s, Loaded %d General Pattern(s)", r.autoIndexPattern.Name(), len(r.autoIndexPattern.GeneralPatterns))
	}

	// ... Init Other Tasks in future
}

func (r *Rubban) registerTasks() error {

	// Register Auto Index Pattern
	if r.config.AutoIndexPattern.Enabled {
		err := r.scheduler.Register(r.config.AutoIndexPattern.Schedule, &r.autoIndexPattern)
		if err != nil {
			return fmt.Errorf("failed to register task, error: %s", err.Error())
		}
	}

	// ... Register Other Tasks in future
	return nil
}

func (r *Rubban) initKibanaClient(ctx context.Context) error {
	r.logger.Info("Initializing Kibana API client...")
	genAPI, err := kibana.NewAPIGen(r.config.Kibana, r.logger.Extend("api"))
	if err != nil {
		r.logger.Fatalw("Could not Initialize Kibana API client", "error", err.Error())
		return errFailedToInitialize
	}

	// Validate Connection to General API (Not versioned yet as we don't have version)
	if err = genAPI.Validate(ctx); err != nil {
		r.logger.Fatalw("Cannot Initialize Rubban without an Initial Connection to Kibana API", "error", err.Error())
		return errFailedToInitialize
	}
	r.logger.Info("Validated Initial Connection to Kibana API")

	// Get Kibana Version (To Determine which set of APIs to use later)
	r.semVer, err = genAPI.GuessVersion(ctx)
	if err != nil {
		r.logger.Fatalw("Couldn't determine kibana version", "error", err.Error())
		return errFailedToInitialize
	}
	r.logger.Infow(fmt.Sprintf("Determined Kibana Version: %s", r.semVer.String()))

	// Determine API
	// TODO for now Rubban only support API V7, when testing other Kibana
	r.api, err = kibana.NewAPIVer7(r.config.Kibana, r.logger)
	if err != nil {
		r.logger.Fatalw("Could not Initialize Kibana API client", "error", err.Error())
		return errFailedToInitialize
	}

	return nil
}
