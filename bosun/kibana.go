package bosun

import (
	"github.com/sherifabdlnaby/bosun/bosun/kibana"
	cfg "github.com/sherifabdlnaby/bosun/config"
)

type Kibana struct {
	config cfg.Kibana
	client kibana.Client
}
