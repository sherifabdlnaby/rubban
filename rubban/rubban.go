package rubban

import (
	"context"
	"fmt"
	"os"

	"github.com/Masterminds/semver/v3"
	"github.com/sherifabdlnaby/rubban/config"
	"github.com/sherifabdlnaby/rubban/log"
	"github.com/sherifabdlnaby/rubban/rubban/autoIndexPattern"
	"github.com/sherifabdlnaby/rubban/rubban/kibana"
)

type rubban struct {
	config           *config.Config
	logger           log.Logger
	semVer           semver.Version
	api              kibana.API
	scheduler        Scheduler
	autoIndexPattern autoIndexPattern.AutoIndexPattern
	mainCtx          context.Context
	cancel           context.CancelFunc
}

func New() *rubban {
	rootCtx, cancel := context.WithCancel(context.Background())
	return &rubban{mainCtx:rootCtx, cancel:cancel, logger:  log.Default()}
}

var ErrFailedToInitialize = fmt.Errorf("failed to Initialize application")

func (R *rubban) Initialize() error {

	var err error

	// Load config
	R.config, err = config.Load("rubban")
	if err != nil {
		R.logger.Fatalw("Failed to load configuration.", "error", err)
		os.Exit(1)
	}

	// Init logger
	R.logger = log.NewZapLoggerImpl("rubban", R.config.Logging)

	R.logger.Info("Successfully Loaded Configuration")

	// Create Scheduler
	R.scheduler = *NewScheduler(R.mainCtx, R.logger.Extend("scheduler"))

	// Init Kibana API client
	err = R.initKibanaClient(R.mainCtx)
	if err != nil {
		R.logger.Fatalw("Failed to initialize Kibana API", "error", err)
	}

	// Init Tasks
	err = R.initTasks()
	if err != nil {
		R.logger.Fatalw("Failed to initialize tasks", "error", err)
	}

	// Register Tasks
	err = R.registerTasks()
	if err != nil {
		R.logger.Fatalw("Failed to initialize scheduler", "error", err)
	}

	return nil
}

func (R *rubban) Start() {

	R.logger.Infof("Starting Rubban...")

	// Start scheduler
	R.scheduler.Start()
}

func (R *rubban) Stop() {
	R.logger.Infof("Rubban is Stopping...")

	// Cancel Main Context
	R.cancel()

	// Stop and Wait for all running jobs to finish
	R.scheduler.Stop()

	R.logger.Infof("Stopped.")
	R.logger.Infof("Goodbye <3")
	_ = R.logger.Sync()
}

func (r *rubban) initTasks() error {

	if r.config.AutoIndexPattern.Enabled{
		r.autoIndexPattern = *autoIndexPattern.NewAutoIndexPattern(r.config.AutoIndexPattern, r.api, r.logger.Extend("autoIndexPattern"))
		r.logger.Infof("Enabled %s, Loaded %d General Pattern(s)", r.autoIndexPattern.Name(), len(r.autoIndexPattern.GeneralPatterns))
	}


	// ... Init Other Tasks in future
	return nil
}

func (R *rubban) registerTasks() error {

	// Register Auto Index Pattern
	if R.config.AutoIndexPattern.Enabled {
		err := R.scheduler.Register(R.config.AutoIndexPattern.Schedule, &R.autoIndexPattern)
		if err != nil {
			return fmt.Errorf("failed to register task, error: %s", err.Error())
		}
	}

	// ... Register Other Tasks in future
	return nil
}

func (R *rubban) initKibanaClient(ctx context.Context) error {
	R.logger.Info("Initializing Kibana API client...")
	genAPI, err := kibana.NewAPIGen(R.config.Kibana, R.logger.Extend("api"))
	if err != nil {
		R.logger.Fatalw("Could not Initialize Kibana API client", "error", err.Error())
		return ErrFailedToInitialize
	}

	// Validate Connection to General API (Not versioned yet as we don't have version)
	if err = genAPI.Validate(ctx); err != nil {
		R.logger.Fatalw("Cannot Initialize Rubban without an Initial Connection to Kibana API", "error", err.Error())
		return ErrFailedToInitialize
	}
	R.logger.Info("Validated Initial Connection to Kibana API")

	// Get Kibana Version (To Determine which set of APIs to use later)
	R.semVer, err = genAPI.GuessVersion(ctx)
	if err != nil {
		R.logger.Fatalw("Couldn't determine kibana version", "error", err.Error())
		return ErrFailedToInitialize
	}
	R.logger.Infow(fmt.Sprintf("Determined Kibana Version: %s", R.semVer.String()))

	// Determine API
	// TODO for now rubban only support API V7, when testing other Kibana
	R.api, err = kibana.NewAPIVer7(R.config.Kibana, R.logger)
	return nil
}

