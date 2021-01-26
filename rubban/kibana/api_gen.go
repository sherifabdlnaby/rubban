package kibana

import (
	"context"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/sherifabdlnaby/rubban/config"
	"github.com/sherifabdlnaby/rubban/log"
)

//APIGen Implements API Calls compatible with Kibana 7^
type APIGen struct {
	client *ClientVer7
	log    log.Logger
}

//NewAPIGen Constructor
func NewAPIGen(config config.Kibana, log log.Logger) (*APIGen, error) {
	client, err := NewKibanaClientVer7(config, log.Extend("ClientVer7"))
	if err != nil {
		return &APIGen{}, err
	}

	return &APIGen{
		client: client,
		log:    log,
	}, nil
}

//Validate Validate connection to Kibana
func (a *APIGen) Validate(ctx context.Context) error {
	return a.client.Validate(ctx, 5, 10*time.Second)
}

//GuessVersion Try to Guess Current Kibana API version
func (a *APIGen) GuessVersion(ctx context.Context) (semver.Version, error) {
	return a.client.GuessVersion(ctx)
}

//Info Return Kibana Info
func (a *APIGen) Info(ctx context.Context) (Info, error) {
	panic("Should Not Be Called from Gen Pattern.")
}

//Indices Get Indices match supported filter (support wildcards)
func (a *APIGen) Indices(ctx context.Context, filter string) ([]Index, error) {
	panic("Should Not Be Called from Gen Pattern.")
}

//IndexPatterns Get IndexPatterns from kibana matching the supplied filter (support wildcards)
func (a *APIGen) IndexPatterns(ctx context.Context, filter string, fields []string) ([]IndexPattern, error) {
	panic("Should Not Be Called from Gen Pattern.")
}

//BulkCreateIndexPattern Add Index Patterns to Kibana
func (a *APIGen) BulkCreateIndexPattern(ctx context.Context, indexPatterns []IndexPattern) error {
	panic("Should Not Be Called from Gen Pattern.")
}
