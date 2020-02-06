package config

type Config struct {
	Kibana Kibana `validate:"required"`
}

type Kibana struct {
	Host     string `validate:"required"`
	Port     int    `validate:"required"`
	User     string `validate:"required_with=Password"`
	Password string `validate:"required_with=User"`
}

type IndexPatternDiscover struct {
	IndicesPatterns []string
}

type Logging struct {
	Level  string `validate:"required, oneof=debug info warn fatal panic"`
	Format string `validate:"required, oneof=console json logfmt"`
	Debug  bool
	Color  bool
}
