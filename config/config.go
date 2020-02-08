package config

// TODO Default config for EVERYTHING
// TODO Validate cron expr
// TODO Extra Validation

type Config struct {
	Kibana           Kibana           `validate:"required"`
	Logging          Logging          `validate:"required"`
	AutoIndexPattern AutoIndexPattern `validate:"required"`
}

type Kibana struct {
	Host     string `validate:"required,uri"`
	User     string `validate:"required_with=password"`
	Password string `validate:"required_with=User"`
}

type GeneralPattern struct {
	Pattern       string
	TimeFieldName string
}

type AutoIndexPattern struct {
	Enabled         bool
	GeneralPatterns []GeneralPattern
	Schedule        string
}

type Logging struct {
	Level  string `validate:"required,oneof=debug info warn fatal panic"`
	Format string `validate:"required,oneof=console json logfmt"`
	Debug  bool
	Color  bool
}
