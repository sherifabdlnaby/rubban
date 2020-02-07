package config

// TODO Default Config for EVERYTHING

type Config struct {
	Kibana  Kibana  `validate:"required"`
	Logging Logging `validate:"required"`
}

type Kibana struct {
	Host     string `validate:"required,uri"`
	User     string `validate:"required_with=password"`
	Password string `validate:"required_with=User"`
}

type IndexPatternDiscover struct {
	IndicesPatterns []string
}

type Logging struct {
	Level  string `validate:"required,oneof=debug info warn fatal panic"`
	Format string `validate:"required,oneof=console json logfmt"`
	Debug  bool
	Color  bool
}
