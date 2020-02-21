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
	return a.validate(ctx, 5 , 10 * time.Second)
}


func (a *APIGen) GuessVersion(ctx context.Context)  (semver.Version, error)  {
	return a.guessVersion(ctx)
}

func (a *APIGen) Info() (Info, error) {
	panic("Should Not Be Called from Gen Pattern.")
}

func (a *APIGen) Indices(filter string) ([]Index, error) {
	panic("Should Not Be Called from Gen Pattern.")
}

func (a *APIGen) IndexPatterns(filter string) ([]IndexPattern, error) {
	panic("Should Not Be Called from Gen Pattern.")
}

func (a *APIGen) BulkCreateIndexPattern(indexPattern []IndexPattern) error {
	panic("Should Not Be Called from Gen Pattern.")
}
