package config

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
	Pattern       string `validate:"required"`
	TimeFieldName string
}

type AutoIndexPattern struct {
	Enabled         bool
	GeneralPatterns []GeneralPattern `validate:"required,min=1"`
	Schedule        string           `validate:"required"`
}

type Logging struct {
	Level  string `validate:"required,oneof=debug info warn fatal panic"`
	Format string `validate:"required,oneof=console json logfmt"`
	Debug  bool
	Color  bool
}
