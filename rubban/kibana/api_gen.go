package kibana

import (
	"context"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/sherifabdlnaby/rubban/config"
	"github.com/sherifabdlnaby/rubban/log"
)

//APIVer7 Implements API Calls compatible with Kibana 7^
type APIGen struct {
	*client
	log log.Logger
}

//NewAPIVer7 Constructor
func NewAPIGen(config config.Kibana, log log.Logger) (*APIGen, error) {
	client, err := NewKibanaClient(config, log.Extend("client"))
	if err != nil {
		return &APIGen{}, err
	}

	return &APIGen{
		client: client,
		log:    log,
	}, nil
}

func (a *APIGen) Validate(ctx context.Context) error {
	return a.validate(ctx, 5, 10*time.Second)
}

func (a *APIGen) GuessVersion(ctx context.Context) (semver.Version, error) {
	return a.guessVersion(ctx)
}

func (a *APIGen) Info(ctx context.Context) (Info, error) {
	panic("Should Not Be Called from Gen Pattern.")
}

func (a *APIGen) Indices(ctx context.Context, filter string) ([]Index, error) {
	panic("Should Not Be Called from Gen Pattern.")
}

func (a *APIGen) IndexPatterns(ctx context.Context, filter string) ([]IndexPattern, error) {
	panic("Should Not Be Called from Gen Pattern.")
}

func (a *APIGen) BulkCreateIndexPattern(ctx context.Context, indexPattern map[string]IndexPattern) error {
	panic("Should Not Be Called from Gen Pattern.")
}
